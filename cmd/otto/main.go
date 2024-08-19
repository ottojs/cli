package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"code.ottojs.org/go/otto"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func main() {
	app := &cli.App{
		Name:    "otto",
		Version: "0.0.1",
		Usage:   "Otto Tools CLI",
		Authors: []*cli.Author{
			{
				Name:  "Otto.js",
				Email: "help@ottojs.org",
			},
		},
		Copyright: "(c) 2024 Otto.js",
		Commands: []*cli.Command{
			{
				Name:  "encrypt",
				Usage: "encrypt [string | file]",
				Subcommands: []*cli.Command{
					{
						Name:  "string",
						Usage: "encrypt a string",
						Action: func(cCtx *cli.Context) error {
							plaintext := cCtx.Args().First()
							plainbytes := []byte(plaintext)
							keyBytes, _ := otto.GenerateKey()
							keyStringHex := otto.BytesToStringHex(keyBytes)
							fmt.Println("ENCRYPTION KEY (SAVE THIS):", keyStringHex)
							encryptedBytes, _ := otto.Encrypt(plainbytes, keyBytes)
							encodedStringHex := otto.BytesToStringHex(encryptedBytes)
							fmt.Println("OUTPUT (HEX STRING):", encodedStringHex)
							return nil
						},
					},
					{
						Name:  "file",
						Usage: "encrypt a file",
						Action: func(cCtx *cli.Context) error {
							plainfilename := strings.TrimSpace(cCtx.Args().First())
							if _, err := os.Stat(plainfilename); errors.Is(err, os.ErrNotExist) {
								return errors.New("provided file does not exist")
							}
							if plainfilename[len(plainfilename)-4:] == ".enc" {
								return errors.New("cannot encrypt a .enc file. it is already encrypted")
							}
							// TODO: May not be optimal for large files
							plainbytes, err := os.ReadFile(plainfilename)
							if err != nil {
								return err
							}
							keyBytes, _ := otto.GenerateKey()
							keyStringHex := otto.BytesToStringHex(keyBytes)
							fmt.Println("ENCRYPTION KEY (SAVE THIS):", keyStringHex)
							encryptedBytes, _ := otto.Encrypt(plainbytes, keyBytes)
							destFilename := fmt.Sprintf("%s.enc", plainfilename)
							err2 := os.WriteFile(destFilename, encryptedBytes, 0660)
							if err2 != nil {
								return err2
							}
							fmt.Println("Saved encrypted file with .enc extension:", destFilename)
							return nil
						},
					},
				},
			},
			{
				Name:  "decrypt",
				Usage: "decrypt [string | file]",
				Subcommands: []*cli.Command{
					{
						Name:  "string",
						Usage: "decrypt a string",
						Action: func(cCtx *cli.Context) error {
							encryptedString := strings.TrimSpace(cCtx.Args().Get(0))
							encryptedBytes, _ := otto.StringHexToBytes(encryptedString)
							fmt.Println("> Provide Key/Password exactly then press enter:")
							keyStringHex, _ := promptSensitive()
							keyBytes, _ := otto.StringHexToBytes(keyStringHex)
							decryptedBytes, _ := otto.Decrypt(encryptedBytes, keyBytes)
							fmt.Println(string(decryptedBytes))
							return nil
						},
					},
					{
						Name:  "file",
						Usage: "decrypt a file",
						Action: func(cCtx *cli.Context) error {
							encfilename := strings.TrimSpace(cCtx.Args().First())
							if _, err := os.Stat(encfilename); errors.Is(err, os.ErrNotExist) {
								return errors.New("provided file does not exist")
							}
							if encfilename[len(encfilename)-4:] != ".enc" {
								return errors.New("this can only decrypt .enc files")
							}
							// TODO: May not be optimal for large files
							encryptedBytes, err := os.ReadFile(encfilename)
							if err != nil {
								return err
							}
							fmt.Println("> Provide Key/Password exactly then press enter:")
							keyStringHex, _ := promptSensitive()
							keyBytes, _ := otto.StringHexToBytes(keyStringHex)
							decryptedBytes, _ := otto.Decrypt(encryptedBytes, keyBytes)
							destFilename := encfilename[0 : len(encfilename)-4]
							err2 := os.WriteFile(destFilename, decryptedBytes, 0666)
							if err2 != nil {
								return err2
							}
							fmt.Println("Decrypted file:", destFilename)
							return nil
						},
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func promptNormal() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Value: ")
	value, err := reader.ReadString('\n')
	return value, err
}

func promptSensitive() (string, error) {
	byteValue, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	password := string(byteValue)
	return strings.TrimSpace(password), nil
}
