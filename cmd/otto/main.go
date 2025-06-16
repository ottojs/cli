package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"code.ottojs.org/go/otto"
	cli "github.com/urfave/cli/v2"
)

// Build-time variables (set via -ldflags)
var (
	Version   = "mainline"
	GoVersion = runtime.Version()
)

func WaitSeconds(seconds int) {
	fmt.Println("")
	fmt.Println("==========")
	fmt.Printf("Waiting %d seconds\n", seconds)
	time.Sleep(time.Second * time.Duration(seconds))
}

// Displays an interactive menu for Windows users
func showInteractiveMenu() error {
	app := createApp()

	// Clear screen on Windows
	otto.ExecuteBinary("cmd", "/c", "cls")

	for {
		fmt.Println("\n========== OTTO CLI MENU ==========")
		fmt.Println("Commands:")
		fmt.Println("  1 | encrypt file <filepath>")
		fmt.Println("  2 | decrypt file <filepath>")
		fmt.Println("  3 | encrypt string <text>")
		fmt.Println("  4 | decrypt string <hex>")
		fmt.Println("  5 | pack <directory>")
		fmt.Println("  6 | unpack <zipfile>")
		fmt.Println("  7 | download <url>")
		fmt.Println("  8 | scandir <directory> <pattern>")
		fmt.Println("  9 | is-admin")
		fmt.Println("")
		fmt.Println("Usage: <number> [arguments...]")
		fmt.Println("Example: 1 myfile.txt")
		fmt.Println("===================================")
		fmt.Print("\nEnter command: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Parse input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		// Map menu numbers to commands
		commandMap := map[string][]string{
			"1": {"encrypt", "file"},
			"2": {"decrypt", "file"},
			"3": {"encrypt", "string"},
			"4": {"decrypt", "string"},
			"5": {"pack"},
			"6": {"unpack"},
			"7": {"download"},
			"8": {"scandir"},
			"9": {"is-admin"},
		}

		// Convert menu number to command
		if cmd, ok := commandMap[parts[0]]; ok {
			// Build new args array: program name + command + user args
			newArgs := []string{"otto"}
			newArgs = append(newArgs, cmd...)
			if len(parts) > 1 {
				newArgs = append(newArgs, parts[1:]...)
			}

			// Run the command through the CLI app
			if err := app.Run(newArgs); err != nil {
				fmt.Printf("\nError: %v\n", err)
			}
		} else {
			// Try running as direct command
			newArgs := append([]string{"otto"}, parts...)
			if err := app.Run(newArgs); err != nil {
				fmt.Printf("\nError: %v\n", err)
			}
		}

		promptToContinue()

		// Clear screen for next iteration
		otto.ExecuteBinary("cmd", "/c", "cls")
	}
}

func promptToContinue() {
	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
}

func showVersionInfo() {
	fmt.Printf("\nOtto CLI %s\n", Version)
	fmt.Printf("Built with: %s\n", GoVersion)
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// Creates and returns the CLI app with all commands
func createApp() *cli.App {
	return &cli.App{
		Name:    "otto",
		Version: Version,
		Usage:   "Otto Tools CLI",
		Authors: []*cli.Author{
			{
				Name:  "Otto.js",
				Email: "help@ottojs.org",
			},
		},
		Copyright: "(c) 2024-2025 Otto.js",
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Built with": GoVersion,
				"Platform":   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}
		},
		Commands: getCommands(),
		Action: func(c *cli.Context) error {
			// Show version when requested
			if c.Bool("version") {
				cli.ShowVersion(c)
				return nil
			}
			// Otherwise show help
			return cli.ShowAppHelp(c)
		},
	}
}

// Returns all CLI commands
func getCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "encrypt",
			Usage: "encrypt [string | file]",
			Subcommands: []*cli.Command{
				{
					Name:  "string",
					Usage: "encrypt a string",
					Action: func(cCtx *cli.Context) error {
						plaintext := cCtx.Args().First()
						if plaintext == "" {
							return errors.New("please provide a string to encrypt")
						}
						plainbytes := []byte(plaintext)
						keyBytes, err := otto.GenerateKey()
						if err != nil {
							return fmt.Errorf("failed to generate encryption key: %w", err)
						}
						keyStringHex := otto.BytesToStringHex(keyBytes)
						fmt.Println("ENCRYPTION KEY (SAVE THIS):", keyStringHex)
						encryptedBytes, err := otto.Encrypt(plainbytes, keyBytes)
						if err != nil {
							return fmt.Errorf("encryption failed: %w", err)
						}
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
						if strings.HasSuffix(plainfilename, ".enc") {
							return errors.New("cannot encrypt a .enc file. it is already encrypted")
						}
						// TODO: May not be optimal for large files
						plainbytes, err := os.ReadFile(plainfilename)
						if err != nil {
							return err
						}
						keyBytes, err := otto.GenerateKey()
						if err != nil {
							return fmt.Errorf("failed to generate encryption key: %w", err)
						}
						keyStringHex := otto.BytesToStringHex(keyBytes)
						fmt.Println("ENCRYPTION KEY (SAVE THIS):", keyStringHex)
						encryptedBytes, err := otto.Encrypt(plainbytes, keyBytes)
						if err != nil {
							return fmt.Errorf("encryption failed: %w", err)
						}
						destFilename := fmt.Sprintf("%s.enc", plainfilename)
						err2 := os.WriteFile(destFilename, encryptedBytes, 0600)
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
						if encryptedString == "" {
							return errors.New("please provide an encrypted string to decrypt")
						}
						encryptedBytes, err := otto.StringHexToBytes(encryptedString)
						if err != nil {
							return fmt.Errorf("invalid hex string: %w", err)
						}
						fmt.Println("> Provide Key/Password exactly then press enter:")
						keyStringHex, err := otto.PromptSensitive()
						if err != nil {
							return fmt.Errorf("failed to read key: %w", err)
						}
						keyBytes, err := otto.StringHexToBytes(keyStringHex)
						if err != nil {
							return fmt.Errorf("invalid key format: %w", err)
						}
						decryptedBytes, err := otto.Decrypt(encryptedBytes, keyBytes)
						if err != nil {
							return fmt.Errorf("decryption failed: %w", err)
						}
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
						if !strings.HasSuffix(encfilename, ".enc") {
							return errors.New("this can only decrypt .enc files")
						}
						// TODO: May not be optimal for large files
						encryptedBytes, err := os.ReadFile(encfilename)
						if err != nil {
							return err
						}
						fmt.Println("> Provide Key/Password exactly then press enter:")
						keyStringHex, err := otto.PromptSensitive()
						if err != nil {
							return fmt.Errorf("failed to read key: %w", err)
						}
						keyBytes, err := otto.StringHexToBytes(keyStringHex)
						if err != nil {
							return fmt.Errorf("invalid key format: %w", err)
						}
						decryptedBytes, err := otto.Decrypt(encryptedBytes, keyBytes)
						if err != nil {
							return fmt.Errorf("decryption failed: %w", err)
						}
						destFilename := strings.TrimSuffix(encfilename, ".enc")
						err2 := os.WriteFile(destFilename, decryptedBytes, 0600)
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
				if !strings.HasSuffix(target, ".zip") {
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
				results, err2 := otto.ScanDirectories([]string{target}, []string{query})
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
				fmt.Println(strings.Join(needle, otto.OSNewLine))
				fmt.Println("Found?:", found)
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
				err3 := otto.AppendToFileWithBackup("test/findchunk.txt", strings.Join(decoded, otto.OSNewLine)+otto.OSNewLine, ".bak")
				return err3
			},
		},
		{
			Name:  "is-admin",
			Usage: "checks for admin",
			Action: func(cCtx *cli.Context) error {
				// Check for Admin
				otto.Log("Admin Check:", strconv.FormatBool(otto.AdminCheck()))
				if !otto.AdminCheck() {
					fmt.Println("this tool must be run with admin/root privileges")
					WaitSeconds(10)
					return errors.New("not admin")
				}
				fmt.Println("You are an admin!")
				fmt.Println("Home Dir:", otto.OSHomeDir())
				fmt.Println("Program Path:", otto.OSProgramPath("Otto"))
				fmt.Println("Env PATH:", otto.EnvVarGet("PATH"))
				WaitSeconds(10)
				return nil
			},
		},
		{
			Name:  "version",
			Usage: "show version information",
			Action: func(cCtx *cli.Context) error {
				showVersionInfo()
				return nil
			},
		},
	}
}

func main() {
	app := createApp()

	// Set custom action for no arguments on Windows
	app.Action = func(c *cli.Context) error {
		// On Windows, if no command is provided, show interactive menu
		if runtime.GOOS == "windows" && c.NArg() == 0 && len(os.Args) == 1 {
			return showInteractiveMenu()
		}
		// On other platforms or if arguments provided, show help
		return cli.ShowAppHelp(c)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
