package configuration

import (
	"bytes"
	"fmt"
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

	Hosts     []Host
	Forwards  []Forward
	HostsMap  map[string]int
	Vpnc      map[string]string
	VpnRoutes []string
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

func (line ForwardLine) SrcString(i int) string {
	if i > line.Repeat {
		panic("index > repeat")
	}
	return fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort+i)
}

func (line ForwardLine) DstString(i int) string {
	if i > line.Repeat {
		panic("index > repeat")
	}
	return fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort+i)
}

func (line ForwardLine) String() string {
	if line.Repeat == 0 {
		src := fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort)
		dst := fmt.Sprintf("%s:%d", line.DstIP, line.DstPort)
		return fmt.Sprintf("%s -> %s", src, dst)
	} else {
		src := fmt.Sprintf("%s:%d-%d", line.SrcIP, line.SrcPort, line.SrcPort+line.Repeat)
		dst := fmt.Sprintf("%s:%d-%d", line.DstIP, line.DstPort, line.DstPort+line.Repeat)
		return fmt.Sprintf("%s -> %s", src, dst)
	}
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

func (c *Config) Remap() {
	port := 10000
	for fi := range c.Forwards {
		for li := range c.Forwards[fi].Lines {
			if c.Forwards[fi].Lines[li].SrcIP != "127.0.0.1" {
				c.Forwards[fi].Lines[li].SrcIP = "127.0.0.1"
				c.Forwards[fi].Lines[li].SrcPort = port
				port += c.Forwards[fi].Lines[li].Repeat + 1
			}
		}
	}
}
