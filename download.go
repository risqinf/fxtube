package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// getDefaultDownloadDir returns ~/Downloads/fxtube
func getDefaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	// Check if custom download directory is saved
	configPath := filepath.Join(home, ".config", "fxtube", "download_dir.txt")
	if data, err := os.ReadFile(configPath); err == nil {
		custom := strings.TrimSpace(string(data))
		if custom != "" {
			_ = os.MkdirAll(custom, 0755)
			return custom
		}
	}

	def := filepath.Join(home, "Downloads", "fxtube")
	_ = os.MkdirAll(def, 0755)
	return def
}

func (a *App) initDownloads() {
	a.downloadMu.Lock()
	defer a.downloadMu.Unlock()

	a.downloadDir = getDefaultDownloadDir()
	a.downloads = make(map[string]*DownloadItem)

	home, _ := os.UserHomeDir()
	metaPath := filepath.Join(home, ".config", "fxtube", "downloads.json")
	if data, err := os.ReadFile(metaPath); err == nil {
		var list []*DownloadItem
		if err := json.Unmarshal(data, &list); err == nil {
			for _, item := range list {
				// Verify file still exists on disk
				if item.FilePath != "" {
					if fi, err := os.Stat(item.FilePath); err == nil && !fi.IsDir() {
						item.FileSize = fi.Size()
						item.FileSizeText = formatBytes(fi.Size())
						item.Status = "completed"
						item.Progress = 100
						a.downloads[item.ID] = item
					}
				}
			}
		}
	}
}

func (a *App) saveDownloads() {
	a.downloadMu.RLock()
	list := make([]*DownloadItem, 0, len(a.downloads))
	for _, item := range a.downloads {
		list = append(list, item)
	}
	a.downloadMu.RUnlock()

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "fxtube")
	_ = os.MkdirAll(dir, 0755)
	metaPath := filepath.Join(dir, "downloads.json")

	data, err := json.MarshalIndent(list, "", "  ")
	if err == nil {
		_ = os.WriteFile(metaPath, data, 0644)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// handleDownloadConfig returns or updates the download directory
func (a *App) handleDownloadConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var req DownloadConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && strings.TrimSpace(req.DownloadDir) != "" {
			newDir := strings.TrimSpace(req.DownloadDir)
			if err := os.MkdirAll(newDir, 0755); err == nil {
				a.downloadMu.Lock()
				a.downloadDir = newDir
				a.downloadMu.Unlock()

				home, _ := os.UserHomeDir()
				_ = os.WriteFile(filepath.Join(home, ".config", "fxtube", "download_dir.txt"), []byte(newDir), 0644)
				log.Printf("[FX Tube] Download directory diubah ke: %s\n", newDir)
			}
		}
	}

	a.downloadMu.RLock()
	dir := a.downloadDir
	a.downloadMu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"download_dir": dir,
	})
}

// handleListDownloads lists all downloaded and downloading videos
func (a *App) handleListDownloads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	a.downloadMu.RLock()
	list := make([]*DownloadItem, 0, len(a.downloads))
	for _, item := range a.downloads {
		list = append(list, item)
	}
	currentDir := a.downloadDir
	a.downloadMu.RUnlock()

	// Sort by newest first
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt > list[j].CreatedAt
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"downloads":    list,
		"download_dir": currentDir,
	})
}

// handleStartDownload initiates video download to local storage
func (a *App) handleStartDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.URL == "" && req.ID == "") {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.URL == "" && req.ID != "" {
		req.URL = "https://www.youtube.com/watch?v=" + req.ID
	}
	req.URL = normalizeYouTubeURL(req.URL)

	quality := req.Quality
	if quality == "" {
		quality = "720"
	}

	downloadKey := req.ID
	if downloadKey == "" {
		downloadKey = req.URL
	}
	downloadKey = fmt.Sprintf("%s_%s", downloadKey, quality)

	a.downloadMu.Lock()
	if existing, exists := a.downloads[downloadKey]; exists && existing.Status == "downloading" {
		a.downloadMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"download": existing,
			"message":  "Video sedang dalam proses unduhan",
		})
		return
	}

	item := &DownloadItem{
		ID:        downloadKey,
		Title:     req.Title,
		Channel:   req.Channel,
		Thumbnail: req.Thumbnail,
		Quality:   quality,
		Status:    "downloading",
		Progress:  0,
		CreatedAt: time.Now().Unix(),
	}
	a.downloads[downloadKey] = item
	targetDir := a.downloadDir
	a.downloadMu.Unlock()

	log.Printf("[FX Tube] Memulai unduhan: %s (%s) ke %s\n", req.Title, quality, targetDir)

	go a.executeDownload(item, req.URL, quality, targetDir)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"download": item,
		"message":  "Unduhan dimulai di latar belakang",
	})
}

func (a *App) executeDownload(item *DownloadItem, targetURL, quality, targetDir string) {
	_ = os.MkdirAll(targetDir, 0755)

	var formatString string
	if quality == "audio" {
		formatString = "bestaudio[ext=m4a]/bestaudio/best"
	} else {
		// Bypass restrictions prioritizing AVC1 H.264
		formatString = fmt.Sprintf("bestvideo[vcodec^=avc1][height<=?%s]+bestaudio[acodec^=mp4a]/bestvideo[height<=?%s]+bestaudio/best[height<=?%s]/best", quality, quality, quality)
	}

	outputTemplate := filepath.Join(targetDir, "%(title)s [%(id)s].%(ext)s")

	// Multi-client extractor args to bypass restrictions
	args := []string{
		"--extractor-args", "youtube:player_client=android,web,tv_embedded",
		"-f", formatString,
		"--merge-output-format", "mp4",
		"-o", outputTemplate,
		"--newline",
		"--no-playlist",
		"--no-warnings",
	}

	if a.accountMode == "firefox" && a.firefoxProfile != "" {
		args = append(args, "--cookies-from-browser", "firefox:"+a.firefoxProfile)
	} else if a.customCookies != "" {
		if fi, err := os.Stat(a.customCookies); err == nil && !fi.IsDir() {
			args = append(args, "--cookies", a.customCookies)
		}
	}

	args = append(args, targetURL)

	cmd := exec.Command("yt-dlp", args...)
	cmd.Env = os.Environ()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.updateDownloadError(item, err.Error())
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		a.updateDownloadError(item, err.Error())
		return
	}

	progressRegex := regexp.MustCompile(`\[download\]\s+([\d\.]+)%`)
	destRegex := regexp.MustCompile(`\[(?:download|Merger)\]\s+(?:Destination:\s+|Merging formats into\s+")([^"]+)`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse progress
		if m := progressRegex.FindStringSubmatch(line); len(m) > 1 {
			if pct, err := strconv.ParseFloat(m[1], 64); err == nil {
				a.downloadMu.Lock()
				item.Progress = pct
				a.downloadMu.Unlock()
			}
		}

		// Parse file destination
		if m := destRegex.FindStringSubmatch(line); len(m) > 1 {
			filePath := strings.Trim(m[1], "\"")
			a.downloadMu.Lock()
			item.FilePath = filePath
			item.FileName = filepath.Base(filePath)
			a.downloadMu.Unlock()
		}
	}

	if err := cmd.Wait(); err != nil {
		a.updateDownloadError(item, "Gagal mengunduh: "+err.Error())
		return
	}

	// Locate resulting file if not detected
	if item.FilePath == "" || !fileExists(item.FilePath) {
		files, _ := os.ReadDir(targetDir)
		for _, f := range files {
			if strings.Contains(f.Name(), item.Quality) || strings.HasSuffix(f.Name(), ".mp4") {
				candidate := filepath.Join(targetDir, f.Name())
				if fi, err := os.Stat(candidate); err == nil && time.Since(fi.ModTime()) < 1*time.Minute {
					item.FilePath = candidate
					item.FileName = f.Name()
					break
				}
			}
		}
	}

	a.downloadMu.Lock()
	if item.FilePath != "" && fileExists(item.FilePath) {
		fi, _ := os.Stat(item.FilePath)
		item.FileSize = fi.Size()
		item.FileSizeText = formatBytes(fi.Size())
		item.FileName = filepath.Base(item.FilePath)
		item.Status = "completed"
		item.Progress = 100
		log.Printf("[FX Tube] Unduhan selesai: %s (%s)\n", item.FileName, item.FileSizeText)
	} else {
		item.Status = "completed"
		item.Progress = 100
	}
	a.downloadMu.Unlock()

	a.saveDownloads()
}

func (a *App) updateDownloadError(item *DownloadItem, errMsg string) {
	a.downloadMu.Lock()
	item.Status = "error"
	item.Error = errMsg
	a.downloadMu.Unlock()
	log.Printf("[FX Tube] Error unduhan %s: %s\n", item.ID, errMsg)
	a.saveDownloads()
}

// handleDeleteDownload deletes an offline downloaded video
func (a *App) handleDeleteDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	a.downloadMu.Lock()
	item, ok := a.downloads[req.ID]
	if ok {
		if item.FilePath != "" && fileExists(item.FilePath) {
			_ = os.Remove(item.FilePath)
		}
		delete(a.downloads, req.ID)
	}
	a.downloadMu.Unlock()

	a.saveDownloads()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Berkas unduhan berhasil dihapus",
	})
}

// handleStreamLocalDownload serves the local video file with Range support
func (a *App) handleStreamLocalDownload(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	a.downloadMu.RLock()
	item, ok := a.downloads[id]
	var filePath string
	if ok && item.FilePath != "" {
		filePath = item.FilePath
	}
	a.downloadMu.RUnlock()

	if filePath == "" || !fileExists(filePath) {
		http.Error(w, "Video tidak ditemukan di penyimpanan lokal", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, filePath)
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}
