//go:build linux

package otto

import (
	"os"
	"os/user"
	"path/filepath"
)

const OperatingSystem = "Linux"
const OSNewLine = "\n"

func OSHomeDir() string {
	// First try HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Fallback to current user's home directory
	currentUser, err := user.Current()
	if err != nil {
		Log("Error getting home directory:", err)
		return ""
	}
	return currentUser.HomeDir
}

func OSProgramPath(appname string) string {
	homedir := OSHomeDir()
	if homedir == "" {
		return ""
	}

	// Use XDG Base Directory specification for Linux
	// First check XDG_CONFIG_HOME
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, appname)
	}

	// Default to ~/.config/appname
	return filepath.Join(homedir, ".config", appname)
}

func AdminCheck() bool {
	// Method 1: Check if we're running as root (UID 0)
	if os.Geteuid() == 0 {
		return true
	}

	// Method 2: Check if we're running with sudo
	// When running with sudo, SUDO_USER is set
	if os.Getenv("SUDO_USER") != "" {
		return true
	}

	// Method 3: Check username as fallback
	currentUser, err := user.Current()
	if err != nil {
		Log("Error checking user:", err)
		return false
	}

	return currentUser.Username == "root"
}

func EnvVarSet(key, value string) error {
	Log("Not implemented: EnvVarSet")
	return nil
}

func EnvVarGet(key string) string {
	return os.Getenv(key)
}

func EnvVarDelete(key string) error {
	Log("Not implemented: EnvVarDelete")
	return nil
}
