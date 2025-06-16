package otto

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func FindChunkInFile(path string, chunk []string) (bool, error) {
	// Validate and sanitize the path
	safePath, err := SecurePath(path)
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
