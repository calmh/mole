package main

import (
	"fmt"
	"strings"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/tmpfileset"
)

func sshConfig(cfg *configuration.Config, fs *tmpfileset.FileSet) {
	var lines []string
	nl := func(line string) {
		lines = append(lines, line)
	}

	for name, host := range cfg.Hosts {
		nl("Host " + name)
		nl("  HostName " + host.Addr)
		nl(fmt.Sprintf("  Port %d", host.Port))
		nl("  User " + host.User)

		if host.Via != "" {
			nl("  ProxyCommand ssh -F {ssh-config} " + host.Via + " nc -W 1800 %h %p")
		}

		nl("  KeychainIntegration no")
		if host.Pass != "" {
			nl("  # password")
			nl("  PasswordAuthentication yes")
			nl("  KbdInteractiveAuthentication yes")
			nl("  PubkeyAuthentication no")
			nl("  RSAAuthentication no")
		} else if host.Key != "" {
			fs.Add("identity-"+host.Unique, []byte(host.Key))
			nl("  # ssh key")
			nl("  IdentityFile {identity-" + host.Unique + "}")
			nl("  PasswordAuthentication no")
			nl("  KbdInteractiveAuthentication no")
			nl("  PubkeyAuthentication yes")
			nl("  RSAAuthentication yes")
		}

		nl("  Compression yes")
		nl("  # 60 second keepalive timeout")
		nl("  ServerAliveInterval 20")
		nl("  ServerAliveCountMax 3")

		if name == cfg.General.Main {
			for name, fwd := range cfg.Forwards {
				nl(fmt.Sprintf("  # forward %q", name))
				for _, l := range fwd.Lines {
					for i := 0; i <= l.Repeat; i++ {
						nl(fmt.Sprintf("  LocalForward %s:%d %s:%d", l.SrcIP, l.SrcPort+i, l.DstIP, l.DstPort+i))
					}
				}
			}
		}
	}

	content := []byte(strings.Join(lines, "\n") + "\n")
	fs.Add("ssh-config", content)
}
