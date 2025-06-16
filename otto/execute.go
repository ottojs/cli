package otto

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"
)

// Comprehensive set of forbidden characters including control characters
// const forbiddenCharactersStrict string = "|&;<>()$`*?[]{}'\"\\\n\r\t!#^~"
const forbiddenCharacters string = "|&;<>()$`*?[]{}'\"\\\n\r\t"

// Returns stdout, stderr, and error/nil
func ExecuteBinary(binary string, args ...string) (stdout, stderr string, err error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sanitize binary path
	if binary == "" {
		return "", "", fmt.Errorf("binary path cannot be empty")
	}
	binary = strings.TrimSpace(binary)

	// Validate binary path for forbidden characters and control characters
	if containsForbiddenCharacters(binary) {
		return "", "", fmt.Errorf("binary path contains forbidden characters")
	}

	// Validate all arguments
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
		if containsForbiddenCharacters(arg) {
			return "", "", fmt.Errorf("argument %d contains forbidden characters", i+1)
		}
	}

	// Execute based on our platform
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// On Windows, use exec.LookPath to find the binary safely
		// Never use cmd.exe /C as it enables command chaining
		binPath, lookupErr := exec.LookPath(binary)
		if lookupErr != nil {
			// If not found in PATH, check if it's an absolute path
			if strings.Contains(binary, "\\") || strings.Contains(binary, "/") {
				binPath = binary
			} else {
				return "", "", fmt.Errorf("binary not found in PATH: %v", lookupErr)
			}
		}
		cmd = exec.CommandContext(ctx, binPath, args...)
	case "linux", "darwin":
		// Not absolute path to binary? Look it up...
		if !strings.HasPrefix(binary, "/") {
			binAbs, lookupErr := exec.LookPath(binary)
			if lookupErr != nil {
				return "", "", fmt.Errorf("binary not found in PATH: %v", lookupErr)
			}
			binary = binAbs
		}
		cmd = exec.CommandContext(ctx, binary, args...)
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Buckets for output
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run command
	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		return stdout, stderr, fmt.Errorf("command execution failed: %v", err)
	}

	return stdout, stderr, nil
}

// Checks for forbidden characters including control characters
func containsForbiddenCharacters(s string) bool {
	// Check for explicitly forbidden characters
	if strings.ContainsAny(s, forbiddenCharacters) {
		return true
	}

	// Check for any control characters (0x00-0x1F, 0x7F)
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}

	return false
}
