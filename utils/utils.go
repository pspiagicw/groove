package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pspiagicw/groove/prettylog"
)

func AlreadyExists(path string) bool {

	_, err := os.Stat(path)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		prettylog.Fatalf("Failed to check path: %v", err)
	}

	return !errors.Is(err, fs.ErrNotExist)

}

func CreateIfNotExist(folder string) error {
	if _, err := os.Stat(folder); errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(folder, 0755)
		if err != nil {
			return fmt.Errorf("Error creating directory: %s", folder)
		}
	} else if err != nil {
		return fmt.Errorf("Error stating file: %s", err)
	}
	return nil
}

func WriteToFile(file string, contents string) error {
	// Create parent directories if doesn't exist.
	err := CreateIfNotExist(filepath.Dir(file))
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(contents), 0644)
}

func ReadFromFile(file string) ([]byte, error) {
	contents, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	return contents, nil
}
func CopyFile(from, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("Error opening file for copying: %v!", err)
	}

	o, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("Error opening destination file for copying: %v!", err)
	}

	_, err = o.ReadFrom(f)
	if err != nil {
		return fmt.Errorf("Error copying file contents: %v!", err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("Error closing file: %v!", err)

	}

	return nil
}

// Helper function to expand home.
// Fatalf if some error occurs!
func ExpandHome(path string) string {

	path, err := homedir.Expand(path)

	if err != nil {
		prettylog.Fatalf("Failed to expand ~: %v", err)
	}

	return path
}

func ExpandAndEnsureExists(path string) string {
	expandPath := ExpandHome(path)

	if !AlreadyExists(expandPath) {
		prettylog.Fatalf("No such path exists (%s)", expandPath)
	}

	return expandPath
}

func IsMusicFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".mp3" || ext == ".flac" || ext == ".opus" || ext == ".wav" || ext == ".m4a" || ext == ".ogg"
}
