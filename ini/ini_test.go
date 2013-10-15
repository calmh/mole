package ini_test

import (
	"bytes"
	"github.com/calmh/mole/ini"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	strs := []string{
		";pre comment",
		"[general]",
		"foo=bar",                     // Standard case
		"baz    = foo quux ",          // Space around equal sign and after value is ignored
		"; comment",                   // Comments are ignored, can start at column 1 only
		`ws= "with some spaces  " `,   // Spaces are significant inside quotes
		`wn=with\nnewline`,            // \n is interpreted as newline
		" baz2 = 32 ",                 // Spaces arund the value are ignored too
		";   The last comment  ",      // Spaces around comments are trimmed too
		`quoted1=a "quoted" word`,     // Quoted words are fine
		`quoted2="a \"quoted\" word"`, // Same same
	}
	buf := bytes.NewBufferString(strings.Join(strs, "\n"))
	inf := ini.Parse(buf)

	if l := len(inf.Sections()); l != 1 {
		t.Errorf("incorrect #sections %d", l)
	}

	correct := map[string]string{
		"foo":     "bar",
		"baz":     "foo quux",
		"baz2":    "32",
		"ws":      "with some spaces  ",
		"wn":      "with\nnewline",
		"quoted1": "a \"quoted\" word",
		"quoted2": "a \"quoted\" word",
		"other":   "",
	}

	for k, v := range correct {
		if v2 := inf.Get("general", k); v2 != v {
			t.Errorf("incorrect general.%s, %q != %q", k, v2, v)
		}
	}

	if opts := inf.Options("general"); len(opts) != len(correct)-1 {
		t.Errorf("incorrect #options %d", len(opts))
	} else {
		correct := []string{"foo", "baz", "ws", "wn", "baz2"}
		for i := range correct {
			if opts[i] != correct[i] {
				t.Errorf("incorrect option #%d, %q != %q", i, opts[i], correct[i])
			}
		}
	}

	if cmts := inf.Comments(""); len(cmts) != 1 {
		t.Errorf("incorrect #comments %d", len(cmts))
	} else {
		correct := []string{"pre comment"}
		for i := range correct {
			if cmts[i] != correct[i] {
				t.Errorf("incorrect comments #%d, %q != %q", i, cmts[i], correct[i])
			}
		}
	}
	if cmts := inf.Comments("general"); len(cmts) != 2 {
		t.Errorf("incorrect #comments %d", len(cmts))
	} else {
		correct := []string{"comment", "The last comment"}
		for i := range correct {
			if cmts[i] != correct[i] {
				t.Errorf("incorrect comments #%d, %q != %q", i, cmts[i], correct[i])
			}
		}
	}
}

func TestWrite(t *testing.T) {
	strs := []string{
		";pre comment",
		"[general]",
		"foo=bar",                   // Standard case
		"baz    = foo quux ",        // Space around equal sign and after value is ignored
		"; comment",                 // Comments are ignored, can start at column 1 only
		`ws= "with some spaces  " `, // Spaces are significant inside quotes
		`wn=with\nnewline`,          // \n is interpreted as newline
		" baz2 = 32 ",               // Spaces arund the value are ignored too
		";   The last comment  ",    // Spaces around comments are trimmed too
	}
	buf := bytes.NewBufferString(strings.Join(strs, "\n"))
	inf := ini.Parse(buf)

	var out bytes.Buffer
	inf.Write(&out)

	correct := `; pre comment

[general]
; comment
; The last comment
foo=bar
baz=foo quux
ws="with some spaces  "
wn="with\nnewline"
baz2=32

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
foo=baz
foo2=bar2
baz=quux

[other]
baz2=quux2

`

	if s := out.String(); s != correct {
		t.Errorf("incorrect INI after set:\n%s", s)
	}
}
