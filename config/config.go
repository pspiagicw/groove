package config

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var Config struct {
	IncomingDir string `toml:"incomingDir"`
	LibraryDir  string `toml:"libraryDir"`
	Database    string

	ImportSettings struct {
		Format string
	} `toml:"import"`

	TranscodingSettings struct {
		Enabled       bool
		TargetCodec   string `toml:"targetCodec"`
		TargetBitrate string `toml:"targetBirate"`
	} `toml:"trancoding"`

	PlaylistSettings struct {
		Location string
	} `toml:"playlists"`
}

const DEFAULT_CONFIG = `
incomingDir = "~/Music/incoming"
libraryDir = "~/Music/library"
database = "~/.local/share/musicmgr/music.db"

[import]
format = "{album}/{track:02} - {title}.mp3"

[transcoding]
enabled = false
targetCodec = "mp3"
targetBitrate = "320k"

[playlists]
location = "~/Music/playlists"
`

func getConfigPath() (string, error) {
	location, err := xdg.ConfigFile("groove/config.toml")

	if err != nil {
		return "", err
	}

	return location, nil
}

func Init() error {
	location, err := getConfigPath()

	if err != nil {
		return err
	}

	if alreadyExists(location) {
		return fmt.Errorf("Config already exists!")
	}

	writeToFile(location, DEFAULT_CONFIG)
	return nil
}

func Show(configPath string) error {
	var err error
	if configPath == "" {
		configPath, err = getConfigPath()

		if err != nil {
			return err
		}
	}

	if !alreadyExists(configPath) {
		return fmt.Errorf("No config found at %s!", configPath)
	}

	err = Validate(configPath)
	if err != nil {
		return fmt.Errorf("Config can't be validated: %v!", err)
	}

	contents, err := readFile(configPath)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(contents, &Config)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", Config)

	return nil
}

func Validate(configPath string) error {
	var err error
	if configPath == "" {
		configPath, err = getConfigPath()

		if err != nil {
			return err
		}
	}

	if !alreadyExists(configPath) {
		return fmt.Errorf("No config found at %s!", configPath)
	}

	contents, err := readFile(configPath)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(contents, &Config)
	if err != nil {
		return err
	}

	fmt.Println("No problems found with the config!")
	return nil
}

// TODO: Extract into a helper library.
func writeToFile(file string, contents string) {
	createIfNotExist(filepath.Dir(file))
	os.WriteFile(file, []byte(contents), 0644)
}

func readFile(file string) ([]byte, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func createIfNotExist(folder string) error {
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

func alreadyExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)

}
