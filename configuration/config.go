package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/calmh/mole/configuration/configparser"
)

type Config struct {
	General struct {
		Description string
		Author      string
		Main        string
		Version     int
		Other       map[string]string
	}

	Hosts    map[string]Host
	Forwards map[string]Forward
}

type Host struct {
	Addr  string
	User  string
	Key   string
	Pass  string
	Other map[string]string
}

type Forward struct {
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

func Load(fname string) (*Config, error) {
	c := Config{}
	c.General.Other = make(map[string]string)
	c.Hosts = make(map[string]Host)
	c.Forwards = make(map[string]Forward)

	rc, e := configparser.Read(fname)
	if e != nil {
		return nil, e
	}

	// Tedious parsing is tedious

	sec, e := rc.Section("general")
	if e != nil {
		return nil, e
	}

	for k, v := range sec.Options() {
		switch k {
		case "description":
			c.General.Description = v
		case "author":
			c.General.Author = v
		case "main":
			c.General.Main = v
		case "version":
			_, e := fmt.Sscan(v, &c.General.Version)
			if e != nil {
				return nil, e
			}
		default:
			c.General.Other[k] = v
		}
	}

	secs, e := rc.Find("hosts\\.")
	if e != nil {
		return nil, e
	}
	for _, sec := range secs {
		name := sec.Name()[6:]
		host := Host{}
		host.Other = make(map[string]string)
		for k, v := range sec.Options() {
			switch k {
			case "addr":
				host.Addr = v
			case "key":
				host.Key = v
			case "user":
				host.User = v
			case "password":
				host.Pass = v
			default:
				host.Other[k] = v
			}
		}
		c.Hosts[name] = host
	}

	secs, e = rc.Find("forwards\\.")
	if e != nil {
		return nil, e
	}
	for _, sec := range secs {
		name := sec.Name()[9:]
		forw := Forward{}
		forw.Other = make(map[string]string)

		var srcfs, dstfs, srcps []string
		var srcport, dstport, repeat int
		for k, v := range sec.Options() {
			srcfs = strings.SplitN(k, ":", 2)
			srcps = strings.SplitN(srcfs[1], "-", 2)
			srcport, _ = strconv.Atoi(srcps[0])
			if len(srcps) == 2 {
				ep, _ := strconv.Atoi(srcps[1])
				repeat = ep - srcport
			} else {
				repeat = 0
			}

			dstfs = strings.SplitN(v, ":", 2)
			if len(dstfs) == 2 {
				dstport, _ = strconv.Atoi(dstfs[1])
			} else {
				dstport = srcport
			}

			l := ForwardLine{
				SrcIP:   srcfs[0],
				SrcPort: srcport,
				DstIP:   dstfs[0],
				DstPort: dstport,
				Repeat:  repeat,
			}

			forw.Lines = append(forw.Lines, l)
		}
		c.Forwards[name] = forw
	}
	return &c, nil
}
