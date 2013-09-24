package main

import (
	"bytes"
	"github.com/calmh/mole/conf"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"os"
	"path/filepath"
)

type cmdPush struct{}

var pushParser *flags.Parser

func init() {
	cmd := cmdPush{}
	pushParser = globalParser.AddCommand("push", msgPushShort, msgPushLong, &cmd)
}

func (c *cmdPush) Usage() string {
	return "<tunnelfile>"
}

func (c *cmdPush) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		digParser.WriteHelp(os.Stdout)
		infoln()
		fatalln("push: missing required option <tunnelfile>")
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
	file.Close()

	// Verify
	_, err = conf.Load(bytes.NewBuffer(bs))
	fatalErr(err)
	file.Close()

	// Push
	tunnelname := filename[:len(filename)-4]
	cl := NewClient(serverIni.address, serverIni.fingerprint)
	_, err = authenticated(cl, func() (interface{}, error) {
		return nil, cl.Put(tunnelname, bytes.NewBuffer(bs))
	})
	fatalErr(err)

	okf(msgOkPushed, tunnelname)
	return nil
}
