package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
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
	mux.HandleFunc("/api/stream", a.handleStream)
	mux.HandleFunc("/api/account-mode", a.handleAccountMode)
	mux.HandleFunc("/api/play", a.handlePlay)
	mux.HandleFunc("/api/stop", a.handleStop)
	mux.HandleFunc("/api/player/status", a.handlePlayerStatus)
	mux.HandleFunc("/api/player/command", a.handlePlayerCommand)
	mux.HandleFunc("/api/ping", a.handlePing)

	return mux
}

func (a *App) handleFeed(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	var target string
	limit := 24

	switch tab {
	case "home":
		if category != "" {
			target = fmt.Sprintf("ytsearch24:%s", category)
		} else {
			if a.accountMode == "firefox" && a.firefoxProfile != "" {
				target = "https://www.youtube.com/"
			} else {
				target = "ytsearch24:trending indonesia"
			}
		}
	case "subscriptions":
		target = "https://www.youtube.com/feed/subscriptions"
	case "history":
		target = "https://www.youtube.com/feed/history"
	case "watchlater":
		target = "https://www.youtube.com/playlist?list=WL"
	case "liked":
		target = "https://www.youtube.com/playlist?list=LL"
	default:
		target = "ytsearch24:trending indonesia"
	}

	videos, err := a.runYtDlp(target, limit)
	// Fallback: jika Home feed via youtube.com kosong/gagal, fallback otomatis ke trending indonesia
	if (err != nil || len(videos) == 0) && tab == "home" && target == "https://www.youtube.com/" {
		log.Printf("[FX Tube] Home feed dari youtube.com nihil/gagal (%v), beralih ke trending indonesia...\n", err)
		videos, err = a.runYtDlp("ytsearch24:trending indonesia", limit)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos})
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "application/json")
	if query == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"videos": []VideoItem{}})
		return
	}

	target := fmt.Sprintf("ytsearch20:%s", query)
	videos, err := a.runYtDlp(target, 20)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos})
}

func (a *App) handleAccountMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req AccountModeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Mode != "" {
			a.accountMode = req.Mode
			log.Printf("[FX Tube] Mode Akun diubah ke: %s\n", a.accountMode)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"mode":    a.accountMode,
		"profile": a.firefoxProfile,
	})
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
	videoURL := strings.TrimSpace(r.URL.Query().Get("url"))
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
		} else if strings.Contains(videoURL, "youtu.be/") {
			parts := strings.Split(videoURL, "youtu.be/")
			id := strings.Split(parts[1], "?")[0]
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

	videos, err := a.runYtDlp(target, 25)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "videos": []VideoItem{}})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"videos": videos})
}

func (a *App) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
