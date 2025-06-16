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
		LogError("Error getting home directory: %v", err)
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
	return UnixAdminCheck()
}

func EnvVarSet(key, value string) error {
	// Validate key
	if err := ValidateEnvVarName(key); err != nil {
		return err
	}

	// Set for current process
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}

	// Get profile file
	profileFile := GetShellProfileFile()
	if profileFile == "" {
		return fmt.Errorf("could not determine shell profile file")
	}

	// First, check if we need to update an existing export
	exportPrefix := fmt.Sprintf("export %s=", key)
	newExportLine := fmt.Sprintf("export %s=\"%s\"", key, value)
	if err := UpdateLineWithPrefix(profileFile, exportPrefix, newExportLine, true); err == nil {
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
		return fmt.Errorf("failed to check shell profile: %w", err)
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
		return fmt.Errorf("failed to update shell profile: %w", err)
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

	// Get profile file
	profileFile := GetShellProfileFile()
	if profileFile == "" {
		return fmt.Errorf("could not determine shell profile file")
	}

	// Delete lines that export this variable
	exportPrefix := fmt.Sprintf("export %s=", key)
	if err := DeleteLinesWithOptions(profileFile, exportPrefix, true); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to update shell profile: %w", err)
	}

	return nil
}
