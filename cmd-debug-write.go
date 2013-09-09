// +build debug

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/tmpfileset"
	"github.com/jessevdk/go-flags"
)

type cmdWrite struct{}

var writeParser *flags.Parser

func init() {
	cmd := cmdWrite{}
	writeParser = globalParser.AddCommand("debug-write", "Show write configuration", "write generates the write configuration file script for the specified tunnel configuration file", &cmd)
}

func (c *cmdWrite) Execute(args []string) error {
	setup()

	if len(args) != 2 {
		writeParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("debug-write: missing required options <filename> <directory>\n")
	}

	cfg, err := configuration.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}

	var fs tmpfileset.FileSet
	sshConfig(cfg, &fs)
	expectConfig(cfg, &fs)
	fs.Save(args[1])

	return nil
}
