package conf

import (
	"fmt"
	"github.com/calmh/mole/ini"
	"regexp"
	"strconv"
	"strings"
)

var ipRe = regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)

func parse(i ini.File) (cp *Config, err error) {
	c := Config{}

	for _, section := range i.Sections() {
		options := i.OptionMap(section)

		if section == "general" {
			err := parseGeneral(&c, options)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(section, "hosts.") {
			err := parseHost(&c, section, options)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(section, "forwards.") {
			err := parseForward(&c, section, options)
			if err != nil {
				return nil, err
			}
		} else if section == "openconnect" {
			c.OpenConnect = getKVs(i, section)
		} else if section == "vpnc" {
			c.Vpnc = getKVs(i, section)
		} else if section == "vpn routes" {
			nets := i.Options(section)
			c.VpnRoutes = make([]string, len(nets))
			for j, net := range nets {
				c.VpnRoutes[j] = net + "/" + i.Get(section, net)
			}
		}
	}

	// Check for existence of either hosts or forwards
	if len(c.Hosts) == 0 && len(c.Forwards) == 0 {
		err = fmt.Errorf(`must exist either "hosts" or "forwards" section`)
		return
	}

	// Check for nonexistant "main"
	if c.General.Main != "" {
		if h := c.GetHost(c.General.Main); h == nil {
			err = fmt.Errorf(`"main" refers to nonexistent host %q`, c.General.Main)
			return
		}
	}

	for _, host := range c.Hosts {
		// Check for errors in "via" links
		if host.Via != "" {
			if h := c.GetHost(host.Via); h == nil {
				err = fmt.Errorf(`host %q "via" refers to nonexistent host %q`, host.Name, host.Via)
				return
			}
		}
	}

	seenSources := map[string]bool{}
	for _, fwd := range c.Forwards {
		for _, line := range fwd.Lines {
			// Check for duplicate forwards
			for i := 0; i <= line.Repeat; i++ {
				src := line.SrcString(i)
				if seenSources[src] {
					err = fmt.Errorf("duplicate forward source %q", src)
					return
				}
				seenSources[src] = true
			}

			// Check for privileged ports
			if line.SrcPort < 1024 {
				err = fmt.Errorf("privileged source port %d in forward source %q", line.SrcPort, line.SrcString(0))
				return
			}
		}
	}

	cp = &c
	return
}

func parseGeneral(c *Config, options map[string]string) (err error) {
	for _, field := range []string{"version", "description", "author"} {
		if _, ok := options[field]; !ok {
			return fmt.Errorf("missing required field %q in general section", field)
		}
	}

	for k, v := range options {
		switch k {
		case "description":
			c.General.Description = v
		case "author":
			c.General.Author = v
		case "main":
			c.General.Main = v
		case "version":
			var f float64
			_, err = fmt.Sscan(v, &f)
			if err != nil {
				return
			}
			c.General.Version = int(100 * f)
		default:
			c.General.Other = append(c.General.Other, KeyValue{k, v})
		}
	}

	if c.General.Version < 400 {
		for _, kv := range c.General.Other {
			return fmt.Errorf("unrecognized field %q in section general not permitted by config version %d", kv.Key, c.General.Version)
		}
	}

	return
}

func parseHost(c *Config, section string, options map[string]string) (err error) {
	name := section[6:]

	for _, field := range []string{"addr", "user"} {
		if _, ok := options[field]; !ok {
			return fmt.Errorf("missing required field %q on host %q", field, name)
		}
	}

	if _, ok1 := options["password"]; !ok1 {
		if _, ok2 := options["key"]; !ok2 {
			return fmt.Errorf(`missing required field "password" or "key" on host %q`, name)
		}
	}

	host := Host{
		Name: name,
		Port: defaultSSHPort,
	}

	for k, v := range options {
		switch k {
		case "addr":
			host.Addr = v
		case "port":
			p, e := strconv.Atoi(v)
			if e != nil {
				return e
			}
			host.Port = p
		case "key":
			host.Key = strings.Replace(v, "\\n", "\n", -1)
		case "user":
			host.User = v
		case "password":
			host.Pass = v
		case "via":
			host.Via = v
		case "socks":
			host.SOCKS = v
		case "prompt":
			// legacy, ignored
		default:
			host.Other = append(host.Other, KeyValue{k, v})
		}
	}
	c.Hosts = append(c.Hosts, host)

	if c.General.Version < 400 {
		for _, kv := range host.Other {
			return fmt.Errorf("unrecognized field %q on host %q not permitted by config version %d", kv.Key, host.Name, c.General.Version)
		}
	}

	if host.SOCKS != "" && host.Via != "" {
		return fmt.Errorf(`cannot combine fields "socks" and "via" for host %q`, host.Name)
	}

	return
}

func parseForward(c *Config, section string, options map[string]string) (err error) {
	name := section[9:]
	forw := Forward{Name: name}

	var srcfs, dstfs, srcps []string
	var srcport, dstport, repeat int
	for k, v := range options {
		if k == "comment" {
			if c.General.Version < 320 {
				err = fmt.Errorf("forward comments are supported in config version 3.2 and above")
				return
			}
			forw.Comment = v
			continue
		}

		srcfs = strings.SplitN(k, ":", 2)
		if len(srcfs) != 2 || len(srcfs[0]) == 0 || len(srcfs[1]) == 0 {
			err = fmt.Errorf("malformed forward source %q", k)
			return
		}

		if !ipRe.MatchString(srcfs[0]) {
			err = fmt.Errorf("malformed forward source %q", k)
			return
		}

		srcps = strings.SplitN(srcfs[1], "-", 2)
		srcport, _ = strconv.Atoi(srcps[0])
		if len(srcps) == 2 {
			ep, _ := strconv.Atoi(srcps[1])
			repeat = ep - srcport
		} else {
			repeat = 0
		}

		dstfs = strings.SplitN(v, ":", 2)

		if !ipRe.MatchString(dstfs[0]) {
			err = fmt.Errorf("malformed forward destination %q", v)
			return
		}

		if len(dstfs) == 2 {
			if repeat > 0 {
				err = fmt.Errorf("malformed forward destination %q (port range)", v)
				return
			}
			if len(dstfs[0]) == 0 || len(dstfs[1]) == 0 {
				err = fmt.Errorf("malformed forward destination %q", v)
				return
			}
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
	c.Forwards = append(c.Forwards, forw)
	return
}

func getKVs(i ini.File, section string) []KeyValue {
	options := i.Options(section)
	res := make([]KeyValue, len(options))
	for j, option := range options {
		value := i.Get(section, option)
		res[j] = KeyValue{option, value}
	}
	return res
}
