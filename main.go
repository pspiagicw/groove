package main

import (
	"github.com/pspiagicw/groove/argparse"
)

var VERSION string = "unversioned"

func main() {
	argparse.Run(VERSION)
}
