package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// runYtDlp executes yt-dlp to fetch video metadata with smart TTL caching, pagination, and concurrency throttling
func (a *App) runYtDlp(target string, maxItems int, startItem int, forceRefresh bool) ([]VideoItem, error) {
	if startItem <= 0 {
		startItem = 1
	}
	cacheKey := fmt.Sprintf("%s::%s::%d::%d", a.accountMode, target, startItem, maxItems)

	// Check feed cache first unless explicitly requested to bypass
	if !forceRefresh {
		a.feedMu.RLock()
		if cached, ok := a.feedCache[cacheKey]; ok && time.Now().Before(cached.ExpiresAt) {
			a.feedMu.RUnlock()
			return cached.Videos, nil
		}
		a.feedMu.RUnlock()
	}

	// Throttle concurrency to protect legacy dual-core CPU from load spikes
	a.ytdlpSemaphore <- struct{}{}
	defer func() { <-a.ytdlpSemaphore }()

	args := []string{
		"--flat-playlist",
		"--dump-json",
		"--socket-timeout", "10",
		"--no-warnings",
	}

	cookieArgs := a.getCookieArgs()
	args = append(args, cookieArgs...)

	if startItem > 1 {
		args = append(args, "--playlist-start", fmt.Sprintf("%d", startItem))
	}
	if maxItems > 0 {
		endItem := startItem + maxItems - 1
		args = append(args, "--playlist-end", fmt.Sprintf("%d", endItem))
	}

	args = append(args, target)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getYtDlpPath(), args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var items []VideoItem
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 512*1024)
	scanner.Buffer(buf, 8*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var item VideoItem
		if err := json.Unmarshal(line, &item); err == nil && item.Title != "" {
			// Skip YouTube Mix playlists or non-video items
			if strings.Contains(item.URL, "playlist?list=") || strings.Contains(item.ID, "RD") {
				continue
			}
			if strings.Contains(item.URL, "/shorts/") {
				parts := strings.Split(item.URL, "/shorts/")
				if len(parts) > 1 {
					id := strings.Split(parts[1], "?")[0]
					item.ID = id
					item.URL = "https://www.youtube.com/watch?v=" + id
				}
				item.IsShorts = true
			}
			if item.URL == "" && item.ID != "" {
				item.URL = "https://www.youtube.com/watch?v=" + item.ID
			}
			titleLower := strings.ToLower(item.Title)
			if strings.Contains(titleLower, "#shorts") || strings.Contains(titleLower, "#short") || (item.Duration > 0 && item.Duration <= 61) {
				item.IsShorts = true
			}
			items = append(items, item)
		}
	}

	_ = cmd.Wait()

	if len(items) > 0 {
		a.feedMu.Lock()
		a.feedCache[cacheKey] = FeedCacheItem{
			Videos:    items,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		a.feedMu.Unlock()
		return items, nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[FX Tube] yt-dlp timeout untuk target: %s\n", target)
		return nil, fmt.Errorf("waktu pengambilan video habis (timeout)")
	}

	return items, nil
}

// getVideoAvailableQualities returns available video heights capped at 1080p
func (a *App) getVideoAvailableQualities(targetURL string) ([]string, error) {
	a.qualitiesMu.RLock()
	if q, ok := a.qualitiesCache[targetURL]; ok && len(q) > 0 {
		a.qualitiesMu.RUnlock()
		return q, nil
	}
	a.qualitiesMu.RUnlock()

	a.ytdlpSemaphore <- struct{}{}
	defer func() { <-a.ytdlpSemaphore }()

	args := []string{
		"--print", "%(formats.:.height)s",
		"--socket-timeout", "10",
		"--no-warnings",
	}
	cookieArgs := a.getCookieArgs()
	args = append(args, cookieArgs...)
	args = append(args, targetURL)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getYtDlpPath(), args...)
	out, err := cmd.Output()
	if err != nil {
		// Default fallback qualities if query fails
		return []string{"360", "480", "720"}, nil
	}

	// Parse heights array like [144, 240, 360, 480, 720, 1080, 1440, 2160]
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(string(out), -1)
	heightSet := make(map[int]bool)
	maxFound := 0
	for _, m := range matches {
		if val, err := strconv.Atoi(m); err == nil && val > 0 {
			heightSet[val] = true
			if val > maxFound {
				maxFound = val
			}
		}
	}

	allStandard := []string{"144", "240", "360", "480", "720", "1080"}
	var available []string
	for _, qStr := range allStandard {
		hVal, _ := strconv.Atoi(qStr)
		if heightSet[hVal] || (maxFound >= hVal) {
			available = append(available, qStr)
		}
	}

	if len(available) == 0 {
		available = []string{"360", "480", "720"}
	}

	a.qualitiesMu.Lock()
	a.qualitiesCache[targetURL] = available
	a.qualitiesMu.Unlock()

	return available, nil
}

// resolveStreamURLs retrieves the direct Google CDN video and audio stream URLs
func (a *App) resolveStreamURLs(targetURL, quality string) (StreamCacheItem, error) {
	cacheKey := targetURL + "::" + quality

	a.streamMu.RLock()
	if item, found := a.streamCache[cacheKey]; found && time.Now().Before(item.ExpiresAt) {
		a.streamMu.RUnlock()
		return item, nil
	}
	a.streamMu.RUnlock()

	a.ytdlpSemaphore <- struct{}{}
	defer func() { <-a.ytdlpSemaphore }()

	// Double check cache after semaphore acquisition
	a.streamMu.RLock()
	if item, found := a.streamCache[cacheKey]; found && time.Now().Before(item.ExpiresAt) {
		a.streamMu.RUnlock()
		return item, nil
	}
	a.streamMu.RUnlock()

	maxHeight := "480"
	switch quality {
	case "144", "240", "360", "480", "720", "1080":
		maxHeight = quality
	}

	// High compatibility format string with reliable fallbacks
	format := fmt.Sprintf(
		"bestvideo[height<=?%s][vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo[height<=?%s][vcodec^=avc1]+bestaudio/best[height<=?%s][vcodec^=avc1]/best[height<=?%s][ext=mp4]/bestvideo[height<=?%s]+bestaudio/best",
		maxHeight, maxHeight, maxHeight, maxHeight, maxHeight,
	)

	args := []string{
		"-g",
		"-f", format,
		"--socket-timeout", "12",
		"--no-warnings",
	}
	cookieArgs := a.getCookieArgs()
	args = append(args, cookieArgs...)
	args = append(args, targetURL)

	ctx, cancel := context.WithTimeout(context.Background(), 16*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getYtDlpPath(), args...)
	out, err := cmd.Output()

	// If primary format string fails, try ultra-permissive fallback
	if err != nil || len(out) == 0 {
		fallbackArgs := []string{
			"-g",
			"-f", fmt.Sprintf("b[height<=?%s]/bv*[height<=?%s]+ba/b", maxHeight, maxHeight),
			"--socket-timeout", "12",
			"--no-warnings",
		}
		fallbackArgs = append(fallbackArgs, cookieArgs...)
		fallbackArgs = append(fallbackArgs, targetURL)

		ctxFallback, cancelFallback := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancelFallback()
		fallbackCmd := exec.CommandContext(ctxFallback, getYtDlpPath(), fallbackArgs...)
		out, err = fallbackCmd.Output()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded || ctxFallback.Err() == context.DeadlineExceeded {
				return StreamCacheItem{}, fmt.Errorf("resolusi stream URL timeout")
			}
			return StreamCacheItem{}, err
		}
	}

	rawLines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var lines []string
	for _, l := range rawLines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "http") {
			lines = append(lines, t)
		}
	}

	if len(lines) == 0 {
		return StreamCacheItem{}, fmt.Errorf("no stream URL resolved")
	}

	item := StreamCacheItem{
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}

	if len(lines) == 1 {
		item.VideoURL = lines[0]
		item.IsCombined = true
	} else {
		item.VideoURL = lines[0]
		item.AudioURL = lines[1]
		item.IsCombined = false
	}

	a.streamMu.Lock()
	a.streamCache[cacheKey] = item
	a.streamMu.Unlock()

	return item, nil
}
