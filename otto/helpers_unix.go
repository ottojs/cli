//go:build darwin || linux

package otto

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// Returns the effective user ID (used for testing/debugging)
func GetEUID() int {
	return os.Geteuid()
}

// Returns the SUDO_USER environment variable (used for testing/debugging)
func GetSudoUser() string {
	return os.Getenv("SUDO_USER")
}

// Determines the appropriate shell profile file for the current system
// If no preferences are provided, it uses platform defaults (zsh for macOS, bash for Linux)
func GetShellProfileFile(preferredShells ...string) string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// Fallback to current user's home
		if u, err := os.UserHomeDir(); err == nil {
			homeDir = u
		} else {
			return ""
		}
	}

	// Define shell profile mappings
	shellProfiles := map[string][]string{
		"zsh":  {".zshrc", ".zprofile"},
		"bash": {".bashrc", ".bash_profile"},
		"sh":   {".profile"},
	}

	// Set default preferences based on platform if none provided
	if len(preferredShells) == 0 {
		if runtime.GOOS == "darwin" {
			preferredShells = []string{"zsh", "bash", "sh"}
		} else {
			preferredShells = []string{"bash", "sh"}
		}
	}

	// Check each preferred shell in order
	for _, shell := range preferredShells {
		if profiles, ok := shellProfiles[shell]; ok {
			for _, profile := range profiles {
				profilePath := filepath.Join(homeDir, profile)
				if _, err := os.Stat(profilePath); err == nil {
					return profilePath
				}
			}
		}
	}

	// If nothing found, return default based on platform and first preference
	if len(preferredShells) > 0 {
		if profiles, ok := shellProfiles[preferredShells[0]]; ok && len(profiles) > 0 {
			return filepath.Join(homeDir, profiles[0])
		}
	}

	// Ultimate fallback
	if runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, ".zshrc")
	}
	return filepath.Join(homeDir, ".bashrc")
}

// Common Unix implementation for checking admin/root privileges
func UnixAdminCheck() bool {
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
		LogError("Error checking user: %v", err)
		return false
	}

	return currentUser.Username == "root"
}

// Validates environment variable names for Unix systems
func ValidateEnvVarName(key string) error {
	if strings.ContainsAny(key, " \t\n=") {
		return fmt.Errorf("invalid environment variable name: %s", key)
	}
	return nil
}
