package main

var CLI struct {
	ConfigPath string `help:"Path to config file."`

	Import struct {
	} `cmd:"" help:"Import Music"`
}
