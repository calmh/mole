package main

type parse struct{}

func init() {
	p := &parse{}
	globalParser.AddCommand("parse", "Parse a tunnel definition", "Parse parses a tunnel configuration file and displays the parse tree and so on and so on", p)
}

func (*parse) Execute(args []string) error {
	println("eh")
	return nil
}
