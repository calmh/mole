package main

import (
	"encoding/json"
	"net/http"
	"path"
	"time"
)

type auditRecord struct {
	Time      time.Time `json:"time"`
	Client    string    `json:"client"`
	UserAgent string    `json:"ua"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	User      string    `json:"user"`
	Comment   string    `json:"comment"`
}

var iv *intervalWriter

func audit(req *http.Request, comment string) {
	if iv == nil {
		iv = &intervalWriter{name: path.Join(storeDir, auditFile), interval: int(auditIntv.Seconds())}
	}
	rec := auditRecord{
		Time:      time.Now(),
		Client:    req.RemoteAddr,
		UserAgent: req.Header.Get("User-Agent"),
		Method:    req.Method,
		Path:      req.URL.Path,
		User:      req.Header.Get("X-Mole-Authenticated"),
		Comment:   comment,
	}
	bs, _ := json.Marshal(rec)
	bs = append(bs, '\n')
	_, err := iv.Write(bs)
	if err != nil {
		panic(err)
	}
}
