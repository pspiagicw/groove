package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/pspiagicw/groove/prettylog"
	"github.com/pspiagicw/groove/utils"

	"github.com/adrg/xdg"
)

func getConfigPath() string {
	location, err := xdg.ConfigFile("groove/config.toml")

	if err != nil {
		prettylog.Errorf("Failed to resolve config path: %v", err)
	}

	return location
}

func Init() error {
	location := getConfigPath()

	if utils.AlreadyExists(location) {
		return fmt.Errorf("Config already exists!")
	}

	utils.WriteToFile(location, DEFAULT_CONFIG)
	prettylog.Successf("Successfully created config at %s", location)
	return nil
}

func Show(configPath string) {
	config := loadConfig(configPath)

	config.PrettyPrint(os.Stdout)
}

func loadConfig(configPath string) *Config {
	if configPath == "" {
		configPath = getConfigPath()
	}

	if !utils.AlreadyExists(configPath) {
		prettylog.Fatalf("No config found at %s", configPath)
	}

	contents, err := utils.ReadFromFile(configPath)

	if err != nil {
		prettylog.Fatalf("Failed to read config: %v", err)
	}

	config := new(Config)

	err = toml.Unmarshal(contents, config)

	if err != nil {
		prettylog.Fatalf("Failed to parse config: %v", err)
	}

	return config
}

func Validate(configPath string) {

	// If config loads, then it's valid!
	_ = loadConfig(configPath)

	prettylog.Successf("Config is valid!")
}

func ConfigProvider(configPath string) *Config {
	config := loadConfig(configPath)

	sanitizeConfig(config)

	return config
}

// DONE: Refactor this one if possible.
// TODO: Sanitize other fields, like birate, codec etc.
func sanitizeConfig(config *Config) {

	config.IncomingDir = utils.ExpandAndEnsureExists(config.IncomingDir)

	config.LibraryDir = utils.ExpandAndEnsureExists(config.LibraryDir)

	config.Database = utils.ExpandHome(config.Database)

	dbFolder := filepath.Base(config.Database)
	err := utils.CreateIfNotExist(dbFolder)

	if err != nil {
		prettylog.Fatalf("Failed to create parent folder for database: %v", err)
	}
}
