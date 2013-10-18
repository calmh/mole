package main

import (
	"flag"
	"os"

	"github.com/calmh/mole/usage"
)

func init() {
	addCommand(command{name: "rm", fn: rmCommand, descr: msgRmShort})
}

func rmCommand(args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	fs.Usage = usage.For(fs, msgPushUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		os.Exit(3)
	}

	tunnelname := args[0]

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	_, err := authenticated(cl, func() (interface{}, error) {
		return nil, cl.Delete(tunnelname)
	})
	fatalErr(err)

	okf(msgOkDeleted, tunnelname)
}
