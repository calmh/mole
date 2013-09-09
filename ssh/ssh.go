package ssh

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/calmh/mole/configuration"
)

type Config struct {
	Lines []string
	Files []string
}

func New(cfg *configuration.Config) *Config {
	s := Config{}

	for name, host := range cfg.Hosts {
		s.push("Host " + name)
		s.push("  HostName " + host.Addr)
		s.push(fmt.Sprintf("  Port %d", host.Port))
		s.push("  User " + host.User)

		if host.Via != "" {
			s.push("  ProxyCommand ssh -F /dev/null " + host.Via + " nc -W 1800 %h %p")
		}

		s.push("  KeychainIntegration no")
		if host.Pass != "" {
			s.push("  # password")
			s.push("  PasswordAuthentication yes")
			s.push("  KbdInteractiveAuthentication yes")
			s.push("  PubkeyAuthentication no")
			s.push("  RSAAuthentication no")
		} else if host.Key != "" {
			ifile := identityFile(host.Key)
			s.Files = append(s.Files, ifile)
			s.push("  # ssh key")
			s.push("  IdentityFile " + ifile)
			s.push("  PasswordAuthentication no")
			s.push("  KbdInteractiveAuthentication no")
			s.push("  PubkeyAuthentication yes")
			s.push("  RSAAuthentication yes")
		}

		s.push("  Compression yes")
		s.push("  # 60 second keepalive timeout")
		s.push("  ServerAliveInterval 20")
		s.push("  ServerAliveCountMax 3")

		if name == cfg.General.Main {
			for name, fwd := range cfg.Forwards {
				s.push(fmt.Sprintf("  # forward %q", name))
				for _, l := range fwd.Lines {
					for i := 0; i <= l.Repeat; i++ {
						s.push(fmt.Sprintf("  LocalForward %s:%d %s:%d", l.SrcIP, l.SrcPort+i, l.DstIP, l.DstPort+i))
					}
				}
			}
		}
	}

	return &s
}

func (c *Config) push(line string) {
	c.Lines = append(c.Lines, line)
}

func (c *Config) String() string {
	return strings.Join(c.Lines, "\n")
}

func identityFile(key string) string {
	f, e := ioutil.TempFile("", "identity")
	if e != nil {
		log.Fatal(e)
	}

	_, e = f.WriteString(key)
	if e != nil {
		log.Fatal(e)
	}
	f.Close()
	return f.Name()
}
