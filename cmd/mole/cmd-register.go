package main

import (
	"crypto/tls"
	"fmt"
	"github.com/calmh/mole/randomart"
	"github.com/jessevdk/go-flags"
	"os"
	"path"
	"regexp"
)

type registerCommand struct {
	Port int `short:"p" long:"port" description:"Server port number" value-name:"PORT"`
}

var registerParser *flags.Parser

func init() {
	cmd := registerCommand{}
	registerParser = globalParser.AddCommand("register", msgRegisterShort, msgRegisterLong, &cmd)
}

func (c *registerCommand) Usage() string {
	return "<server> [register-OPTIONS]"
}

func (c *registerCommand) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		showParser.WriteHelp(os.Stdout)
		infoln()
		fatalln("register: missing required option <server>")
	}

	if c.Port == 0 {
		c.Port = 9443
	}

	server := fmt.Sprintf("%s:%d", args[0], c.Port)
	conn, err := tls.Dial("tcp", server, &tls.Config{InsecureSkipVerify: true})
	fatalErr(err)

	fp := certFingerprint(conn)
	twoDigits := regexp.MustCompile("([0-9a-f]{2})")
	fpstr := twoDigits.ReplaceAllString(fmt.Sprintf("%x", fp), "$1:")
	fpstr = fpstr[:len(fpstr)-1] // trailing colon

	ini := fmt.Sprintf("[server]\nhost = %s\nport = %d\nfingerprint = %s\n", args[0], c.Port, fpstr)
	fd, err := os.Create(path.Join(globalOpts.Home, "mole.ini"))
	fatalErr(err)
	_, err = fd.WriteString(ini)
	fatalErr(err)
	err = fd.Close()
	fatalErr(err)

	infof("%s", randomart.Generate(fp, "mole"))
	okf(msgRegistered, args[0])
	return nil
}
