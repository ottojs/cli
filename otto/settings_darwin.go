//go:build darwin

package otto

import (
	"os"
	"os/user"
	"strings"
)

const OperatingSystem = "macOS"
const OSNewLine = "\n"

// Available for your use
// fmt.Println(os.Getenv("SUDO_COMMAND"))
// fmt.Println(os.Getenv("SUDO_USER"))
// fmt.Println(os.Getenv("SUDO_UID"))
// fmt.Println(os.Getenv("SUDO_GID"))

func OSHomeDir() string {
	// Get the user
	currentUser, err := user.Current()
	if err != nil {
		Log("Error:", err)
		return ""
	}
	theusername := currentUser.Username
	if currentUser.Username == "root" {
		envSudoUser := os.Getenv("SUDO_USER")
		if envSudoUser != "" {
			theusername = envSudoUser
		}
	}
	// TODO: Compare return values
	// currentUser.HomeDir
	return "/Users/" + theusername
}

func OSProgramPath(appname string) string {
	homedir := OSHomeDir()
	pathlist := []string{
		homedir, "Library", "Application Support", appname,
	}
	return strings.Join(pathlist, string(os.PathSeparator))
}

func AdminCheck() bool {
	currentUser, err := user.Current()
	if err != nil {
		Log("Error:", err)
		return false
	}
	return currentUser.Username == "root"
}

func EnvVarSet(key, value string) error {
	Log("Not implemented: EnvVarSet")
	return nil
}

func EnvVarGet(key string) string {
	return os.Getenv(key)
	return ""
}

func EnvVarDelete(key string) error {
	Log("Not implemented: EnvVarDelete")
	return nil
}
