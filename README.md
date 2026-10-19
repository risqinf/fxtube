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
- **Dynamic Resolutions (144p - 1080p)**: Automatically detects and unlocks all available video resolutions up to 1080p Full HD.
- **Smooth Playlists & Infinite Feed**: Next videos preload seamlessly in the background, and the home feed supports live refresh and infinite scroll.
- **Keeps Your Account**: Syncs subscriptions, history, and playlists from your existing browser (Firefox, Chrome, Brave, Edge) without needing to re-enter your password.
- **Zero Configuration**: Bundled with helpers or auto-downloads dependencies on startup.

---

### Download & Installation Packages

Choose the installer or bundle matching your operating system from [GitHub Releases](https://github.com/risqinf/fxtube/releases):

#### Windows
- **Setup Wizard (.exe installer)**: Download `fxtube-windows-amd64-installer.exe` and run. It installs FX Tube to Program Files, creates Start Menu and Desktop shortcuts, and bundles all dependencies.
- **MSI Installer (.msi)**: Download `fxtube-windows-amd64.msi` for automated deployment.
- **Standalone Portable (.zip)**: Download `fxtube-windows-amd64.zip`. Extract anywhere and double-click `fxtube.exe`. `yt-dlp` and `ffmpeg` are already bundled inside!

#### Linux (Debian, Ubuntu, Fedora, Arch Linux)
- **Ubuntu / Debian / Linux Mint (.deb)**:
  ```bash
  sudo apt install ./fxtube_*_amd64.deb
  ```
  *(Automatically registers Start Menu application icon and resolves dependencies)*
- **Fedora / openSUSE / RHEL (.rpm)**:
  ```bash
  sudo dnf install ./fxtube-*.x86_64.rpm
  ```
- **Arch Linux / Manjaro / EndeavourOS (.pkg.tar.zst)**:
  ```bash
  sudo pacman -U ./fxtube-*-x86_64.pkg.tar.zst
  ```
- **Generic Tarball (.tar.gz)**:
  ```bash
  tar -xvf fxtube-linux-amd64.tar.gz
  ./fxtube
  ```

#### macOS (Apple Silicon M1/M2/M3/M4)
1. Download `fxtube-darwin-arm64.tar.gz` from [Releases](https://github.com/risqinf/fxtube/releases).
2. Extract and run:
   ```bash
   tar -xvf fxtube-darwin-arm64.tar.gz
   ./fxtube
   ```

#### Android (Mobile Phones & Tablets - Android 4.0+ / 1GB RAM)
FX Tube is ultra-optimized for legacy smartphones (Android 4.0 Ice Cream Sandwich up to Android 14):
- **Wi-Fi LAN Access (Zero Install)**: Connect your phone to the same Wi-Fi as your computer running FX Tube and open `http://<computer-ip>:18201` in your phone's browser (Chrome, stock WebKit, or Opera). Consumes only ~25 MB RAM!
- **Native Android APK (`android/`)**: Project source included with `minSdkVersion 14` for building an ultra-lightweight standalone APK wrapper.
- **Full YouTube Shorts**: Dedicated Shorts tab with 9:16 vertical layout and auto-looping.

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
- **Resolusi Lengkap Dinamis (144p hingga 1080p)**: Otomatis mendeteksi resolusi video dan menyediakan pilihan dari 144p (hemat kuota) hingga 1080p Full HD.
- **Playlist Cepat & Mulus**: Video selanjutnya di playlist sudah di-preload secara otomatis di latar belakang sehingga pergantian video instan dan tidak macet.
- **Beranda Segar & Scroll Tanpa Batas**: Tombol "Muat Ulang" langsung menyegarkan rekomendasi terkini dari YouTube, dan daftar video otomatis memuat ke bawah (infinite scroll).
- **Sangat Hemat RAM & CPU**: Menggunakan sumber daya komputer yang jauh lebih sedikit dibanding membuka YouTube di tab browser biasa.
- **Tetap Bisa Pakai Akun Anda**: Langsung membaca daftar Subscribe, Riwayat, dan Playlist dari browser Anda (Firefox, Chrome, Brave, Edge) tanpa perlu login ulang.
- **Siap Pakai (Zero Setup)**: Bundling biner sudah lengkap atau otomatis mengunduh alat bantu saat pertama kali dibuka jika belum ada.

---

### Pilihan Installer & Paket Aplikasi Siap Pakai

Unduh berkas instalasi sesuai sistem operasi Anda di halaman [GitHub Releases](https://github.com/risqinf/fxtube/releases):

#### Pengguna Windows
- **Installer Wizard (.exe)**: Unduh `fxtube-windows-amd64-installer.exe` lalu klik dua kali. Program akan terpasang rapi di Start Menu dan Desktop Anda, lengkap dengan alat bantu di dalamnya.
- **Installer MSI (.msi)**: Unduh `fxtube-windows-amd64.msi` untuk instalasi standar Windows.
- **Versi Portabel (.zip)**: Unduh `fxtube-windows-amd64.zip`, ekstrak ke folder mana saja, lalu klik dua kali `fxtube.exe`. Semua dependensi sudah ada di dalam satu folder.

#### Pengguna Linux (Ubuntu, Debian, Fedora, Arch Linux)
- **Ubuntu / Debian / Linux Mint (.deb)**:
  Cukup unduh file `.deb` lalu pasang dengan:
  ```bash
  sudo apt install ./fxtube_*_amd64.deb
  ```
  *(Ikon aplikasi otomatis muncul di menu aplikasi/desktop Anda)*
- **Fedora / openSUSE / RHEL (.rpm)**:
  ```bash
  sudo dnf install ./fxtube-*.x86_64.rpm
  ```
- **Arch Linux / Manjaro / EndeavourOS (.pkg.tar.zst)**:
  ```bash
  sudo pacman -U ./fxtube-*-x86_64.pkg.tar.zst
  ```
- **Tarball Portabel (.tar.gz)**:
  ```bash
  tar -xvf fxtube-linux-amd64.tar.gz
  ./fxtube
  ```

#### Pengguna macOS (Apple Silicon M1/M2/M3/M4)
1. Unduh berkas `fxtube-darwin-arm64.tar.gz` dari [Releases](https://github.com/risqinf/fxtube/releases).
2. Ekstrak dan jalankan:
   ```bash
   tar -xvf fxtube-darwin-arm64.tar.gz
   ./fxtube
   ```

#### Pengguna HP Android (Android 4.0+ & RAM 1GB)
FX Tube dirancang sangat ringan sehingga dapat berjalan mulus di HP Android jadul (mulai dari Android 4.0 Ice Cream Sandwich hingga Android modern):
- **Akses Langsung via Wi-Fi (Tanpa Install di HP)**: Hubungkan HP Anda ke Wi-Fi yang sama dengan komputer Anda, lalu buka browser di HP dan ketik `http://<ip-komputer>:18201` (alamat IP otomatis tertera di menu Setting FX Tube). Hanya memakan RAM ~25 MB!
- **Aplikasi Asli Android APK (`android/`)**: Tersedia proyek aplikasi Android native dengan `minSdkVersion 14` yang hemat RAM, mendukung layar penuh, dan akselerasi grafis perangkat keras lama.
- **Mendukung Penuh YouTube Shorts**: Tab khusus Shorts dengan pemutar video vertikal 9:16 yang otomatis memutar berulang (looping) tanpa macet.

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
