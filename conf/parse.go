package conf

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/calmh/ini"
)

var ipRe = regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)

func parse(ic ini.Config) (cp *Config, err error) {
	c := Config{}
	c.General.Other = make(map[string]string)
	c.HostsMap = make(map[string]int)

	c.Comments = ic.Comments("")

	for _, section := range ic.Sections() {
		options := ic.OptionMap(section)

		if section == "general" {
			err := parseGeneral(&c, options)
			if err != nil {
				return nil, err
			}
			c.General.Comments = ic.Comments(section)
		} else if strings.HasPrefix(section, "hosts.") {
			host, err := parseHost(ic, section)
			if err != nil {
				return nil, err
			}
			if c.General.Version < 400 {
				for k := range host.Other {
					return nil, fmt.Errorf("unrecognized field %q on host %q not permitted by config version %d", k, host.Name, c.General.Version)
				}
			}

			c.Hosts = append(c.Hosts, host)
			c.HostsMap[host.Name] = len(c.Hosts) - 1
		} else if strings.HasPrefix(section, "forwards.") {
			forw, err := parseForward(ic, section)
			if err != nil {
				return nil, err
			}
			if cmt := ic.Get(section, "comment"); cmt != "" && c.General.Version < 320 {
				return nil, fmt.Errorf("forward comments are supported in config version 3.2 and above")
			}
			c.Forwards = append(c.Forwards, forw)
		} else if section == "openconnect" {
			c.OpenConnect = options
		} else if section == "vpnc" {
			c.Vpnc = options
		} else if section == "vpn routes" {
			for net, mask := range options {
				c.VpnRoutes = append(c.VpnRoutes, net+"/"+mask)
			}
			sort.Strings(c.VpnRoutes)
		}
	}

	// Check for existence of either hosts or forwards
	if len(c.Hosts) == 0 && len(c.Forwards) == 0 {
		err = fmt.Errorf(`must exist either "hosts" or "forwards" section`)
		return
	}

	// Check for nonexistant "main"
	if c.General.Main != "" {
		if _, ok := c.HostsMap[c.General.Main]; !ok {
			err = fmt.Errorf(`"main" refers to nonexistent host %q`, c.General.Main)
			return
		}
	}

	for _, host := range c.Hosts {
		// Check for errors in "via" links
		if host.Via != "" {
			if _, ok := c.HostsMap[host.Via]; !ok {
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
			c.General.Other[k] = v
		}
	}

	if c.General.Version < 400 {
		for k := range c.General.Other {
			return fmt.Errorf("unrecognized field %q in section general not permitted by config version %d", k, c.General.Version)
		}
	}

	return
}

func parseHost(ic ini.Config, section string) (host Host, err error) {
	name := section[6:]
	options := ic.OptionMap(section)

	for _, field := range []string{"addr", "user"} {
		if _, ok := options[field]; !ok {
			err = fmt.Errorf("missing required field %q on host %q", field, name)
			return
		}
	}

	if _, ok1 := options["password"]; !ok1 {
		if _, ok2 := options["key"]; !ok2 {
			err = fmt.Errorf(`missing required field "password" or "key" on host %q`, name)
			return
		}
	}

	host = Host{
		Name: name,
		Port: defaultSSHPort,
	}
	host.Other = make(map[string]string)
	for k, v := range options {
		switch k {
		case "addr":
			host.Addr = v
		case "port":
			p, e := strconv.Atoi(v)
			if e != nil {
				err = e
				return
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
			host.Other[k] = v
		}
	}

	if host.SOCKS != "" && host.Via != "" {
		err = fmt.Errorf(`cannot combine fields "socks" and "via" for host %q`, host.Name)
		return
	}

	host.Comments = ic.Comments(section)
	return
}

func parseForward(ic ini.Config, section string) (forw Forward, err error) {
	name := section[9:]
	options := ic.OptionMap(section)
	forw = Forward{Name: name}
	forw.Other = make(map[string]string)

	var srcfs, dstfs, srcps []string
	var srcport, dstport, repeat int
	for k, v := range options {
		if k == "comment" {
			forw.Comments = append(forw.Comments, strings.Split(v, "\n")...)
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

	forw.Comments = append(forw.Comments, ic.Comments(section)...)
	return
}
