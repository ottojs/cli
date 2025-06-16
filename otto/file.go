package otto

import (
	"fmt"
	"os"
	"strings"
)

// Same as below but with stricter security
func DeleteLines(filename, prefix string) error {
	return DeleteLinesWithOptions(filename, prefix, false)
}

// Removes lines with the given prefix from a file
// If allowAbsolute is true, it allows absolute paths and home directory expansion
func DeleteLinesWithOptions(filename, prefix string, allowAbsolute bool) error {
	// Validate and sanitize the path
	safePath, err := SecurePathWithOptions(filename, allowAbsolute)
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
		// Use TrimSpace to handle indented lines
		if !strings.HasPrefix(strings.TrimSpace(line), prefix) {
			filtered = append(filtered, line)
		}
	}

	// Preserve file permissions if it exists
	info, err := os.Stat(safePath)
	mode := os.FileMode(0600)
	if err == nil {
		mode = info.Mode()
	}

	return os.WriteFile(safePath, []byte(strings.Join(filtered, OSNewLine)), mode)
}

func WriteFile(content []byte, destination string) error {
	return WriteFileWithOptions(content, destination, false)
}

// Writes content to a file
// If allowAbsolute is true, it allows absolute paths and home directory expansion
func WriteFileWithOptions(content []byte, destination string, allowAbsolute bool) error {
	// Validate and sanitize the path
	safePath, err := SecurePathWithOptions(destination, allowAbsolute)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Preserve file permissions if it exists
	info, err := os.Stat(safePath)
	mode := os.FileMode(0600)
	if err == nil {
		mode = info.Mode()
	}

	file, err1 := os.OpenFile(safePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
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

// Updates a line that starts with the given prefix in a file
// If no line with the prefix exists, returns an error
// If allowAbsolute is true, it allows absolute paths and home directory expansion
func UpdateLineWithPrefix(filename, prefix, newLine string, allowAbsolute bool) error {
	// Validate and sanitize the path
	safePath, err := SecurePathWithOptions(filename, allowAbsolute)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	contents, err := os.ReadFile(safePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(contents), OSNewLine)
	updated := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = newLine
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("no line found with prefix: %s", prefix)
	}

	// Preserve file permissions if it exists
	info, err := os.Stat(safePath)
	mode := os.FileMode(0600)
	if err == nil {
		mode = info.Mode()
	}

	newContent := strings.Join(lines, OSNewLine)
	return os.WriteFile(safePath, []byte(newContent), mode)
}
