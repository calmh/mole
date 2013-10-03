package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/calmh/mole/randomart"
	"os"
	"path"
	"regexp"
)

func init() {
	commands["register"] = command{registerCommand, msgRegisterShort}
}

func hexBytes(bs []byte) string {
	twoDigits := regexp.MustCompile("([0-9A-F]{2})")
	str := fmt.Sprintf("%X", bs)
	str = twoDigits.ReplaceAllString(str, "$1:")
	str = str[:len(str)-1] // trailing colon
	return str
}

func registerCommand(args []string) error {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	port := fs.Int("port", 9443, "Server port number")
	fs.Usage = usageFor(fs, msgRegisterUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		os.Exit(3)
	}

	server := fmt.Sprintf("%s:%d", args[0], *port)
	conn, err := tls.Dial("tcp", server, &tls.Config{InsecureSkipVerify: true})
	fatalErr(err)

	fp := certFingerprint(conn)
	fpstr := hexBytes(fp)

	ini := fmt.Sprintf("[server]\nhost = %s\nport = %d\nfingerprint = %s\n", args[0], *port, fpstr)
	fd, err := os.Create(path.Join(globalOpts.Home, "mole.ini"))
	fatalErr(err)
	_, err = fd.WriteString(ini)
	fatalErr(err)
	err = fd.Close()
	fatalErr(err)

	infof("%s", randomart.Generate(fp, "mole"))
	infoln(fpstr)
	okf(msgRegistered, args[0])
	return nil
}
