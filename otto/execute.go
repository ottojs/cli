package otto

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const forbiddenCharacters string = "|&;<>()$`*?[]{}'\""
const forbiddenCharactersStrict string = "|&;<>()$`*?[]{}'\"\\"

// Returns stdout, stderr, and error/nil
func ExecuteBinary(binary string, args ...string) (stdout, stderr string, err error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sanitize Binary/Arguments (not guaranteed, but it's something)
	if binary == "" {
		return "", "", fmt.Errorf("binary path cannot be empty")
	}
	binary = strings.TrimSpace(binary)
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
		if strings.ContainsAny(arg, forbiddenCharacters) {
			return "", "", fmt.Errorf("argument contains forbidden characters")
		}
	}

	// Execute based on our platform
	cmd := exec.CommandContext(ctx, binary, args...)
	switch runtime.GOOS {
	case "windows":
		// On Windows, use cmd.exe to handle PATH resolution
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", binary)
		if len(args) > 0 {
			cmd.Args = append(cmd.Args, args...)
		}
	case "linux", "darwin":
		// Not absolute path to binary? Look it up...
		if !strings.HasPrefix(binary, "/") {
			var binAbs string
			binAbs, err = exec.LookPath(binary)
			if err != nil {
				return "", "", fmt.Errorf("binary not found in PATH: %v", err)
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
