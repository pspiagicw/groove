package main

import (
	"github.com/pspiagicw/muzic/argparse"
)

var VERSION string = "unversioned"

func main() {
	argparse.Run(VERSION)
}
