package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// launchGUI opens the FX Tube application window in the most lightweight and hardware-safe mode
func launchGUI(port string) {
	url := "http://127.0.0.1:" + port

	home, _ := os.UserHomeDir()
	profileDir := filepath.Join(home, ".config", "fxtube", "browser-profile")
	cleanBrowserProfile(profileDir)

	// 1. Windows platform handling
	if runtime.GOOS == "windows" {
		windowsBrowsers := []string{
			"msedge",
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
			"chrome",
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			"brave",
			"firefox",
		}
		for _, b := range windowsBrowsers {
			if path, err := exec.LookPath(b); err == nil {
				cmd := exec.Command(path, "--app="+url, "--user-data-dir="+profileDir)
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}
		// Fallback for Windows
		_ = exec.Command("cmd", "/c", "start", url).Start()
		return
	}

	// 2. macOS platform handling
	if runtime.GOOS == "darwin" {
		macAppPaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, path := range macAppPaths {
			if _, err := os.Stat(path); err == nil {
				cmd := exec.Command(path, "--app="+url, "--user-data-dir="+profileDir)
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}
		// Fallback for macOS
		_ = exec.Command("open", url).Start()
		return
	}

	// 3. Linux platform handling: detect ANY installed Chromium-based browser
	linuxChromiumCandidates := []string{
		"thorium-browser",
		"google-chrome-stable",
		"google-chrome",
		"chromium-browser",
		"chromium",
		"brave-browser",
		"brave",
		"microsoft-edge-stable",
		"microsoft-edge",
	}

	for _, browserName := range linuxChromiumCandidates {
		if path, err := exec.LookPath(browserName); err == nil {
			cmd := exec.Command(path,
				"--app="+url,
				"--user-data-dir="+profileDir,
				"--class=fxtube",
				"--password-store=basic",
				"--no-first-run",
				"--disable-sync",
				// Wayland native (bypasses Xwayland overhead if Wayland is present)
				"--ozone-platform-hint=auto",
				// Prevent GPU hang by enforcing rock-solid 2D software rendering
				"--disable-gpu",
				"--disable-gpu-compositing",
				"--disable-accelerated-video-decode",
				"--disable-accelerated-video-encode",
				"--disable-gpu-rasterization",
				// Extreme memory optimizations for legacy hardware
				"--renderer-process-limit=2",
				"--disable-extensions",
				"--disable-background-networking",
				"--disable-default-apps",
				"--no-default-browser-check",
			)
			if err := cmd.Start(); err == nil {
				return
			}
		}
	}

	// Fallback to Firefox if no Chromium browser is present
	if path, err := exec.LookPath("firefox"); err == nil {
		cmd := exec.Command(path, "--new-window", url)
		if err := cmd.Start(); err == nil {
			return
		}
	}

	// Final fallback: standard system web launcher
	_ = exec.Command("xdg-open", url).Start()
}
