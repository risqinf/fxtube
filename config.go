package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

// getYtDlpPath returns the optimal path for the yt-dlp binary
func getYtDlpPath() string {
	if p, err := exec.LookPath("yt-dlp"); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	localBin := filepath.Join(home, ".config", "fxtube", "bin", "yt-dlp")
	if _, err := os.Stat(localBin); err == nil {
		return localBin
	}
	return "yt-dlp"
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
