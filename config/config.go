package config

import (
	"os"

	"github.com/adrg/xdg"
)

const DEFAULT_CONFIG = `
incoming_dir = "~/Music/incoming"
library_dir = "~/Music/library"
database = "~/.local/share/musicmgr/music.db"

[import]
default_format = "{album}/{track:02} - {title}.mp3"
copy_mode = "move"

[transcoding]
enabled = false
target_codec = "mp3"
target_bitrate = "320k"

[playlists]
generate_artist = true
generate_genre = true
location = "~/Music/playlists"

[hash]
algorithm = "sha256"
`

func getConfigPath() (string, error) {
	location, err := xdg.ConfigFile("muzic/config.toml")

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

	writeToFile(location, DEFAULT_CONFIG)
	return nil
}

func Show(configPath string) error {
	return nil
}

func Validate(configPath string) error {
	return nil
}

// TODO: Extract into a helper library.
func writeToFile(filepath string, contents string) {
	// TODO: Sanitize this path (create parent directories if it doesn't exist.)
	os.WriteFile(filepath, contents)
}
