package main

import (
	"log"
	"runtime"

	"github.com/jessevdk/go-flags"
)

var (
	buildVersion string
	buildDate    string
	buildUser    string
)

type cmdVersion struct{}

var versionParser *flags.Parser

func init() {
	cmd := cmdVersion{}
	versionParser = globalParser.AddCommand("version", "Show version", "'version' shows current and latest available client and server versions", &cmd)
}

func (c *cmdVersion) Execute(args []string) error {
	setup()

	log.Printf("mole (%s-%s)", runtime.GOOS, runtime.GOARCH)
	log.Printf("  %s (%s)", buildVersion, buildKind)
	if buildDate != "" {
		log.Printf("  %s by %s", buildDate, buildUser)
	}

	return nil
}
