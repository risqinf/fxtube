# FX Tube 🚀

**Ultra-Lightweight, Zero-Bloat, 100% Ad-Free YouTube Client for Legacy & Modern Hardware.**

FX Tube is a high-performance, minimalist YouTube desktop client designed to run smoothly on legacy hardware (such as ThinkPad X201, Core 2 Duo, 1st Gen Core i-series) as well as modern Linux, Windows, and macOS systems.

---

## ✨ Features

- **🛡️ 100% Ad-Free**: Direct stream extraction completely bypasses YouTube's JavaScript ad-injection, mid-rolls, sponsor cards, and tracking scripts.
- **🌐 Embedded Web Player (Default)**: Clean, embedded HTML5 player with interactive timeline scrubber, -10s / +10s seek controls, and resolution switching (360p, 480p, 720p).
- **⚡ MPV Native Player Support**: One-click pop-out to ultra-lightweight MPV player (~40 MB RAM usage vs 600+ MB for standard web browsers).
- **🎶 Real YouTube Mix & Playlist Queue**: Seamlessly loads the true ~25-track YouTube Mix (`RD<id>`) or native playlist for non-stop autoplay.
- **🔒 Flexible Account Sync (No Password Required)**:
  - **Firefox / Chrome / Brave / Edge Sync**: Automatically reads existing YouTube cookies without entering your credentials.
  - **cookies.txt**: Supports custom exported Netscape cookie files.
  - **Guest Mode**: Privacy-focused mode with zero tracking.
- **💻 Cross-Platform & Self-Contained**: Single static binary with zero external runtime dependencies (`CGO_ENABLED=0`).
- **🪟 No Heavy Browser Required**: Automatically uses any existing browser already installed on your system (Edge on Windows, Chrome/Brave on macOS, Firefox/Chromium on Linux).

---

## 🛠️ System Requirements

FX Tube requires only two lightweight CLI tools installed in your system PATH:
1. [yt-dlp](https://github.com/yt-dlp/yt-dlp) (for YouTube metadata and direct stream resolution)
2. [ffmpeg](https://ffmpeg.org/) (for on-the-fly DASH stream remuxing)
3. *(Optional)* [mpv](https://mpv.io/) (if you wish to use the native MPV player mode)

---

## 📥 Installation & Running

### Linux
```bash
# Download and make executable
chmod +x fxtube-linux-amd64
./fxtube-linux-amd64
```

### Windows
1. Download `fxtube-windows-amd64.zip`.
2. Extract and double-click `fxtube-windows-amd64.exe`.
3. FX Tube will automatically open in lightweight app mode using Microsoft Edge or Chrome.

### macOS
```bash
chmod +x fxtube-darwin-arm64 # Or fxtube-darwin-amd64 for Intel Macs
./fxtube-darwin-arm64
```

---

## ⌨️ Keyboard Shortcuts

| Shortcut | Action |
| :--- | :--- |
| `Space` | Play / Pause |
| `Arrow Right` | Seek forward 10 seconds |
| `Arrow Left` | Seek backward 10 seconds |
| `/` | Focus search bar |
| `Esc` | Close modal / Return to browse view |

---

## 📦 Building from Source

```bash
# Requires Go 1.22+
git clone https://github.com/<your-username>/fxtube.git
cd fxtube

# Build static binary
CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -o fxtube .
```

---

## 📄 License
MIT License.
