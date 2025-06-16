package otto

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func PromptNormal() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Value: ")
	value, err := reader.ReadString('\n')
	return value, err
}

func PromptSensitive() (string, error) {
	byteValue, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	password := string(byteValue)
	return strings.TrimSpace(password), nil
}
