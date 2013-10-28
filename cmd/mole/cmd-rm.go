package main

import (
	"flag"
)

func init() {
	addCommand(command{name: "rm", fn: rmCommand, descr: msgRmShort})
}

func rmCommand(args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	fs.Usage = usageFor(fs, msgPushUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		exit(3)
	}

	tunnelname := args[0]

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	_, err := authenticated(cl, func() (interface{}, error) {
		return nil, cl.Delete(tunnelname)
	})
	fatalErr(err)

	okf(msgOkDeleted, tunnelname)
}
