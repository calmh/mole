package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type rmCommand struct{}

var rmParser *flags.Parser

func init() {
	cmd := rmCommand{}
	rmParser = globalParser.AddCommand("rm", msgRmShort, msgRmLong, &cmd)
}

func (c *rmCommand) Usage() string {
	return "<tunnel>"
}

func (c *rmCommand) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		digParser.WriteHelp(os.Stdout)
		infoln()
		fatalln("rm: missing required option <tunnel>")
	}

	tunnelname := args[0]

	cl := NewClient(serverIni.address, serverIni.fingerprint)
	_, err := authenticated(cl, func() (interface{}, error) {
		return nil, cl.Delete(tunnelname)
	})
	fatalErr(err)

	okf(msgOkDeleted, tunnelname)
	return nil
}
