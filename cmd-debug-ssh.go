package main

import (
	"fmt"
	"log"
	"os"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/ssh"
	"github.com/jessevdk/go-flags"
)

type cmdSsh struct{}

var sshParser *flags.Parser

func init() {
	cmd := cmdSsh{}
	sshParser = globalParser.AddCommand("debug-ssh", "Show ssh configuration", "Ssh generates the ssh configuration file script for the specified tunnel configuration file", &cmd)
}

func (c *cmdSsh) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		sshParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("debug-ssh: missing required option <filename>\n")
	}

	cfg, err := configuration.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}

	ssh := ssh.New(cfg)
	fmt.Println(ssh)

	return nil
}
