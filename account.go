package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// getProfileConfigPath returns the path to the stored Google/YouTube profile JSON
func getProfileConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "fxtube")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "profile.json")
}

// loadProfile loads the saved Google profile from ~/.config/fxtube/profile.json
func (a *App) loadProfile() {
	path := getProfileConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		// Default to Guest profile
		a.profileMu.Lock()
		a.googleProfile = GoogleProfile{
			Name:       "Tamu",
			Handle:     "@guest",
			Avatar:     "",
			ChannelURL: "",
			IsLoggedIn: false,
		}
		a.profileMu.Unlock()
		return
	}

	var prof GoogleProfile
	if err := json.Unmarshal(data, &prof); err == nil {
		a.profileMu.Lock()
		a.googleProfile = prof
		a.profileMu.Unlock()
	}
}

// saveProfile persists the current Google profile to ~/.config/fxtube/profile.json
func (a *App) saveProfile() error {
	path := getProfileConfigPath()
	a.profileMu.RLock()
	data, err := json.MarshalIndent(a.googleProfile, "", "  ")
	mode := a.accountMode
	a.profileMu.RUnlock()
	if err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "fxtube")
	_ = os.MkdirAll(dir, 0755)
	if mode != "" {
		_ = os.WriteFile(filepath.Join(dir, "account_mode.txt"), []byte(mode), 0644)
	}
	return os.WriteFile(path, data, 0644)
}

// handleGoogleProfile returns the active Google/YouTube profile
func (a *App) handleGoogleProfile(w http.ResponseWriter, r *http.Request) {
	a.profileMu.RLock()
	prof := a.googleProfile
	a.profileMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"mode":    a.accountMode,
		"profile": prof,
	})
}

// handleGoogleLogin handles logging into Google/YouTube
func (a *App) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 1. Manual Profile provided
	if req.Type == "manual" && req.Name != "" {
		a.profileMu.Lock()
		a.googleProfile = GoogleProfile{
			Name:       req.Name,
			Handle:     req.Handle,
			Avatar:     req.Avatar,
			ChannelURL: req.ChannelURL,
			IsLoggedIn: true,
		}
		a.accountMode = "google"
		a.profileMu.Unlock()
		_ = a.saveProfile()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"profile": a.googleProfile,
			"mode":    a.accountMode,
		})
		return
	}

	// 2. Handle or Channel URL provided
	inputHandle := strings.TrimSpace(req.Handle)
	if inputHandle == "" && req.ChannelURL != "" {
		inputHandle = strings.TrimSpace(req.ChannelURL)
	}

	if inputHandle != "" {
		var targetURL string
		var cleanHandle string
		if strings.HasPrefix(inputHandle, "http://") || strings.HasPrefix(inputHandle, "https://") {
			targetURL = inputHandle
			cleanHandle = "@user"
		} else {
			if !strings.HasPrefix(inputHandle, "@") {
				cleanHandle = "@" + inputHandle
			} else {
				cleanHandle = inputHandle
			}
			targetURL = "https://www.youtube.com/" + cleanHandle
		}

		log.Printf("[FX Tube] Mengambil profil channel untuk %s...\n", targetURL)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "yt-dlp",
			"--dump-single-json",
			"--playlist-items", "0",
			"--no-warnings",
			targetURL,
		)

		var out bytes.Buffer
		cmd.Stdout = &out

		type ytThumbnail struct {
			URL string `json:"url"`
			ID  string `json:"id"`
		}
		type ytChannelInfo struct {
			Channel    string        `json:"channel"`
			Uploader   string        `json:"uploader"`
			ChannelURL string        `json:"channel_url"`
			Thumbnails []ytThumbnail `json:"thumbnails"`
		}

		var info ytChannelInfo
		if err := cmd.Run(); err == nil && out.Len() > 0 {
			_ = json.Unmarshal(out.Bytes(), &info)
		}

		channelName := info.Channel
		if channelName == "" {
			channelName = info.Uploader
		}
		if channelName == "" {
			channelName = strings.TrimPrefix(cleanHandle, "@")
		}

		channelURL := info.ChannelURL
		if channelURL == "" {
			channelURL = targetURL
		}

		avatar := ""
		if len(info.Thumbnails) > 0 {
			avatar = info.Thumbnails[len(info.Thumbnails)-1].URL
		}
		if avatar == "" {
			avatar = fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=dc2626&color=fff", channelName)
		}

		a.profileMu.Lock()
		a.googleProfile = GoogleProfile{
			Name:       channelName,
			Handle:     cleanHandle,
			Avatar:     avatar,
			ChannelURL: channelURL,
			IsLoggedIn: true,
		}
		a.accountMode = "google"
		a.profileMu.Unlock()
		_ = a.saveProfile()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"profile": a.googleProfile,
			"mode":    a.accountMode,
		})
		return
	}

	// 3. Browser session import (Firefox / Chrome / Brave)
	browser := strings.ToLower(req.Browser)
	if browser == "" {
		browser = "firefox"
	}

	a.accountMode = browser
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "fxtube")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "account_mode.txt"), []byte(a.accountMode), 0644)

	a.feedMu.Lock()
	a.feedCache = make(map[string]FeedCacheItem)
	a.feedMu.Unlock()

	a.profileMu.Lock()
	a.googleProfile = GoogleProfile{
		Name:       fmt.Sprintf("Pengguna %s", strings.Title(browser)),
		Handle:     fmt.Sprintf("@%s_sync", browser),
		Avatar:     "https://www.gstatic.com/youtube/img/creator/avatar/creator_avatar_default.png",
		ChannelURL: "https://www.youtube.com",
		IsLoggedIn: true,
	}
	a.profileMu.Unlock()
	_ = a.saveProfile()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"profile": a.googleProfile,
		"mode":    a.accountMode,
	})
}

// handleGoogleLogout logs out of Google and switches back to Guest mode
func (a *App) handleGoogleLogout(w http.ResponseWriter, r *http.Request) {
	a.profileMu.Lock()
	a.googleProfile = GoogleProfile{
		Name:       "Tamu",
		Handle:     "@guest",
		Avatar:     "",
		ChannelURL: "",
		IsLoggedIn: false,
	}
	a.accountMode = "guest"
	a.profileMu.Unlock()
	_ = a.saveProfile()

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "fxtube")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "account_mode.txt"), []byte("guest"), 0644)

	a.feedMu.Lock()
	a.feedCache = make(map[string]FeedCacheItem)
	a.feedMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"mode":    "guest",
		"profile": a.googleProfile,
	})
}

// handleComments returns the comments for a given video
func (a *App) handleComments(w http.ResponseWriter, r *http.Request) {
	videoURL := normalizeYouTubeURL(r.URL.Query().Get("url"))
	videoID := strings.TrimSpace(r.URL.Query().Get("id"))

	if videoID == "" && videoURL != "" {
		if strings.Contains(videoURL, "v=") {
			parts := strings.Split(videoURL, "v=")
			videoID = strings.Split(parts[1], "&")[0]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if videoID == "" && videoURL == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"comments": []CommentItem{}})
		return
	}

	cacheKey := videoID
	if cacheKey == "" {
		cacheKey = videoURL
	}

	a.commentsMu.RLock()
	cached, ok := a.commentsCache[cacheKey]
	a.commentsMu.RUnlock()

	if ok && len(cached) > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"comments": cached})
		return
	}

	target := videoURL
	if target == "" {
		target = "https://www.youtube.com/watch?v=" + videoID
	}

	comments := a.fetchVideoComments(target)

	a.commentsMu.Lock()
	if a.commentsCache == nil {
		a.commentsCache = make(map[string][]CommentItem)
	}
	a.commentsCache[cacheKey] = comments
	a.commentsMu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{"comments": comments})
}

// handleCreateComment adds a comment to the video
func (a *App) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Text) == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	videoID := req.VideoID
	if videoID == "" && req.VideoURL != "" {
		if strings.Contains(req.VideoURL, "v=") {
			parts := strings.Split(req.VideoURL, "v=")
			videoID = strings.Split(parts[1], "&")[0]
		} else {
			videoID = req.VideoURL
		}
	}

	a.profileMu.RLock()
	author := a.googleProfile.Name
	avatar := a.googleProfile.Avatar
	if author == "" {
		author = "Pengguna FX Tube"
	}
	if avatar == "" {
		avatar = fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=dc2626&color=fff", author)
	}
	a.profileMu.RUnlock()

	newComment := CommentItem{
		Author: author,
		Avatar: avatar,
		Text:   strings.TrimSpace(req.Text),
		Time:   "Baru saja",
		Likes:  "0",
	}

	a.commentsMu.Lock()
	if a.commentsCache == nil {
		a.commentsCache = make(map[string][]CommentItem)
	}
	// Prepend new comment
	a.commentsCache[videoID] = append([]CommentItem{newComment}, a.commentsCache[videoID]...)
	a.commentsMu.Unlock()

	log.Printf("[FX Tube] Komentar baru ditambahkan untuk video %s oleh %s: %s\n", videoID, author, newComment.Text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"comment": newComment,
	})
}

// fetchVideoComments retrieves YouTube comments using yt-dlp
func (a *App) fetchVideoComments(videoURL string) []CommentItem {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--dump-json",
		"--get-comments",
		"--extractor-args", "youtube:max_comments=25",
		"--playlist-items", "0",
		"--no-warnings",
		videoURL,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil || out.Len() == 0 {
		return []CommentItem{}
	}

	type ytComment struct {
		Author          string `json:"author"`
		AuthorThumbnail string `json:"author_thumbnail"`
		Text            string `json:"text"`
		TimeText        string `json:"_time_text"`
		LikeCount       int    `json:"like_count"`
	}

	type ytResp struct {
		Comments []ytComment `json:"comments"`
	}

	var parsed ytResp
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return []CommentItem{}
	}

	comments := make([]CommentItem, 0, len(parsed.Comments))
	for _, c := range parsed.Comments {
		if strings.TrimSpace(c.Text) == "" {
			continue
		}
		likes := ""
		if c.LikeCount > 0 {
			likes = fmt.Sprintf("%d", c.LikeCount)
		}
		timeText := c.TimeText
		if timeText == "" {
			timeText = "Baru saja"
		}
		comments = append(comments, CommentItem{
			Author: c.Author,
			Avatar: c.AuthorThumbnail,
			Text:   c.Text,
			Time:   timeText,
			Likes:  likes,
		})
	}

	return comments
}
