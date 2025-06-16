//go:build linux

package otto

import "os"

const OperatingSystem = "Linux"
const OSNewLine = "\n"

func OSHomeDir() string {
	Log("Not implemented: OSHomeDir")
	return ""
}

func OSProgramPath(appname string) string {
	Log("Not implemented: OSProgramPath")
	return ""
}

func AdminCheck() bool {
	Log("Not implemented: AdminCheck")
	return false
}

func EnvVarSet(key, value string) error {
	Log("Not implemented: EnvVarSet")
	return nil
}

func EnvVarGet(key string) string {
	return os.Getenv(key)
}

func EnvVarDelete(key string) error {
	Log("Not implemented: EnvVarDelete")
	return nil
}
