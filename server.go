package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//go:embed ui/*
var uiFS embed.FS

// setupRoutes registers all REST APIs and static web assets
func (a *App) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Static UI assets embedded into the binary
	sub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatalf("Failed to initialize embedded UI: %v", err)
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", fileServer)

	// API endpoints
	mux.HandleFunc("/api/feed", a.handleFeed)
	mux.HandleFunc("/api/search", a.handleSearch)
	mux.HandleFunc("/api/playlist", a.handlePlaylist)
	mux.HandleFunc("/api/video/info", a.handleVideoInfo)
	mux.HandleFunc("/api/stream", a.handleStream)
	mux.HandleFunc("/api/stream/preload", a.handleStreamPreload)
	mux.HandleFunc("/api/network/info", a.handleNetworkInfo)
	mux.HandleFunc("/api/account-mode", a.handleAccountMode)
	mux.HandleFunc("/api/play", a.handlePlay)
	mux.HandleFunc("/api/stop", a.handleStop)
	mux.HandleFunc("/api/player/status", a.handlePlayerStatus)
	mux.HandleFunc("/api/player/command", a.handlePlayerCommand)
	mux.HandleFunc("/api/account/profile", a.handleGoogleProfile)
	mux.HandleFunc("/api/account/login", a.handleGoogleLogin)
	mux.HandleFunc("/api/account/logout", a.handleGoogleLogout)
	mux.HandleFunc("/api/comments", a.handleComments)
	mux.HandleFunc("/api/comments/create", a.handleCreateComment)
	mux.HandleFunc("/api/downloads/start", a.handleStartDownload)
	mux.HandleFunc("/api/downloads/list", a.handleListDownloads)
	mux.HandleFunc("/api/downloads/delete", a.handleDeleteDownload)
	mux.HandleFunc("/api/downloads/config", a.handleDownloadConfig)
	mux.HandleFunc("/api/downloads/play", a.handleStreamLocalDownload)
	mux.HandleFunc("/api/open-external", a.handleOpenExternal)
	mux.HandleFunc("/api/ping", a.handlePing)

	return mux
}

func (a *App) handleFeed(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	isRefresh := r.URL.Query().Get("refresh") == "1" || r.URL.Query().Get("refresh") == "true"

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize := 24
	startItem := (page - 1) * pageSize + 1
	var target string

	switch tab {
	case "home":
		if category != "" {
			target = fmt.Sprintf("ytsearch%d:%s indonesia", startItem+pageSize-1, category)
		} else {
			if (a.accountMode == "firefox" || a.accountMode == "google") && a.firefoxProfile != "" {
				target = "https://www.youtube.com/"
			} else if a.accountMode == "custom" && a.customCookies != "" {
				target = "https://www.youtube.com/"
			} else {
				target = fmt.Sprintf("ytsearch%d:video populer viral indonesia", startItem+pageSize-1)
			}
		}
	case "shorts":
		target = fmt.Sprintf("ytsearch%d:#shorts trending indonesia", startItem+pageSize-1)
	case "subscriptions":
		target = "https://www.youtube.com/feed/subscriptions"
	case "history":
		target = "https://www.youtube.com/feed/history"
	case "watchlater":
		target = "https://www.youtube.com/playlist?list=WL"
	case "liked":
		target = "https://www.youtube.com/playlist?list=LL"
	default:
		target = "https://www.youtube.com/feed/trending"
	}

	videos, err := a.runYtDlp(target, pageSize, startItem, isRefresh)
	// Fallback jika feed utama kosong pada halaman 1
	if (err != nil || len(videos) == 0) && tab == "home" && page == 1 && target != "ytsearch24:trending indonesia" {
		log.Printf("[FX Tube] Home feed dari %s nihil/gagal (%v), beralih ke trending indonesia...\n", target, err)
		target = "ytsearch24:trending indonesia"
		videos, err = a.runYtDlp(target, pageSize, startItem, isRefresh)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}, "page": page})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos, "page": page})
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	isRefresh := r.URL.Query().Get("refresh") == "1" || r.URL.Query().Get("refresh") == "true"
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize := 20
	startItem := (page - 1) * pageSize + 1

	w.Header().Set("Content-Type", "application/json")
	if query == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"videos": []VideoItem{}, "page": page})
		return
	}

	target := fmt.Sprintf("ytsearch%d:%s", startItem+pageSize-1, query)
	videos, err := a.runYtDlp(target, pageSize, startItem, isRefresh)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}, "page": page})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos, "page": page})
}

func (a *App) handleVideoInfo(w http.ResponseWriter, r *http.Request) {
	targetURL := normalizeYouTubeURL(r.URL.Query().Get("url"))
	w.Header().Set("Content-Type", "application/json")
	if targetURL == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"qualities": []string{"360", "480", "720"}})
		return
	}

	qualities, err := a.getVideoAvailableQualities(targetURL)
	if err != nil || len(qualities) == 0 {
		qualities = []string{"360", "480", "720"}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url":       targetURL,
		"qualities": qualities,
	})
}

func (a *App) handleStreamPreload(w http.ResponseWriter, r *http.Request) {
	targetURL := normalizeYouTubeURL(r.URL.Query().Get("url"))
	quality := strings.TrimSpace(r.URL.Query().Get("quality"))
	if quality == "" {
		quality = "480"
	}
	if targetURL != "" {
		go func() {
			_, _ = a.resolveStreamURLs(targetURL, quality)
			_, _ = a.getVideoAvailableQualities(targetURL)
		}()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handleNetworkInfo(w http.ResponseWriter, r *http.Request) {
	ip := getLocalIP()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"local_ip":   ip,
		"port":       DefaultPort,
		"mobile_url": fmt.Sprintf("http://%s:%s", ip, DefaultPort),
	})
}

func (a *App) handleAccountMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req AccountModeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Mode != "" {
			a.accountMode = req.Mode
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".config", "fxtube")
			_ = os.MkdirAll(dir, 0755)
			_ = os.WriteFile(filepath.Join(dir, "account_mode.txt"), []byte(req.Mode), 0644)

			if req.Mode == "firefox" {
				a.profileMu.Lock()
				a.googleProfile = GoogleProfile{
					Name:       "Firefox YouTube",
					Handle:     "@firefox_sync",
					Avatar:     "https://www.gstatic.com/youtube/img/creator/avatar/creator_avatar_default.png",
					ChannelURL: "https://www.youtube.com",
					IsLoggedIn: true,
				}
				a.profileMu.Unlock()
				_ = a.saveProfile()
			} else if req.Mode == "guest" {
				a.profileMu.Lock()
				a.googleProfile = GoogleProfile{
					Name:       "Tamu",
					Handle:     "@guest",
					Avatar:     "",
					ChannelURL: "",
					IsLoggedIn: false,
				}
				a.profileMu.Unlock()
				_ = a.saveProfile()
			}
			a.feedMu.Lock()
			a.feedCache = make(map[string]FeedCacheItem)
			a.feedMu.Unlock()
			log.Printf("[FX Tube] Mode Akun diubah & disimpan ke: %s (Profil diupdate)\n", a.accountMode)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"mode":    a.accountMode,
		"profile": a.firefoxProfile,
	})
}

// handleOpenExternal launches a URL in the user's default desktop browser
func (a *App) handleOpenExternal(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing url", http.StatusBadRequest)
		return
	}

	log.Printf("[FX Tube] Membuka URL di browser sistem: %s\n", targetURL)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	cmd.Env = os.Environ()
	_ = cmd.Start()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	req.URL = normalizeYouTubeURL(req.URL)

	if err := a.startMpv(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handleStop(w http.ResponseWriter, r *http.Request) {
	a.stopCurrentPlayback()
	a.stopCurrentRemux()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (a *App) handlePlayerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := a.getPlayerStatus()
	json.NewEncoder(w).Encode(status)
}

func (a *App) handlePlayerCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayerCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	a.executePlayerCommand(req.Command)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	videoURL := normalizeYouTubeURL(r.URL.Query().Get("url"))
	videoID := strings.TrimSpace(r.URL.Query().Get("id"))

	var target string
	if strings.Contains(videoURL, "list=") {
		target = videoURL
	} else if videoID != "" {
		target = fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=RD%s", videoID, videoID)
	} else if videoURL != "" {
		if strings.Contains(videoURL, "v=") {
			parts := strings.Split(videoURL, "v=")
			id := strings.Split(parts[1], "&")[0]
			target = fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=RD%s", id, id)
		} else {
			target = fmt.Sprintf("%s&list=RD", videoURL)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if target == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"videos": []VideoItem{}})
		return
	}

	videos, err := a.runYtDlp(target, 40, 1, false)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos})
}

func (a *App) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
