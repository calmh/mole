package conf

import (
	"fmt"
	"strconv"
	"strings"

	"nym.se/mole/ini"
)

func parse(i ini.File) (*Config, error) {
	c := Config{}
	c.General.Other = make(map[string]string)
	c.HostsMap = make(map[string]int)

	var hostId int
	for _, section := range i.SectionNames {
		options := i.Sections[section]

		if section == "general" {
			for k, v := range options {
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
		} else if strings.HasPrefix(section, "hosts.") {
			name := section[6:]
			host := Host{
				Name:   name,
				Port:   DefaultPort,
				Prompt: DefaultPrompt,
				Unique: fmt.Sprintf("host%d", hostId),
			}
			host.Other = make(map[string]string)
			for k, v := range options {
				switch k {
				case "addr":
					host.Addr = v
				case "port":
					p, e := strconv.Atoi(v)
					if e != nil {
						return nil, e
					}
					host.Port = p
				case "key":
					host.Key = strings.Replace(v, "\\n", "\n", -1)
				case "user":
					host.User = v
				case "password":
					host.Pass = v
				case "prompt":
					host.Prompt = v
				case "via":
					host.Via = v
				default:
					host.Other[k] = v
				}
			}
			c.Hosts = append(c.Hosts, host)
			c.HostsMap[name] = hostId
			hostId++
		} else if strings.HasPrefix(section, "forwards.") {
			name := section[9:]
			forw := Forward{Name: name}
			forw.Other = make(map[string]string)

			var srcfs, dstfs, srcps []string
			var srcport, dstport, repeat int
			for k, v := range options {
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
			c.Forwards = append(c.Forwards, forw)
		} else if section == "vpnc" {
			c.Vpnc = options
		} else if section == "vpn routes" {
			for net, mask := range options {
				c.VpnRoutes = append(c.VpnRoutes, net+"/"+mask)
			}
		}
	}
	return &c, nil
}
