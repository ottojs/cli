//go:build linux

package otto

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	// Validate key - no spaces or special characters
	if strings.ContainsAny(key, " \t\n=") {
		return fmt.Errorf("invalid environment variable name: %s", key)
	}

	// Set for current process
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}

	// Get bash profile file
	homeDir := OSHomeDir()
	if homeDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Use .bashrc for Linux (most common)
	profileFile := filepath.Join(homeDir, ".bashrc")

	// First, check if we need to update an existing export
	if err := updateExistingEnvVar(profileFile, key, value); err == nil {
		return nil // Successfully updated existing var
	}

	// If not found, append new export
	exportLine := fmt.Sprintf("export %s=\"%s\"", key, value)

	// Add a marker comment for organization
	marker := []string{
		"",
		"# Environment variables added by Otto CLI",
	}

	// Check if our marker already exists
	markerExists, err := FindChunkInFileWithOptions(profileFile, marker[1:], true) // Skip empty line for search, allow absolute paths
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check .bashrc: %w", err)
	}

	// Prepare content to append
	var content string
	if !markerExists {
		// Add marker and export
		content = strings.Join(marker, OSNewLine) + OSNewLine + exportLine + OSNewLine
	} else {
		// Just add the export after existing Otto variables
		content = exportLine + OSNewLine
	}

	// Append to file with backup
	if err := AppendToFileWithBackupAndOptions(profileFile, content, ".bak", true); err != nil {
		return fmt.Errorf("failed to update .bashrc: %w", err)
	}

	return nil
}

func EnvVarGet(key string) string {
	return os.Getenv(key)
}

func EnvVarDelete(key string) error {
	// Unset for current process
	if err := os.Unsetenv(key); err != nil {
		return fmt.Errorf("failed to unset environment variable: %w", err)
	}

	// Get bash profile file
	homeDir := OSHomeDir()
	if homeDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Use .bashrc for Linux
	profileFile := filepath.Join(homeDir, ".bashrc")

	// Delete lines that export this variable
	exportPrefix := fmt.Sprintf("export %s=", key)
	if err := DeleteLinesWithOptions(profileFile, exportPrefix, true); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to update .bashrc: %w", err)
	}

	return nil
}

// Updates an existing environment variable export in the file
func updateExistingEnvVar(filepath, key, value string) error {
	// Read file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err // Not found or can't read
	}

	lines := strings.Split(string(content), OSNewLine)
	exportPrefix := fmt.Sprintf("export %s=", key)
	newExportLine := fmt.Sprintf("export %s=\"%s\"", key, value)
	updated := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, exportPrefix) {
			lines[i] = newExportLine
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("environment variable %s not found", key)
	}

	// Write back the updated content
	newContent := strings.Join(lines, OSNewLine)
	return WriteFileWithOptions([]byte(newContent), filepath, true)
}
