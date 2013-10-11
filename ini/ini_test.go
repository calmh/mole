package ini_test

import (
	"bytes"
	"github.com/calmh/mole/ini"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	strs := []string{
		"[general]",
		"foo=bar",
		"baz    = foo quux ",
		"; comment",
		`ws="with some spaces  "`,
		`wn="with\nnewline"`,
		" baz2 = 32 ",
	}
	instr := strings.Join(strs, "\n")
	buf := bytes.NewBufferString(instr)
	inf := ini.Parse(buf)

	if l := len(inf.Sections()); l != 1 {
		t.Errorf("incorrect #sections %d", l)
	}

	correct := map[string]string{
		"foo":   "bar",
		"baz":   "foo quux",
		"baz2":  "32",
		"ws":    "with some spaces  ",
		"wn":    "with\nnewline",
		"other": "",
	}

	for k, v := range correct {
		if v2 := inf.Get("general", k); v2 != v {
			t.Errorf("incorrect general.%s, %q != %q", k, v2, v)
		}
	}

	if opts := inf.Options("general"); len(opts) != 5 {
		t.Errorf("incorrect #options %d", len(opts))
	} else {
		correct := []string{"foo", "baz", "ws", "wn", "baz2"}
		for i := range correct {
			if opts[i] != correct[i] {
				t.Errorf("incorrect option #%d, %q != %q", i, opts[i], correct[i])
			}
		}
	}
}

func TestWrite(t *testing.T) {
	buf := bytes.NewBufferString("[general]\nfoo=bar\nbaz    = foo quux \n;comment\nws=\"with some spaces  \"\n baz2 = 32 \n")
	inf := ini.Parse(buf)

	var out bytes.Buffer
	inf.Write(&out)

	correct := `[general]
foo = bar
baz = foo quux
ws = "with some spaces  "
baz2 = 32

`
	if s := out.String(); s != correct {
		t.Errorf("incorrect written .INI:\n%s\ncorrect:\n%s", s, correct)
	}
}

func TestSet(t *testing.T) {
	buf := bytes.NewBufferString("[general]\nfoo=bar\nfoo2=bar2\n")
	inf := ini.Parse(buf)

	inf.Set("general", "foo", "baz")
	inf.Set("general", "baz", "quux")
	inf.Set("other", "baz2", "quux2")

	var out bytes.Buffer
	inf.Write(&out)

	correct := `[general]
foo = baz
foo2 = bar2
baz = quux

[other]
baz2 = quux2

`

	if s := out.String(); s != correct {
		t.Errorf("incorrect INI after set:\n%s", s)
	}
}
