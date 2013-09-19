package conf

import (
	"fmt"
	"github.com/calmh/mole/ini"
	"strconv"
	"strings"
)

func parse(i ini.File) (cp *Config, err error) {
	c := Config{}
	c.General.Other = make(map[string]string)
	c.HostsMap = make(map[string]int)

	var hostID int
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
		} else if strings.HasPrefix(section, "hosts.") {
			name := section[6:]
			host := Host{
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
				default:
					host.Other[k] = v
				}
			}
			c.Hosts = append(c.Hosts, host)
			c.HostsMap[name] = hostID
			hostID++
		} else if strings.HasPrefix(section, "forwards.") {
			name := section[9:]
			forw := Forward{Name: name}
			forw.Other = make(map[string]string)

			var srcfs, dstfs, srcps []string
			var srcport, dstport, repeat int
			for k, v := range options {
				if c.General.Version >= 320 {
					if k == "comment" {
						forw.Comment = v
						continue
					}
				}

				srcfs = strings.SplitN(k, ":", 2)
				if len(srcfs) != 2 {
					err = fmt.Errorf("incorrect forward source %q", k)
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
		} else if section == "openconnect" {
			c.OpenConnect = options
		} else if section == "vpnc" {
			c.Vpnc = options
		} else if section == "vpn routes" {
			for net, mask := range options {
				c.VpnRoutes = append(c.VpnRoutes, net+"/"+mask)
			}
		}
	}

	cp = &c
	return
}
