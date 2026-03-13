package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func AlreadyExists(path string) bool {

	_, err := os.Stat(path)
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
