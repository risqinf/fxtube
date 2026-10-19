package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// runYtDlp executes yt-dlp to fetch video metadata with smart TTL caching and concurrency throttling
func (a *App) runYtDlp(target string, maxItems int) ([]VideoItem, error) {
	cacheKey := fmt.Sprintf("%s::%s::%d", a.accountMode, target, maxItems)

	// Check feed cache first for instant sub-millisecond response
	a.feedMu.RLock()
	if cached, ok := a.feedCache[cacheKey]; ok && time.Now().Before(cached.ExpiresAt) {
		a.feedMu.RUnlock()
		return cached.Videos, nil
	}
	a.feedMu.RUnlock()

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

	if maxItems > 0 {
		args = append(args, "--playlist-end", fmt.Sprintf("%d", maxItems))
	}

	args = append(args, target)

	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
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
			if item.URL == "" && item.ID != "" {
				item.URL = "https://www.youtube.com/watch?v=" + item.ID
			}
			items = append(items, item)
		}
	}

	_ = cmd.Wait()

	// If we got items (even if context timed out or process exited with non-zero), return them
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

	var format string
	switch quality {
	case "360":
		format = "bestvideo[height<=?360][vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo[height<=?360]+bestaudio/18/best[height<=?360]"
	case "480":
		format = "bestvideo[height<=?480][vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo[height<=?480]+bestaudio/best[height<=?480]"
	case "720":
		format = "bestvideo[height<=?720][vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo[height<=?720]+bestaudio/best[height<=?720]"
	default:
		format = "bestvideo[height<=?480][vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo[height<=?480]+bestaudio/best[height<=?480]"
	}

	args := []string{
		"-g",
		"-f", format,
		"--socket-timeout", "10",
		"--no-warnings",
	}
	cookieArgs := a.getCookieArgs()
	args = append(args, cookieArgs...)
	args = append(args, targetURL)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getYtDlpPath(), args...)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return StreamCacheItem{}, fmt.Errorf("resolusi stream URL timeout")
		}
		return StreamCacheItem{}, err
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
