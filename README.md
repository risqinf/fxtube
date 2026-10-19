# FX Tube

[![Release](https://img.shields.io/github/v/release/risqinf/fxtube?style=flat-square)](https://github.com/risqinf/fxtube/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-informational?style=flat-square)](#)

[**[English]**](#english-documentation) | [**[Bahasa Indonesia]**](#dokumentasi-bahasa-indonesia)

---

<a name="english-documentation"></a>
## English Documentation

### What is FX Tube? (Simple Explanation)
Watching YouTube in modern web browsers (such as Chrome or Edge) can be frustrating on older or lower-spec computers. Modern YouTube loads hundreds of heavy background scripts, tracking services, and unskippable advertisements that cause high CPU usage, loud cooling fans, rapid battery drain, and video stuttering.

**FX Tube** is a lightweight desktop app that lets you watch YouTube smoothly without the browser bloat:
- **Zero Ads**: Streams video directly from YouTube's media servers, completely bypassing video ads, banners, and tracking scripts.
- **Ultra Lightweight**: Uses a fraction of the CPU and RAM compared to opening YouTube in a standard browser tab.
- **Keeps Your Account**: Syncs subscriptions, history, and playlists from your existing browser (Firefox, Chrome, Brave, Edge) without needing to re-enter your password.
- **Built for Any Hardware**: Runs smoothly on modern systems as well as 15-year-old laptops (like ThinkPad X201 or Core 2 Duo PCs).

---

### What Do You Need to Install?

#### Requirements:
You only need two lightweight helper tools installed on your system:
1. **yt-dlp** &mdash; Used by FX Tube to fetch video details and stream links.
2. **ffmpeg** &mdash; Used to merge the separate video and audio streams seamlessly.
3. *(Optional)* **mpv** &mdash; Only needed if you want to use the optional native MPV player mode for ultra-low (~40 MB) RAM usage.

#### What You DO NOT Need to Install:
- **No new browser needed**: FX Tube automatically uses the browser already built into your operating system (Microsoft Edge on Windows, Chrome/Safari on macOS, Firefox/Chromium on Linux).
- **No runtimes needed**: No need to install Node.js, Python, Java, or heavy dependencies. FX Tube is distributed as a single standalone executable.

---

### Installation Guide

#### Windows
1. Download `fxtube-windows-amd64.zip` from [Releases](https://github.com/risqinf/fxtube/releases).
2. Extract the ZIP archive.
3. Ensure `yt-dlp.exe` and `ffmpeg.exe` are in your system PATH (or placed in the same folder as `fxtube.exe`).
   * *Tip:* You can install both via winget in PowerShell:
     ```powershell
     winget install yt-dlp Gyan.FFmpeg
     ```
4. Double-click `fxtube-windows-amd64.exe` to run.

#### Linux
1. Install `yt-dlp` and `ffmpeg` using your package manager:
   * **Arch / Manjaro**: `sudo pacman -S yt-dlp ffmpeg`
   * **Ubuntu / Debian**: `sudo apt install yt-dlp ffmpeg`
   * **Fedora**: `sudo dnf install yt-dlp ffmpeg`
2. Download `fxtube-linux-amd64.tar.gz` (or `fxtube-linux-arm64.tar.gz` for Raspberry Pi) from [Releases](https://github.com/risqinf/fxtube/releases).
3. Extract and run:
   ```bash
   tar -xvf fxtube-linux-amd64.tar.gz
   chmod +x fxtube-linux-amd64
   ./fxtube-linux-amd64
   ```

#### macOS (Apple Silicon M1/M2/M3/M4)
1. Install `yt-dlp` and `ffmpeg` via Homebrew:
   ```bash
   brew install yt-dlp ffmpeg
   ```
2. Download and extract `fxtube-darwin-arm64.tar.gz` from [Releases](https://github.com/risqinf/fxtube/releases).
3. Make executable and run:
   ```bash
   chmod +x fxtube-darwin-arm64
   ./fxtube-darwin-arm64
   ```

---

### Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| `Space` | Play / Pause |
| `Right Arrow` | Skip forward 10 seconds |
| `Left Arrow` | Skip backward 10 seconds |
| `/` | Focus search bar |
| `Esc` | Close modal / Back to browse view |

---
---

<a name="dokumentasi-bahasa-indonesia"></a>
## Dokumentasi Bahasa Indonesia

### Apa itu FX Tube? (Penjelasan untuk Orang Awam)
Pernahkah laptop Anda mendadak lemot, kipasnya berputar kencang dan bising, atau baterainya cepat habis saat membuka YouTube di Google Chrome atau browser biasa? 

Hal itu wajar, karena situs web YouTube modern saat ini sangat berat. Setiap kali Anda membuka YouTube di browser, komputer Anda dipaksa memproses ratusan kode pelacak, iklan video yang tidak bisa di-skip, dan animasi yang memakan banyak memori RAM.

**FX Tube** hadir sebagai aplikasi pemutar YouTube di desktop yang dirancang khusus agar **super ringan, bersih, dan hemat baterai**:
- **Bebas Iklan 100%**: Langsung memutar video dari server media YouTube tanpa memuat iklan banner maupun video sponsor.
- **Sangat Hemat RAM & CPU**: Menggunakan sumber daya komputer yang jauh lebih sedikit dibanding membuka YouTube di tab browser biasa.
- **Tetap Bisa Pakai Akun Anda**: Langsung membaca daftar Subscribe, Riwayat, dan Playlist dari browser yang biasa Anda pakai (Firefox, Chrome, Brave, Edge) tanpa perlu mengetik ulang password.
- **Dukungan Komputer Lawas**: Mampu berjalan mulus di laptop lama (bahkan laptop berusia 10-15 tahun seperti ThinkPad X201 atau PC Core 2 Duo).

---

### Apa Saja yang Perlu Di-install dan yang TIDAK Perlu?

#### Yang PERLU Di-install (Hanya 2 Program Ringan):
Agar FX Tube bisa mengambil dan memutar video dari YouTube, sistem Anda membutuhkan dua alat bantu:
1. **yt-dlp** &mdash; Alat yang bertugas mencari video dan mengambil link stream dari YouTube.
2. **ffmpeg** &mdash; Alat yang bertugas menggabungkan stream gambar dan suara secara langsung.
3. *(Opsional)* **mpv** &mdash; Hanya dibutuhkan bila Anda ingin memakai mode pemutar MPV Native (opsi darurat yang sangat hemat RAM, hanya ~40 MB).

#### Yang TIDAK PERLU Di-install:
- **TIDAK PERLU install browser baru**: FX Tube otomatis mendeteksi dan memakai browser yang sudah ada di komputer Anda (Microsoft Edge di Windows, Chrome/Safari di Mac, atau Firefox di Linux).
- **TIDAK PERLU install Node.js, Python, atau database**: FX Tube berbentuk file program mandiri (*standalone executable*) yang langsung siap dijalankan.

---

### Panduan Pemasangan & Menjalankan

#### Pengguna Windows
1. Unduh file `fxtube-windows-amd64.zip` dari halaman [Releases](https://github.com/risqinf/fxtube/releases).
2. Ekstrak file ZIP tersebut.
3. Pastikan `yt-dlp` dan `ffmpeg` sudah terpasang. Cara termudah cukup buka PowerShell dan ketik:
   ```powershell
   winget install yt-dlp Gyan.FFmpeg
   ```
4. Klik dua kali pada file `fxtube-windows-amd64.exe` untuk langsung menggunakannya.

#### Pengguna Linux
1. Pasang `yt-dlp` dan `ffmpeg` melalui terminal:
   * **Arch / Manjaro**: `sudo pacman -S yt-dlp ffmpeg`
   * **Ubuntu / Debian**: `sudo apt install yt-dlp ffmpeg`
   * **Fedora**: `sudo dnf install yt-dlp ffmpeg`
2. Unduh `fxtube-linux-amd64.tar.gz` dari halaman [Releases](https://github.com/risqinf/fxtube/releases).
3. Ekstrak dan jalankan:
   ```bash
   tar -xvf fxtube-linux-amd64.tar.gz
   chmod +x fxtube-linux-amd64
   ./fxtube-linux-amd64
   ```

#### Pengguna macOS (Apple Silicon M1/M2/M3/M4)
1. Pasang `yt-dlp` dan `ffmpeg` via Homebrew:
   ```bash
   brew install yt-dlp ffmpeg
   ```
2. Unduh dan ekstrak `fxtube-darwin-arm64.tar.gz` dari [Releases](https://github.com/risqinf/fxtube/releases).
3. Beri izin eksekusi lalu jalankan:
   ```bash
   chmod +x fxtube-darwin-arm64
   ./fxtube-darwin-arm64
   ```

---

### Tombol Pintasan Keyboard

| Tombol | Fungsi |
| :--- | :--- |
| `Spasi` | Putar / Jeda video (Play / Pause) |
| `Panah Kanan` | Lompat maju 10 detik (+10s) |
| `Panah Kiri` | Lompat mundur 10 detik (-10s) |
| `/` | Sorot kolom pencarian video |
| `Esc` | Menutup jendela pengaturan / Kembali ke beranda |

---

## License
Didistribusikan di bawah lisensi terbuka [MIT License](LICENSE). Hak Cipta (c) 2026 Risqi Nur Fadilah.
