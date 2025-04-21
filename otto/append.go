package otto

import (
	"fmt"
	"os"
)

func AppendToFileWithBackup(path, content, backup_ext string) error {
	// Validate input
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		// If we want a backup (not "")
		if backup_ext != "" {
			// Back it up if it exists
			if err := CopyFile(path, path+backup_ext); err != nil {
				return fmt.Errorf("failed to create backup: %v", err)
			}
		}
	} else if !os.IsNotExist(err) {
		// Unexpected error checking file
		return fmt.Errorf("failed to check file existence: %v", err)
	}

	// Open file in append mode (or create it)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Append content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to append content: %v", err)
	}

	return nil
}

func CopyFile(source, destination string) error {
	// Read
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	// Write
	if err := os.WriteFile(destination, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %v", err)
	}

	return nil
}
