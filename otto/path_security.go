package otto

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// SecurePath validates and sanitizes a file path to prevent directory traversal attacks.
// It returns an error if the path attempts to escape the current working directory
// or contains other security risks.
func SecurePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	// Check for null bytes early
	if strings.Contains(path, "\x00") {
		return "", errors.New("path contains null bytes")
	}

	// Clean the path to resolve . and .. elements
	cleaned := filepath.Clean(path)

	// Check if it's an absolute path on any OS
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	// Additional check for Windows absolute paths that might not be caught
	if len(cleaned) >= 2 && cleaned[1] == ':' {
		return "", fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	// Get absolute path for validation
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Get current working directory
	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ensure the path is within or relative to the current working directory
	// This prevents directory traversal attacks
	if !strings.HasPrefix(absPath, cwd) {
		return "", fmt.Errorf("path traversal detected: %s is outside working directory", path)
	}

	// Additional security checks
	// Check for suspicious patterns that might bypass security
	suspiciousPatterns := []string{
		"..", // Parent directory references (should be caught by Clean but double-check)
		"~",  // Home directory expansion
		"$",  // Environment variable expansion
		"%",  // Windows environment variable expansion
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, pattern) {
			// Only block if these appear in ways that could be dangerous
			if pattern == ".." && strings.Contains(path, ".."+string(filepath.Separator)) {
				return "", fmt.Errorf("path contains dangerous pattern: %s", pattern)
			}
			if pattern == "~" && strings.HasPrefix(path, "~") {
				return "", fmt.Errorf("path contains dangerous pattern: %s", pattern)
			}
		}
	}

	// Return the cleaned, safe path
	return cleaned, nil
}

// SecureJoinPath safely joins path elements and validates the result.
// It ensures the resulting path doesn't escape the base directory.
func SecureJoinPath(base string, elements ...string) (string, error) {
	// Start with the base path
	result := base

	// Join each element
	for _, elem := range elements {
		// Clean each element
		elem = filepath.Clean(elem)

		// Reject elements that try to go up directories
		if strings.Contains(elem, "..") {
			return "", fmt.Errorf("path element contains '..': %s", elem)
		}

		result = filepath.Join(result, elem)
	}

	// Validate the final path
	return SecurePath(result)
}
