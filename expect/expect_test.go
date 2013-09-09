package expect_test

import (
	"fmt"
	"testing"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/expect"
)

func TestNew(t *testing.T) {
	cfg := configuration.Config{}

	cfg.General.Main = "foo"
	cfg.Hosts = map[string]configuration.Host{
		"foo": {
			Addr:   "1.2.3.4",
			User:   "username",
			Pass:   "password",
			Prompt: configuration.DefaultPrompt,
			Unique: "host1",
		},
	}

	s := expect.New(&cfg)

	if s == nil {
		t.Errorf("got nil expect.Script")
	}

	if s.Lines[0] != "set timeout 30" {
		t.Errorf("first line should set timeout")
	}

	fmt.Println(s)
}
