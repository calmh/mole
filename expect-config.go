package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/tmpfileset"
)

func expectConfig(cfg *configuration.Config, fs *tmpfileset.FileSet) {
	if cfg.General.Main == "" {
		log.Fatal("Cannot generate expect script for empty destination host")
	}

	_, ok := cfg.HostsMap[cfg.General.Main]
	if !ok {
		log.Fatalf("Cannot generate expect script for non-existent host %q", cfg.General.Main)
	}

	var lines []string
	nl := func(line string) {
		lines = append(lines, line)
	}

	nl("set timeout 30")
	if globalOpts.Debug {
		nl("spawn ssh -v -F {ssh-config} " + cfg.General.Main)
	} else {
		nl("spawn ssh -F {ssh-config} " + cfg.General.Main)
	}

	nl("expect {")
	for _, host := range cfg.Hosts {
		if host.User != "" && host.Pass != "" {
			nl("  # " + host.Name)
			nl(fmt.Sprintf(`  "%s@%s" {`, host.User, host.Addr))
			nl(fmt.Sprintf(`    send %q; send "\n";`, host.Pass))
			nl("    exp_continue;")
			nl("  }")

			if host.Name == cfg.General.Main {
				nl("  # " + host.Name + " (as main)")
				nl(`  "Password:" {`)
				nl(fmt.Sprintf(`    send %q; send "\n";`, host.Pass))
				nl("    exp_continue;")
				nl("  }")
			}
		} else {
			nl("  # " + host.Name + " does not need password authentication")
		}
		nl("")
	}

	nl("  # prompt")
	nl(fmt.Sprintf("  -re %q {", cfg.Hosts[cfg.HostsMap[cfg.General.Main]].Prompt))
	nl(`    send_user "\nThe login sequence seems to have worked.\n\n";`)
	nl(`    send "\r";`)

	for _, fwd := range cfg.Forwards {
		nl(`    send_user "` + fwd.Name + `:\n";`)
		for _, line := range fwd.Lines {
			if line.Repeat > 0 {
				nl(fmt.Sprintf("    send_user \"    %s %d-%d -> %s %d-%d\\n\";", line.SrcIP, line.SrcPort, line.SrcPort+line.Repeat, line.DstIP, line.DstPort, line.DstPort+line.Repeat))
			} else {
				nl(fmt.Sprintf("    send_user \"    %s %d -> %s %d\\n\";", line.SrcIP, line.SrcPort, line.DstIP, line.DstPort))
			}
		}
		nl(`    send_user "\n";`)
	}

	nl("    interact;")
	nl("  }")
	nl("")

	nl(`  "Permission denied" {`)
	nl(`    send_user "\nPermission denied, failed to set up tunneling.\n\n";`)
	nl(`    exit 2;`)
	nl("  }")
	nl("")

	nl(`  timeout {`)
	nl(`    send_user "\nUnknown error, failed to set up tunneling.\n\n";`)
	nl(`    exit 2;`)
	nl("  }")

	nl("}")

	nl("catch wait reason;")
	nl("exit [lindex $reason 3];")

	content := []byte(strings.Join(lines, "\n") + "\n")
	fs.Add("expect-config", content)
}
