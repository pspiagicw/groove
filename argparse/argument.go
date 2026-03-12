package argparse

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/pspiagicw/muzic/config"
)

type Opts struct {
	ConfigPath string
}

type InitCMD struct {
}

func (i *InitCMD) Run(opts *Opts) error {
	fmt.Printf("Create default config at %s\n", opts.ConfigPath)
	return config.Init()
}

type ConfigShowCMD struct {
}

func (c *ConfigShowCMD) Run(opts *Opts) error {
	fmt.Printf("Show the config at %s\n", opts.ConfigPath)
	return config.Show(opts.ConfigPath)
}

type ConfigValidateCMD struct {
}

func (c *ConfigValidateCMD) Run(opts *Opts) error {
	fmt.Printf("Validate the config at %s\n", opts.ConfigPath)
	return config.Validate(opts.ConfigPath)
}

type ConfigCMD struct {
	Show     ConfigShowCMD     `cmd:"" help:"Show the current config."`
	Validate ConfigValidateCMD `cmd:"" help:"Validate the current config."`
}

var CLI struct {
	ConfigPath string `help:"Path to config file."`

	Init   InitCMD   `cmd:"" help:"Initialize the default config."`
	Config ConfigCMD `cmd:"" help:"Validate or show the config."`
}

func Run(version string) {
	ctx := kong.Parse(&CLI)
	// TODO: Add path sanitizer right here!
	// Probably a helper library or something.
	err := ctx.Run(&Opts{CLI.ConfigPath})
	ctx.FatalIfErrorf(err)
}
