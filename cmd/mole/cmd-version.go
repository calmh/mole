package main

import (
	"github.com/jessevdk/go-flags"
)

type cmdVersion struct{}

var versionParser *flags.Parser

func init() {
	cmd := cmdVersion{}
	versionParser = globalParser.AddCommand("version", msgVersionShort, msgVersionLong, &cmd)
}

func (c *cmdVersion) Execute(args []string) error {
	setup()
	printVersion()
	return nil
}
