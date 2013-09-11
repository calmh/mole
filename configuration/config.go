package configuration

import (
	"bytes"
	"os"
	"sort"

	"github.com/calmh/mole/ini"
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
	return parse(ini.Parse(f))
}

func LoadString(data string) (*Config, error) {
	f := bytes.NewBufferString(data)
	return parse(ini.Parse(f))
}

func (c *Config) SourceAddresses() []string {
	addrMap := make(map[string]bool)
	for _, fwd := range c.Forwards {
		for _, line := range fwd.Lines {
			addrMap[line.SrcIP] = true
		}
	}

	addrs := make([]string, 0, len(addrMap))
	for ip := range addrMap {
		addrs = append(addrs, ip)
	}

	sort.Strings(addrs)
	return addrs
}
