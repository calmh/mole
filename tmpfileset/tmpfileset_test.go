package tmpfileset_test

import (
	"nym.se/mole/tmpfileset"
	"testing"
)

func TestFileSet(t *testing.T) {
	var fs tmpfileset.FileSet
	fs.Add("t1", []byte("some t1 content referring to {t2}"))
	fs.Add("t2", []byte("this is t2, referred to by {t1}"))
	e := fs.Save("/tmp")
	if e != nil {
		t.Error(e)
	}

	e = fs.Save("/tmp")
	if e != tmpfileset.ErrAlreadySaved {
		t.Errorf("expected ErrAlreadySaved")
	}

	e = fs.Remove()
	if e != nil {
		t.Error(e)
	}

	e = fs.Remove()
	if e != tmpfileset.ErrNotSaved {
		t.Errorf("expected ErrNotSaved")
	}
}
