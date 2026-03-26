package argparse

import (
	"github.com/alecthomas/kong"
	"github.com/pspiagicw/groove/commands"
	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/prettylog"
)

type Opts struct {
	ConfigPath string
}

type ConfigInitCMD struct {
}

func (i *ConfigInitCMD) Run(opts *Opts) error {
	return config.Init()
}

type ConfigShowCMD struct {
}

func (c *ConfigShowCMD) Run(opts *Opts) error {
	config.Show(opts.ConfigPath)
	return nil
}

type ConfigValidateCMD struct {
}

func (c *ConfigValidateCMD) Run(opts *Opts) error {
	config.Validate(opts.ConfigPath)
	return nil
}

type ConfigCMD struct {
	Show     ConfigShowCMD     `cmd:"" help:"Show the current config."`
	Validate ConfigValidateCMD `cmd:"" help:"Validate the current config."`
	Init     ConfigInitCMD     `cmd:"" help:"Initialize the default config."`
}

type ScanCMD struct {
}

func (s *ScanCMD) Run(opts *Opts) error {
	commands.Scan(opts.ConfigPath)
	return nil
}

type ImportCMD struct {
}

func (s *ImportCMD) Run(opts *Opts) error {
	commands.Import(opts.ConfigPath)
	return nil
}

type CopyCMD struct {
	DryRun bool `help:"Dry run the import process."`
}

func (c *CopyCMD) Run(opts *Opts) error {
	commands.Move(opts.ConfigPath)
	return nil
}

var CLI struct {
	ConfigPath string `help:"Path to config file."`

	Config ConfigCMD `cmd:"" help:"Validate, init or show the config."`
	Scan   ScanCMD   `cmd:""  help:"Scan incoming directory for music."`
	Import ImportCMD `cmd:""  help:"Import music."`
	Copy   CopyCMD   `cmd:"" help:"Copy music files"`
}

func Run(version string) {
	ctx := kong.Parse(&CLI)
	err := ctx.Run(&Opts{CLI.ConfigPath})
	if err != nil {
		prettylog.Fatalf("CLI command failed: %v", err)
	}
}
