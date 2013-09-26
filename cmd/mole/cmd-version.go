package main

import (
	"github.com/jessevdk/go-flags"
)

type versionCommand struct{}

var versionParser *flags.Parser

func init() {
	cmd := versionCommand{}
	versionParser = globalParser.AddCommand("version", msgVersionShort, msgVersionLong, &cmd)
}

func (c *versionCommand) Execute(args []string) error {
	setup()
	printVersion()
	return nil
}
