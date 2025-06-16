package otto

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func DownloadURL(theurl string) error {
	// Validate URL
	parsed, err := url.ParseRequestURI(theurl)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Make the Request
	resp, err := http.Get(theurl)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Make sure the URL returns 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Extract filename from URL
	filename := path.Base(parsed.Path)

	// Handle edge cases for filename
	if filename == "" || filename == "." || filename == "/" {
		filename = "download"
	}

	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)

	// Remove any potentially dangerous characters
	filename = strings.ReplaceAll(filename, "..", "")

	fmt.Println("Filename:", filename)

	// Create the file with secure permissions
	out, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy contents to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// Attempt to remove partially downloaded file
		os.Remove(filename)
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
