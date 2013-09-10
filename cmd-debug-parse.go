// +build debug

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/calmh/mole/configuration"
	"github.com/jessevdk/go-flags"
)

type cmdParse struct{}

var parseParser *flags.Parser

func init() {
	cmd := cmdParse{}
	parseParser = globalParser.AddCommand("debug-parse", "Show tunnel definition", "Parse parses a tunnel configuration file and displays the internal object in JSON format", &cmd)
}

func (c *cmdParse) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		parseParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("debug-parse: missing required option <filename>\n")
	}

	cfg, err := configuration.LoadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	bs, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	json.Indent(&buf, bs, "", "    ")
	log.Printf("%s\n", buf.Bytes())

	return nil
}
