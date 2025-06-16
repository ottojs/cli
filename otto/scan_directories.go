package otto

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func Exists(path string) (bool, bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return true, stat.IsDir(), nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, false, nil
	}
	return false, false, err
}

// Scans list of directories for file name pattern
func ScanDirectories(dirs, patterns []string) ([]string, error) {
	var matches []string

	for _, dir := range dirs {
		// Try to cleanse path based on the platform
		// Helpful if we use "/" on windows out of habit
		dir = filepath.Clean(dir)

		// Walk the directory
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			// Permission denied
			if err != nil {
				return nil
				//return err
			}

			// No directories please (you can change this)
			if !d.IsDir() {
				matched := false
				for _, p := range patterns {
					gotmatch, err := filepath.Match(p, filepath.Base(path))
					// Invalid pattern
					if err != nil {
						return err
					}
					if gotmatch {
						matched = true
					}
				}
				// Do we have a match?
				if matched {
					matches = append(matches, path)
				}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return matches, nil
}
