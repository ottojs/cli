//go:build darwin

package otto

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const OperatingSystem = "macOS"
const OSNewLine = "\n"

// Available for your use
// fmt.Println(os.Getenv("SUDO_COMMAND"))
// fmt.Println(os.Getenv("SUDO_USER"))
// fmt.Println(os.Getenv("SUDO_UID"))
// fmt.Println(os.Getenv("SUDO_GID"))

func OSHomeDir() string {
	// First try HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		Log("Error:", err)
		return ""
	}

	// If running as root with sudo, try to get the actual user's home
	if currentUser.Username == "root" {
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
			if u, err := user.Lookup(sudoUser); err == nil {
				return u.HomeDir
			}
		}
	}

	return currentUser.HomeDir
}

func OSProgramPath(appname string) string {
	homedir := OSHomeDir()
	return filepath.Join(homedir, "Library", "Application Support", appname)
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
		content = strings.Join(marker, OSNewLine) + OSNewLine + newExportLine + OSNewLine
	} else {
		// Just add the export after existing Otto variables
		content = newExportLine + OSNewLine
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
