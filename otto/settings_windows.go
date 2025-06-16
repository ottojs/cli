//go:build windows

package otto

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const OperatingSystem = "Windows"
const OSNewLine = "\r\n"

func OSHomeDir() string {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	// currentUser.Username
	return currentUser.HomeDir
}

func OSProgramPath(appname string) string {
	homedir := OSHomeDir()
	return filepath.Join(homedir, "AppData", "Local", appname)
}

func AdminCheck() bool {
	filehandle, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	filehandle.Close()
	return true
}

// https://stackoverflow.com/questions/13222724/command-line-to-remove-an-environment-variable-from-the-os-level-configuration
func EnvVarSet(key, value string) error {
	// User
	// setx MYVAR
	//
	// System (Avoid)
	// setx MYVAR "" /M
	args := []string{
		strings.ToUpper(key),
		value,
	}
	_, _, err := ExecuteBinary("setx", args...)
	if err != nil {
		// fmt.Printf("Error: %v"+OSNewLine, err)
		// fmt.Printf("Stderr: %s"+OSNewLine, stderr)
		return errors.New("error setting environment variable")
	}
	return nil
}

func EnvVarGet(key string) string {
	// echo %MYVAR%
	args := []string{
		"%" + strings.ToUpper(key) + "%",
	}
	stdout, _, err := ExecuteBinary("echo", args...)
	if err != nil {
		// fmt.Printf("Error: %v"+OSNewLine, err)
		// fmt.Printf("Stderr: %s"+OSNewLine, stderr)
		return ""
	}
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "%"+key+"%" {
		trimmed = ""
	}
	return trimmed
}

func EnvVarDelete(key string) error {
	// User
	// REG DELETE HKCU\Environment /f /v MYVAR
	//
	// System (Avoid)
	// REG DELETE "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /f /v MYVAR
	args := []string{
		"DELETE",
		"HKCU\\Environment",
		"/F",
		"/V",
		strings.ToUpper(key),
	}
	_, _, err := ExecuteBinary("REG", args...)
	if err != nil {
		// fmt.Printf("Error: %v"+OSNewLine, err)
		// fmt.Printf("Stderr: %s"+OSNewLine, stderr)
		return errors.New("error deleting environment variable")
	}
	return nil
}
