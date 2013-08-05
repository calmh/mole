package configuration

import (
	"fmt"
	"github.com/alyu/configparser"
	"log"
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
	Other map[string]string
}

type Forward struct {
}

func Load(fname string) Config {
	c := Config{}

	rc, e := configparser.Read(fname)
	if e != nil {
		log.Fatal(e)
	}

	// Tedious parsing is tedious

	sec, e := rc.Section("general")
	if e != nil {
		log.Fatal(e)
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
				log.Fatal(e)
			}
		default:
			if c.General.Other == nil {
				c.General.Other = make(map[string]string)
			}
			c.General.Other[k] = v
		}
	}

	return c
}
