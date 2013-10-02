package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type intervalWriter struct {
	name     string
	interval int
	cur      struct {
		file io.WriteCloser
		name string
		sec  int64
	}
	sync.Mutex
}

func (i *intervalWriter) Write(b []byte) (n int, err error) {
	defer i.Unlock()
	i.Lock()

	curSec := (time.Now().Unix() / int64(i.interval)) * int64(i.interval)
	if curSec != i.cur.sec {
		i.close()
	}

	if i.cur.file == nil {
		i.cur.sec = curSec
		i.cur.name = fmt.Sprintf("%s.%d", i.name, i.cur.sec)
		i.cur.file, err = os.OpenFile(i.cur.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return
		}
	}

	return i.cur.file.Write(b)
}

func (i *intervalWriter) Close() error {
	defer i.Unlock()
	i.Lock()
	return i.close()
}

func (i *intervalWriter) close() (err error) {
	if i.cur.file == nil {
		return nil
	}

	err = i.cur.file.Close()
	i.cur.file = nil
	return
}
