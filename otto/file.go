package otto

import (
	"fmt"
	"os"
	"strings"
)

func DeleteLines(filename, prefix string) error {
	contents, err1 := os.ReadFile(filename)
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

	return os.WriteFile(filename, []byte(strings.Join(filtered, OSNewLine)), 0644)
}

func WriteFile(content []byte, destination string) error {
	file, err1 := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0644)
	if err1 != nil {
		return fmt.Errorf("failed to open file: %v", err1)
	}
	defer file.Close()
	if _, err2 := file.Write(content); err2 != nil {
		return fmt.Errorf("failed to append content: %v", err2)
	}
	return nil
}

func EnsureChunkInFile(path string, content []string) error {
	// Check if needed
	found, err1 := FindChunkInFile(path, content)
	if err1 != nil {
		return err1
	}
	// Skip if found
	if found {
		return nil
	}
	// Append to test file
	err2 := AppendToFileWithBackup(path, strings.Join(content, OSNewLine), "")
	if err2 != nil {
		return err2
	}
	return nil
}
