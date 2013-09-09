package expect

import (
	"fmt"
	"log"
	"strings"

	"github.com/calmh/mole/configuration"
)

type Script struct {
	Lines []string
}

func New(cfg *configuration.Config) *Script {
	if cfg.General.Main == "" {
		log.Fatal("Cannot generate expect script for empty destination host")
	}

	_, ok := cfg.Hosts[cfg.General.Main]
	if !ok {
		log.Fatalf("Cannot generate expect script for non-existent host %q", cfg.General.Main)
	}

	var s Script

	s.push("set timeout 30")
	s.push("spawn ssh -F /dev/null " + cfg.General.Main)

	s.push("expect {")
	for name, host := range cfg.Hosts {
		if host.User != "" && host.Pass != "" {
			s.push("  # " + name)
			s.push(fmt.Sprintf(`  "%s@%s" {`, host.User, host.Addr))
			s.push(fmt.Sprintf(`    send "$env(%s_pass)\n";`, host.Unique))
			s.push("    exp_continue;")
			s.push("  }")

			if name == cfg.General.Main {
				s.push("  # " + name + " (as main)")
				s.push(`  "Password:" {`)
				s.push(fmt.Sprintf(`    send "$env(%s_pass)\n";`, host.Unique))
				s.push("    exp_continue;")
				s.push("  }")
			}
		} else {
			s.push("  # " + name + " does not need password authentication")
		}
		s.push("")
	}

	s.push("  # prompt")
	s.push(fmt.Sprintf("  -re %q {", cfg.Hosts[cfg.General.Main].Prompt))
	s.push(`    send_user "\nThe login sequence seems to have worked.\n\n";`)
	s.push(`    send "\r";`)
	s.push("    interact;")
	s.push("  }")
	s.push("")

	s.push(`  "Permission denied" {`)
	s.push(`    send_user "\nPermission denied, failed to set up tunneling.\n\n";`)
	s.push(`    exit 2;`)
	s.push("  }")
	s.push("")

	s.push(`  timeout {`)
	s.push(`    send_user "\nUnknown error, failed to set up tunneling.\n\n";`)
	s.push(`    exit 2;`)
	s.push("  }")

	s.push("}")

	s.push("catch wait reason;")
	s.push("exit [lindex $reason 3];")

	return &s
}

func (s *Script) push(c string) {
	s.Lines = append(s.Lines, c)
}

func (s *Script) String() string {
	return "#!/usr/bin/env expect -f\n\n" + strings.Join(s.Lines, "\n")
}
