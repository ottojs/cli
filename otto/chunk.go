package otto

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Same as below but with stricter security
func FindChunkInFile(path string, chunk []string) (bool, error) {
	return FindChunkInFileWithOptions(path, chunk, false)
}

// Searches for a chunk of lines in a file
// If allowAbsolute is true, it allows absolute paths and home directory expansion
func FindChunkInFileWithOptions(path string, chunk []string, allowAbsolute bool) (bool, error) {
	// Validate and sanitize the path
	safePath, err := SecurePathWithOptions(path, allowAbsolute)
	if err != nil {
		return false, fmt.Errorf("invalid file path: %w", err)
	}

	// Open the file
	filehandle, err := os.Open(safePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file %s: %w", safePath, err)
	}
	defer filehandle.Close()

	// Use a sliding window
	var window []string
	scanner := bufio.NewScanner(filehandle)

	// Read lines into the window
	for scanner.Scan() {
		line := scanner.Text()
		// Add line to scanning window
		window = append(window, line)

		// Check for a match
		if len(window) == len(chunk) {
			// Compare
			matches := true
			for i := range chunk {
				// Trim pre/trailing whitespace
				if strings.TrimSpace(window[i]) != strings.TrimSpace(chunk[i]) {
					matches = false
					break
				}
			}
			if matches {
				return true, nil
			}
			// Remove first/oldest line
			// Probably not efficient but fine for our purposes
			window = window[1:]
		}
	}

	// Check for error
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading file %s: %v", path, err)
	}

	// Not found
	return false, nil
}
