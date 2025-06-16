//go:build windows

package otto

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const OperatingSystem = "Windows"
const OSNewLine = "\r\n"

func OSHomeDir() string {
	currentUser, err := user.Current()
	if err != nil {
		LogError("Error getting current user: %v", err)
		return ""
	}
	// currentUser.Username
	return currentUser.HomeDir
}

func OSProgramPath(appname string) string {
	homedir := OSHomeDir()
	return filepath.Join(homedir, "AppData", "Local", appname)
}

// Windows API constants for admin check
const (
	SECURITY_BUILTIN_DOMAIN_RID = 0x00000020
	DOMAIN_ALIAS_RID_ADMINS     = 0x00000220
)

// Checks if the current process is running with administrator privileges
// Uses Windows APIs for reliable detection
func AdminCheck() bool {
	// Method 1: Try using CheckTokenMembership API (most reliable)
	if isAdmin, err := checkTokenMembership(); err == nil {
		return isAdmin
	}

	// Method 2: Fallback to checking if we can write to a protected location
	return canWriteToProtectedLocation()
}

// Uses Windows APIs to check admin status
func checkTokenMembership() (bool, error) {
	// This is a simplified check - we'll use a different approach
	// The syscall package doesn't expose all the needed functions
	// So we'll use a more direct approach

	// Get current process token
	var token syscall.Token
	proc, err := syscall.GetCurrentProcess()
	if err != nil {
		return false, err
	}

	err = syscall.OpenProcessToken(proc, syscall.TOKEN_QUERY, &token)
	if err != nil {
		return false, err
	}
	defer token.Close()

	// Check if token has elevated privileges (simpler approach)
	var elevation uint32
	var returnedLen uint32
	err = syscall.GetTokenInformation(
		token,
		syscall.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&returnedLen,
	)
	if err != nil {
		return false, err
	}

	return elevation != 0, nil
}

// Tests if we can write to Windows directory
func canWriteToProtectedLocation() bool {
	// Try to create a temporary file in Windows\Temp
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		winDir = "C:\\Windows"
	}

	testPath := filepath.Join(winDir, "Temp", "otto_admin_test.tmp")

	// Try to create the file
	file, err := os.Create(testPath)
	if err != nil {
		return false
	}

	// Clean up
	file.Close()
	os.Remove(testPath)

	return true
}

// Validates environment variable names
func ValidateEnvVarName(key string) error {
	if strings.ContainsAny(key, " \t\n=") {
		return fmt.Errorf("invalid environment variable name: %s", key)
	}
	return nil
}

// https://stackoverflow.com/questions/13222724/command-line-to-remove-an-environment-variable-from-the-os-level-configuration
func EnvVarSet(key, value string) error {
	// Validate key
	if err := ValidateEnvVarName(key); err != nil {
		return err
	}

	// Set for current process
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("failed to set environment variable: %w", err)
	}
	// User
	// setx MYVAR
	//
	// System (Avoid)
	// setx MYVAR "" /M

	// Open the user environment registry key
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	// Set the environment variable in the registry
	if err := k.SetStringValue(key, value); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	// Broadcast WM_SETTINGCHANGE to notify other applications
	broadcastEnvironmentChange()

	return nil
}

func EnvVarGet(key string) string {
	// First check current process environment
	if value := os.Getenv(key); value != "" {
		return value
	}

	// echo %MYVAR%
	// Check registry for persistent value
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	value, _, err := k.GetStringValue(key)
	if err != nil {
		return ""
	}
	return value
}

func EnvVarDelete(key string) error {
	// User
	// REG DELETE HKCU\Environment /f /v MYVAR
	//
	// System (Avoid)
	// REG DELETE "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /f /v MYVAR

	// Unset for current process
	if err := os.Unsetenv(key); err != nil {
		return fmt.Errorf("failed to unset environment variable: %w", err)
	}

	// Open the user environment registry key
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	// Delete the environment variable from the registry
	if err := k.DeleteValue(key); err != nil {
		// If the value doesn't exist, that's okay
		if err != registry.ErrNotExist {
			return fmt.Errorf("failed to delete registry value: %w", err)
		}
	}

	// Broadcast WM_SETTINGCHANGE to notify other applications
	broadcastEnvironmentChange()

	return nil
}

// Notifies Windows applications that environment variables have changed
// Probably not vital but adding for completeness. Don't like that it uses unsafe
func broadcastEnvironmentChange() {
	// Load the Windows DLLs
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	// Windows constants
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)

	// Convert "Environment" to UTF-16 for Windows API
	env, _ := syscall.UTF16PtrFromString("Environment")

	// Send the broadcast message
	sendMessageTimeout.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(SMTO_ABORTIFHUNG),
		5000, // 5 second timeout
		0,
	)
}
