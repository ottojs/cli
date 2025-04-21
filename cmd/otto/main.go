package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
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
			{
				Name:  "pack",
				Usage: "packs up directory of files into zip",
				Action: func(cCtx *cli.Context) error {
					target := strings.TrimSpace(cCtx.Args().Get(0))
					if target == "" {
						return errors.New("please provide a directory")
					}
					filestat, err := os.Stat(target)
					if errors.Is(err, os.ErrNotExist) {
						return errors.New("provided directory does not exist")
					}
					if !filestat.IsDir() {
						return errors.New("provided file is not a directory")
					}
					filename := fmt.Sprintf("%s.zip", target)
					err2 := otto.ZipDirectory(target, filename)
					fmt.Println("Saved ZIP File", filename)
					return err2
				},
			},
			{
				Name:  "unpack",
				Usage: "unpacks zip file into its directory",
				Action: func(cCtx *cli.Context) error {
					target := strings.TrimSpace(cCtx.Args().Get(0))
					if target == "" {
						return errors.New("please provide a value")
					}
					if target[len(target)-4:] != ".zip" {
						return errors.New("you need to provide a .zip file")
					}
					// Unzip to current directory
					err := otto.UnzipDirectory(target, "")
					return err
				},
			},
			{
				Name:  "download",
				Usage: "downloads a file from URL (HTTP GET, filename is from URL)",
				Action: func(cCtx *cli.Context) error {
					target := strings.TrimSpace(cCtx.Args().Get(0))
					if target == "" {
						return errors.New("please provide a URL")
					}
					_, err := url.ParseRequestURI(target)
					if err != nil {
						return errors.New("provided URL is invalid")
					}
					err2 := otto.DownloadURL(target)
					return err2
				},
			},
			{
				Name:  "scandir",
				Usage: "scans directory for matching file names",
				Action: func(cCtx *cli.Context) error {
					target := strings.TrimSpace(cCtx.Args().Get(0))
					if target == "" {
						return errors.New("please provide a directory path")
					}
					exists, isDir, err1 := otto.Exists(target)
					if err1 != nil {
						return err1
					}
					if !exists || !isDir {
						return errors.New("invalid target, it either does not exist or is not a directory")
					}
					query := strings.TrimSpace(cCtx.Args().Get(1))
					results, err2 := otto.ScanDirectories([]string{target}, query)
					if err2 != nil {
						return err2
					}
					for _, f := range results {
						fmt.Println(f)
					}
					return nil
				},
			},
			{
				Name:  "os",
				Usage: "shows the operating system name",
				Action: func(cCtx *cli.Context) error {
					// This is to demonstrate os-specific file includes with build flags
					// See the "settings_*.go" files
					fmt.Println(otto.OperatingSystem)
					return nil
				},
			},
			{
				Name:  "findchunk",
				Usage: "checks for chunk in file",
				Action: func(cCtx *cli.Context) error {
					needle := []string{
						"where's otto?",
						"another line",
					}
					target := strings.TrimSpace(cCtx.Args().Get(0))
					if target == "" {
						return errors.New("please provide a path to a file")
					}
					exists, isDir, err1 := otto.Exists(target)
					if err1 != nil {
						return err1
					}
					if !exists || isDir {
						return errors.New("invalid target, it either does not exist or is a directory")
					}
					found, err2 := otto.FindChunkInFile(target, needle)
					if err2 != nil {
						return err2
					}
					fmt.Println("Searched:", target)
					fmt.Println("... for...")
					fmt.Println(strings.Join(needle, "\n"))
					fmt.Println("Found?:", found)
					return nil
				},
			},
			{
				Name:  "execute",
				Usage: "executes a cli commanmd",
				Action: func(cCtx *cli.Context) error {
					// Default to Linux/macOS
					binary := "ls"
					args := []string{"-l"}
					if runtime.GOOS == "windows" {
						binary = "dir"
						args = nil
					}

					// Run
					stdout, stderr, err := otto.ExecuteBinary(binary, args...)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						fmt.Printf("Stderr: %s\n", stderr)
						return nil
					}

					// All went well
					fmt.Printf("Stdout: %s\n", stdout)
					if stderr != "" {
						fmt.Printf("Stderr: %s\n", stderr)
					}
					return nil
				},
			},
			{
				Name:  "append",
				Usage: "appends content to a file and creates a backup beforehand",
				Action: func(cCtx *cli.Context) error {
					content := []string{
						"one fish",
						"two fish",
						"red fish",
						"blue fish",
					}
					// Obfuscate if you like (16-char, but use whatever)
					encoded, err1 := otto.EncodeStrings(content, "4a7b9c2d5e8f1b3a")
					if err1 != nil {
						return err1
					}
					fmt.Println("Encoded:")
					for _, v := range encoded {
						fmt.Println(v)
					}
					decoded, err2 := otto.DecodeStrings(encoded, "4a7b9c2d5e8f1b3a")
					if err2 != nil {
						return err2
					}
					err3 := otto.AppendToFileWithBackup("test/findchunk.txt", strings.Join(decoded, "\n")+"\n", ".bak")
					return err3
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
