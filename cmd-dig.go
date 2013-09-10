package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/tmpfileset"
	"github.com/jessevdk/go-flags"
)

type cmdDig struct {
	Local bool `short:"l" long:"local" description:"Local file, not remote tunnel definition"`
}

var digParser *flags.Parser

func init() {
	cmd := cmdDig{}
	digParser = globalParser.AddCommand("dig", "Dig a tunnel", "'dig' connects to a remote destination and sets up configured local TCP tunnels", &cmd)
}

func (c *cmdDig) Usage() string {
	return "<tunnelname> [dig-OPTIONS]"
}

func (c *cmdDig) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		digParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("dig: missing required option <tunnelname>\n")
	}

	var cfg *configuration.Config
	var err error

	if c.Local {
		cfg, err = configuration.LoadFile(args[0])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cert := certificate()
		cl := NewClient("mole.nym.se:9443", cert)
		cfg, err = configuration.LoadString(cl.Get(args[0]))
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg == nil {
		return fmt.Errorf("no tunnel loaded")
	}

	var fs tmpfileset.FileSet
	sshConfig(cfg, &fs)
	expectConfig(cfg, &fs)

	defer fs.Remove()
	fs.Save(homeDir)

	params := []string{"-f", fs.PathFor("expect-config")}
	if globalOpts.Debug {
		params = append(params, "-d")
		log.Println("expect", params)
	}
	cmd := exec.Command("expect", params...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()

	return nil
}
