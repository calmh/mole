package configuration

import (
	"bytes"
	"os"
)

const (
	DefaultPrompt = `(%|\$|#|>)\s*$`
	DefaultPort   = 22
)

type Config struct {
	General struct {
		Description string
		Author      string
		Main        string
		Version     int
		Other       map[string]string
	}

	Hosts    []Host
	Forwards []Forward
	HostsMap map[string]int
}

type Host struct {
	Name   string
	Addr   string
	Port   int
	User   string
	Key    string
	Pass   string
	Prompt string
	Via    string
	Unique string
	Other  map[string]string
}

type Forward struct {
	Name  string
	Lines []ForwardLine
	Other map[string]string
}

type ForwardLine struct {
	SrcIP   string
	SrcPort int
	DstIP   string
	DstPort int
	Repeat  int
}

func LoadFile(fname string) (*Config, error) {
	f, e := os.Open(fname)
	if e != nil {
		return nil, e
	}
	return parse(tokenize(f))
}

func LoadString(data string) (*Config, error) {
	f := bytes.NewBufferString(data)
	return parse(tokenize(f))
}
