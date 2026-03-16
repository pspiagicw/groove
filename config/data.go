package config

import (
	"io"

	"github.com/pspiagicw/groove/prettylog"
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

const DEFAULT_CONFIG = `
incomingDir = "~/Music/incoming"
libraryDir = "~/Music/library"
database = "~/.local/share/groove/music.db"

[import]
format = "{album}/{track:02} - {title}.mp3"

[transcoding]
enabled = false
targetCodec = "mp3"
targetBitrate = "320k"

[playlists]
location = "~/Music/playlists"
`

func (c Config) PrettyPrint(w io.Writer) {
	prettylog.PrintBlock(
		w,
		"Configuration",
		prettylog.Section("Paths"),
		prettylog.KV("Incoming Dir", c.IncomingDir),
		prettylog.KV("Library Dir", c.LibraryDir),
		prettylog.KV("Database", c.Database),
		"",
		prettylog.Section("Import"),
		prettylog.KV("Format", c.ImportSettings.Format),
		"",
		prettylog.Section("Transcoding"),
		prettylog.BoolKV("Status", c.TranscodingSettings.Enabled),
		prettylog.KV("Target Codec", c.TranscodingSettings.TargetCodec),
		prettylog.KV("Target Bitrate", c.TranscodingSettings.TargetBitrate),
		"",
		prettylog.Section("Playlists"),
		prettylog.KV("Location", c.PlaylistSettings.Location),
	)
}
