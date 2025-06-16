package otto

import (
	"fmt"
	"os"
	"path/filepath"
)

func AppendToFileWithBackup(path, content, backup_ext string) error {
	// Validate input
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Validate and sanitize the main path
	safePath, err := SecurePath(path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(safePath); err == nil {
		// If we want a backup (not "")
		if backup_ext != "" {
			// Validate backup extension doesn't contain path separators
			if filepath.IsAbs(backup_ext) || filepath.Dir(backup_ext) != "." {
				return fmt.Errorf("invalid backup extension: must not contain path separators")
			}

			// Create backup path
			backupPath := safePath + backup_ext

			// Validate the backup path is still safe
			safeBackupPath, err := SecurePath(backupPath)
			if err != nil {
				return fmt.Errorf("invalid backup path: %w", err)
			}

			// Back it up if it exists
			if err := CopyFile(safePath, safeBackupPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		// Unexpected error checking file
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Open file in append mode (or create it) with secure permissions
	file, err := os.OpenFile(safePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Append content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to append content: %w", err)
	}

	return nil
}

func CopyFile(source, destination string) error {
	// Validate and sanitize source path
	safeSource, err := SecurePath(source)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	// Validate and sanitize destination path
	safeDest, err := SecurePath(destination)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Read
	data, err := os.ReadFile(safeSource)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write with secure permissions
	if err := os.WriteFile(safeDest, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
