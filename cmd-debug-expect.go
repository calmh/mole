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

type cmdExpect struct{}

var expectParser *flags.Parser

func init() {
	cmd := cmdExpect{}
	expectParser = globalParser.AddCommand("debug-expect", "Show expect script", "Expect generates the expect script for the specified tunnel configuration file", &cmd)
}

func (c *cmdExpect) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		expectParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("debug-expect: missing required option <filename>\n")
	}

	cfg, err := configuration.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}

	var fs tmpfileset.FileSet
	expectConfig(cfg, &fs)
	fmt.Println(fs)

	return nil
}
