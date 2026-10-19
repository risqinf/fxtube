package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// findFirefoxProfile locates the active default Firefox profile for cookie extraction across Linux, Windows, and macOS
func findFirefoxProfile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	searchDirs := []string{
		filepath.Join(home, ".config", "mozilla", "firefox"),
		filepath.Join(home, ".mozilla", "firefox"),
		filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"),
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		searchDirs = append(searchDirs, filepath.Join(appData, "Mozilla", "Firefox", "Profiles"))
	}

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), ".default-release") {
				return filepath.Join(dir, entry.Name())
			}
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(entry.Name(), "default") {
				return filepath.Join(dir, entry.Name())
			}
		}
	}
	return ""
}

// getYtDlpPath returns the optimal path for the yt-dlp binary (local bundled, system, or user directory)
func getYtDlpPath() string {
	binName := "yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
	}

	// 1. Check directory alongside running executable (bundled distribution)
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, binName)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate
		}
	}

	// 2. Check system PATH
	if p, err := exec.LookPath(binName); err == nil {
		return p
	}

	// 3. Check user config bin directory
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, ".config", "fxtube", "bin", binName)
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			candidate = filepath.Join(appData, "fxtube", "bin", binName)
		}
	}
	if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
		return candidate
	}

	return binName
}

// getFfmpegPath returns the optimal path for the ffmpeg binary
func getFfmpegPath() string {
	binName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binName = "ffmpeg.exe"
	}

	// 1. Check directory alongside running executable (bundled distribution)
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, binName)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate
		}
	}

	// 2. Check system PATH
	if p, err := exec.LookPath(binName); err == nil {
		return p
	}

	// 3. Check user config bin directory
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, ".config", "fxtube", "bin", binName)
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			candidate = filepath.Join(appData, "fxtube", "bin", binName)
		}
	}
	if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
		return candidate
	}

	return binName
}

// ensureYtDlp verifies yt-dlp is available, or downloads standalone binary automatically
func ensureYtDlp() {
	currentPath := getYtDlpPath()
	if _, err := exec.LookPath(currentPath); err == nil {
		return
	}
	if fi, err := os.Stat(currentPath); err == nil && !fi.IsDir() {
		return
	}

	binName := "yt-dlp"
	downloadURL := "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	} else if runtime.GOOS == "darwin" {
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	}

	home, _ := os.UserHomeDir()
	targetDir := filepath.Join(home, ".config", "fxtube", "bin")
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			targetDir = filepath.Join(appData, "fxtube", "bin")
		}
	}
	_ = os.MkdirAll(targetDir, 0755)
	targetFile := filepath.Join(targetDir, binName)

	log.Printf("[FX Tube] yt-dlp belum terpasang, mengunduh otomatis ke %s...\n", targetFile)
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("[FX Tube] Gagal mengunduh yt-dlp: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[FX Tube] Gagal mengunduh yt-dlp: status HTTP %d\n", resp.StatusCode)
		return
	}

	out, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Printf("[FX Tube] Gagal menulis berkas yt-dlp: %v\n", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("[FX Tube] Gagal menyimpan berkas yt-dlp: %v\n", err)
		return
	}
	_ = os.Chmod(targetFile, 0755)
	log.Printf("[FX Tube] yt-dlp berhasil dipasang mandiri di: %s\n", targetFile)
}

// getCookieArgs constructs the cookie parameters based on current account mode
func (a *App) getCookieArgs() []string {
	switch a.accountMode {
	case "guest":
		return []string{}
	case "custom":
		if a.customCookies != "" {
			if _, err := os.Stat(a.customCookies); err == nil {
				return []string{"--cookies", a.customCookies}
			}
		}
		return []string{}
	case "chrome":
		return []string{"--cookies-from-browser", "chrome"}
	case "brave":
		return []string{"--cookies-from-browser", "brave"}
	case "edge":
		return []string{"--cookies-from-browser", "edge"}
	default: // "firefox"
		if a.firefoxProfile != "" {
			return []string{"--cookies-from-browser", "firefox:" + a.firefoxProfile}
		}
		return []string{"--cookies-from-browser", "firefox"}
	}
}

// cleanBrowserProfile removes redundant cache directories to save disk space and RAM
func cleanBrowserProfile(profileDir string) {
	if profileDir == "" {
		return
	}
	junkDirs := []string{
		filepath.Join(profileDir, "Crashpad"),
		filepath.Join(profileDir, "GraphiteDawnCache"),
		filepath.Join(profileDir, "GrShaderCache"),
		filepath.Join(profileDir, "ShaderCache"),
		filepath.Join(profileDir, "Default", "GPUCache"),
	}
	for _, dir := range junkDirs {
		_ = os.RemoveAll(dir)
	}
}

// normalizeYouTubeURL converts Shorts and shortlinks to standard watch URLs
func normalizeYouTubeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "/shorts/") {
		parts := strings.Split(raw, "/shorts/")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "?")[0]
			id = strings.Split(id, "&")[0]
			return "https://www.youtube.com/watch?v=" + id
		}
	} else if strings.Contains(raw, "youtu.be/") {
		parts := strings.Split(raw, "youtu.be/")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "?")[0]
			id = strings.Split(id, "&")[0]
			return "https://www.youtube.com/watch?v=" + id
		}
	}
	return raw
}

// getLocalIP returns the primary non-loopback local IPv4 address for mobile LAN access
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
