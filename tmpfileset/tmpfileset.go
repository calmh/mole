package tmpfileset

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

var (
	ErrAlreadySaved = errors.New("fileset is already saved")
	ErrNotSaved     = errors.New("fileset is not saved")
)

type FileSet struct {
	chunks map[string][]byte
	names  map[string]string
}

func (fs *FileSet) lazyInit() {
	if fs.chunks == nil {
		fs.chunks = make(map[string][]byte)
	}
}

func (fs *FileSet) Add(id string, content []byte) {
	fs.lazyInit()
	fs.chunks[id] = content
}

func (fs *FileSet) Save(dir string) error {
	if fs.names != nil {
		return ErrAlreadySaved
	}

	if fs.chunks == nil {
		return nil
	}

	tfiles := make(map[string]*os.File)
	fs.names = make(map[string]string)

	for id, _ := range fs.chunks {
		tfile, err := ioutil.TempFile(dir, id+".")
		if err != nil {
			return err
		}

		fs.names[id] = tfile.Name()
		for tid, content := range fs.chunks {
			fs.chunks[tid] = bytes.Replace(content, []byte("{"+id+"}"), []byte(fs.names[id]), -1)
		}

		tfiles[id] = tfile
	}

	for id, content := range fs.chunks {
		_, err := tfiles[id].Write(content)
		if err != nil {
			// TODO: erase those already created
			return err
		}
		tfiles[id].Close()
	}

	return nil
}

func (fs *FileSet) PathFor(id string) string {
	if fs.names == nil {
		return ""
	}
	return fs.names[id]
}

func (fs *FileSet) ContentFor(id string) []byte {
	if fs.chunks == nil {
		return nil
	}
	return fs.chunks[id]
}

func (fs *FileSet) Remove() (err error) {
	if fs.names == nil {
		return ErrNotSaved
	}

	for _, path := range fs.names {
		if e := os.Remove(path); e != nil && err == nil {
			err = e
		}
	}

	fs.names = nil
	return
}

func (fs FileSet) String() string {
	var res []string

	for id, content := range fs.chunks {
		res = append(res, "+++ "+id+" +++\n")
		res = append(res, string(content)+"\n")
	}

	return strings.Join(res, "\n")
}
