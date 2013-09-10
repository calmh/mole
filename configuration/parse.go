package configuration

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type iniSection map[string]string
type iniFile map[string]iniSection

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^\s*([^\s]+)\s*=\s*(.+)\s*$`)
)

func tokenize(stream io.Reader) iniFile {
	iniFile := make(iniFile)
	var curSection string
	scanner := bufio.NewScanner(bufio.NewReader(stream))
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")) && len(line) > 0 {
			if m := iniSectionRe.FindStringSubmatch(line); len(m) > 0 {
				curSection = m[1]
				iniFile[curSection] = make(iniSection)
			} else if m := iniOptionRe.FindStringSubmatch(line); len(m) > 0 {
				iniFile[curSection][m[1]] = m[2]
			}
		}
	}
	return iniFile
}

func parse(ini iniFile) (*Config, error) {
	c := Config{}
	c.General.Other = make(map[string]string)
	c.HostsMap = make(map[string]int)

	var hostId int
	for section, options := range ini {
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
				v = strings.Trim(v, `"`)
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
		}
	}
	return &c, nil
}
