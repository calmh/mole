package main

import (
	"errors"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/jessevdk/go-flags"
)

var errParams = errors.New("incorrect command line parameters")

var homeDir string

var globalOpts struct {
	Debug bool `short:"d" long:"debug" description:"Show debug output"`
}

var globalParser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	globalParser.ApplicationName = "mole"
	if _, e := globalParser.Parse(); e != nil {
		os.Exit(1)
	}
}

func setup() {
	globalOpts.Debug = globalOpts.Debug || isDebug

	if globalOpts.Debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		log.Println("Debug enabled")
	} else {
		log.SetFlags(0)
	}

	u, e := user.Current()
	if e != nil {
		log.Fatal(e)
	}
	homeDir = path.Join(u.HomeDir, ".mole")
}
