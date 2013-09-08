package main

import (
	"github.com/jessevdk/go-flags"
)

var globalOpts struct {
	Foo string `long:"foo"`
}
var globalParser = flags.NewParser(&globalOpts, flags.PassDoubleDash|flags.HelpFlag|flags.PrintErrors)

func main() {
	_, e := globalParser.Parse()
	if e != nil {
		return
	}
	println("foo")
}
