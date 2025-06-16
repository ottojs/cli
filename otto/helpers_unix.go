//go:build darwin || linux

package otto

import "os"

// Returns the effective user ID (used for testing/debugging)
func GetEUID() int {
	return os.Geteuid()
}

// Returns the SUDO_USER environment variable (used for testing/debugging)
func GetSudoUser() string {
	return os.Getenv("SUDO_USER")
}
