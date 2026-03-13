package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"github.com/pspiagicw/groove/utils"

	"github.com/adrg/xdg"
)

type Config struct {
	IncomingDir string `toml:"incomingDir"`
	LibraryDir  string `toml:"libraryDir"`
	Database    string `toml:"database"`

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

func (c Config) PrettyPrint(w io.Writer) {
	fmt.Fprintln(w, "Configuration")
	fmt.Fprintln(w, "-------------")

	fmt.Fprintf(w, "Incoming Directory : %s\n", c.IncomingDir)
	fmt.Fprintf(w, "Library Directory  : %s\n", c.LibraryDir)
	fmt.Fprintf(w, "Database           : %s\n", c.Database)

	fmt.Fprintln(w)

	fmt.Fprintln(w, "[Import]")
	fmt.Fprintf(w, "  Format           : %s\n", c.ImportSettings.Format)

	fmt.Fprintln(w)

	fmt.Fprintln(w, "[Transcoding]")
	fmt.Fprintf(w, "  Enabled          : %t\n", c.TranscodingSettings.Enabled)
	fmt.Fprintf(w, "  Target Codec     : %s\n", c.TranscodingSettings.TargetCodec)
	fmt.Fprintf(w, "  Target Bitrate   : %s\n", c.TranscodingSettings.TargetBitrate)

	fmt.Fprintln(w)

	fmt.Fprintln(w, "[Playlists]")
	fmt.Fprintf(w, "  Location         : %s\n", c.PlaylistSettings.Location)
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

	if utils.AlreadyExists(location) {
		return fmt.Errorf("Config already exists!")
	}

	utils.WriteToFile(location, DEFAULT_CONFIG)
	return nil
}

func Show(configPath string) error {
	config, err := loadConfig(configPath)

	if err != nil {
		return fmt.Errorf("Error while loading config: %v", err)
	}

	config.PrettyPrint(os.Stdout)

	return nil
}

func loadConfig(configPath string) (*Config, error) {
	var err error

	if configPath == "" {
		configPath, err = getConfigPath()

		if err != nil {
			return nil, err
		}
	}

	if !utils.AlreadyExists(configPath) {
		return nil, fmt.Errorf("No config found at %s!", configPath)
	}

	contents, err := utils.ReadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	config := new(Config)

	err = toml.Unmarshal(contents, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func Validate(configPath string) error {

	_, err := loadConfig(configPath)

	if err != nil {
		return fmt.Errorf("Error while loading config: %v", err)
	}

	return nil
}

func ConfigProvier(configPath string) (*Config, error) {
	config, err := loadConfig(configPath)

	if err != nil {
		return nil, fmt.Errorf("Error loading config: %v!", err)
	}

	err = sanitizeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Sanitization failed: %v!", err)
	}

	return config, nil
}
func sanitizeConfig(config *Config) error {
	// TODO: Add homedir.expand to each one.

	path, err := homedir.Expand(config.IncomingDir)
	if err != nil {
		return err
	}
	config.IncomingDir = path

	if !utils.AlreadyExists(config.IncomingDir) {
		return fmt.Errorf("No such directory for incoming files: %v!", config.IncomingDir)
	}

	path, err = homedir.Expand(config.LibraryDir)
	if err != nil {
		return err
	}

	config.LibraryDir = path

	if !utils.AlreadyExists(config.LibraryDir) {
		return fmt.Errorf("No such directory for incoming files: %v!", config.IncomingDir)
	}

	path, err = homedir.Expand(config.Database)

	if err != nil {
		return err
	}
	config.Database = path

	dbFolder := filepath.Base(config.Database)
	err = utils.CreateIfNotExist(dbFolder)
	if err != nil {
		return fmt.Errorf("Error creating parent folders for database: %v!", err)
	}
	// TODO: Sanitize other fields, like birate, codec etc.

	return nil
}
