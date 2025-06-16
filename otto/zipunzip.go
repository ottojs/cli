package otto

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Docs
// https://pkg.go.dev/archive/zip

func ZipDirectory(directory, target string) error {
	// Validate and sanitize source directory
	safeDirectory, err := SecurePath(directory)
	if err != nil {
		return fmt.Errorf("invalid source directory: %w", err)
	}

	// Validate and sanitize target file
	safeTarget, err := SecurePath(target)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// Create Target ZIP File with secure permissions
	filehandle, err := os.OpenFile(safeTarget, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer filehandle.Close()

	// Transform filehandle into a specialized ZipWriter
	writer := zip.NewWriter(filehandle)
	// We'll close the writer after all files have been scanned and saved (end of function)
	defer writer.Close()

	// Loop through files in directory
	return filepath.Walk(safeDirectory, ZipFileFunction(writer, safeDirectory))
}

// Closure function factory
func ZipFileFunction(writer *zip.Writer, directory string) func(path string, info fs.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Create ZIP file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate

		// Add filepath to header
		header.Name, err = filepath.Rel(filepath.Dir(directory), path)
		if err != nil {
			return err
		}
		// ZIP requires explicit trailing slash for directories (rightfully)
		if info.IsDir() {
			header.Name += "/"
		}

		// Save the file header information
		// Note: We do not want to close the writer yet, because we're adding more files
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		// Directories have no file content
		// So we are done and return
		if info.IsDir() {
			return nil
		}

		// Open the file
		filehandle, err := os.Open(path)
		if err != nil {
			return err
		}
		defer filehandle.Close()

		// Push the file contents into the header
		_, err = io.Copy(headerWriter, filehandle)
		return err
	}
}

func UnzipDirectory(source, destination string) error {
	// Validate and sanitize source ZIP file
	safeSource, err := SecurePath(source)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	// If destination is empty, use current directory
	if destination == "" {
		destination = "."
	}

	// Validate and sanitize destination directory
	safeDest, err := SecurePath(destination)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Get absolute path for destination
	absDestination, err := filepath.Abs(safeDest)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Open ZIP File
	zipreader, err := zip.OpenReader(safeSource)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipreader.Close()

	// Loop to extract all files
	for _, f := range zipreader.File {
		err := UnzipFile(f, absDestination)
		if err != nil {
			return err
		}
	}

	return nil
}

func UnzipFile(filehandle *zip.File, destination string) error {
	// ZipSlip Check
	// https://security.snyk.io/research/zip-slip-vulnerability
	thefilepath := filepath.Join(destination, filehandle.Name)
	if !strings.HasPrefix(thefilepath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("potential zipslip attack during file: %s", thefilepath)
	}

	// If directory, create subdirectories
	// Then return to continue to next item
	if filehandle.FileInfo().IsDir() {
		if err := os.MkdirAll(thefilepath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	// Open Compressed file for reading
	zippedFile, err := filehandle.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(thefilepath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open Destination file for writing with secure permissions
	// Use more restrictive permissions than in the zip file
	mode := filehandle.Mode()
	if mode.Perm() > 0600 {
		mode = 0600
	}
	destinationFile, err := os.OpenFile(thefilepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer destinationFile.Close()

	// Copy from A => B
	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
