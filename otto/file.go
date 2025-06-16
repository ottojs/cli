package otto

import (
	"fmt"
	"os"
	"strings"
)

func DeleteLines(filename, prefix string) error {
	// Validate and sanitize the path
	safePath, err := SecurePath(filename)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	contents, err1 := os.ReadFile(safePath)
	if err1 != nil {
		return err1
	}

	lines := strings.Split(string(contents), OSNewLine)
	filtered := []string{}

	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			filtered = append(filtered, line)
		}
	}

	return os.WriteFile(safePath, []byte(strings.Join(filtered, OSNewLine)), 0600)
}

func WriteFile(content []byte, destination string) error {
	// Validate and sanitize the path
	safePath, err := SecurePath(destination)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	file, err1 := os.OpenFile(safePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err1 != nil {
		return fmt.Errorf("failed to open file: %w", err1)
	}
	defer file.Close()
	if _, err2 := file.Write(content); err2 != nil {
		return fmt.Errorf("failed to write content: %w", err2)
	}
	return nil
}

func EnsureChunkInFile(path string, content []string) error {
	// Validate and sanitize the path
	safePath, err := SecurePath(path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if needed
	found, err1 := FindChunkInFile(safePath, content)
	if err1 != nil {
		return err1
	}
	// Skip if found
	if found {
		return nil
	}
	// Append to test file
	err2 := AppendToFileWithBackup(safePath, strings.Join(content, OSNewLine), "")
	if err2 != nil {
		return err2
	}
	return nil
}
