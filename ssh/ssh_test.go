package ssh_test

import (
	"fmt"
	"testing"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/ssh"
)

func TestNew(t *testing.T) {
	cfg := configuration.Config{}

	cfg.General.Main = "foo"
	cfg.Hosts = map[string]configuration.Host{
		"foo": {
			Addr:   "1.2.3.4",
			Port:   22,
			User:   "username",
			Pass:   "password",
			Prompt: configuration.DefaultPrompt,
			Via:    "other",
			Unique: "host1",
		},
	}

	s := ssh.New(&cfg)

	if s == nil {
		t.Errorf("got nil ssh.Config")
	}

	fmt.Println(s)
}
