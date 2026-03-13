package argparse

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/pspiagicw/groove/config"
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
	Init     ConfigInitCMD     `cmd:"" help:"Initialize the defaultconfig."`
}

var CLI struct {
	ConfigPath string `help:"Path to config file."`

	Config ConfigCMD `cmd:"" help:"Validate, init or show the config."`
}

func Run(version string) {
	ctx := kong.Parse(&CLI)
	// TODO: Add path sanitizer right here!
	// Probably a helper library or something.
	err := ctx.Run(&Opts{CLI.ConfigPath})
	ctx.FatalIfErrorf(err)
}
