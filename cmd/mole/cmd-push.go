package main

import (
	"bytes"
	"flag"
	"github.com/calmh/mole/conf"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	addCommand(command{name: "push", fn: pushCommand, descr: msgPushShort})
}

func pushCommand(args []string) {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	fs.Usage = usageFor(fs, msgPushUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		exit(3)
	}

	filename := filepath.Base(args[0])
	if ext := filepath.Ext(filename); ext != ".ini" {
		fatalf(msgFileNotInit, filename)
	}

	// Read
	file, err := os.Open(args[0])
	fatalErr(err)
	bs, err := ioutil.ReadAll(file)
	fatalErr(err)
	_ = file.Close()

	// Verify
	_, err = conf.Load(bytes.NewBuffer(bs))
	fatalErr(err)

	// Push
	tunnelname := filename[:len(filename)-4]
	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	_, err = authenticated(cl, func() (interface{}, error) {
		return nil, cl.Put(tunnelname, bytes.NewBuffer(bs))
	})
	fatalErr(err)

	okf(msgOkPushed, tunnelname)
}
