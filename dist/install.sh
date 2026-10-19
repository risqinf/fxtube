#!/usr/bin/env bash
# FX Tube 1-Click Installer
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$HOME/.local/bin"
APP_DIR="$HOME/.local/share/applications"

mkdir -p "$BIN_DIR" "$APP_DIR"
cp "$DIR/fxtube" "$BIN_DIR/fxtube"
chmod +x "$BIN_DIR/fxtube"

cat <<EOF > "$APP_DIR/fxtube.desktop"
[Desktop Entry]
Name=FX Tube
GenericName=YouTube Client
Comment=Ultra Lightweight YouTube Client
Exec=$BIN_DIR/fxtube
Icon=video-display
Terminal=false
Type=Application
Categories=AudioVideo;Video;Player;
StartupWMClass=fxtube
EOF

update-desktop-database "$APP_DIR" 2>/dev/null || true

echo "✅ FX Tube berhasil dipasang!"
echo "👉 Jalankan dengan mengetik 'fxtube' atau cari 'FX Tube' di menu aplikasi."
