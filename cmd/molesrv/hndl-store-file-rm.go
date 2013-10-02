package main

import (
	"net/http"
	"os"
	"path"
)

func init() {
	addHandler(handler{
		pattern: "/store/",
		method:  "DELETE",
		fn:      rmFile,
		auth:    true,
		ro:      false,
	})
}

func rmFile(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		defer listCacheLock.Unlock()
		listCacheLock.Lock()
		listCache = nil
	}()

	iniFile := path.Join(storeDir, "data", req.URL.Path[7:])
	if err := os.Rename(iniFile, iniFile+".deleted"); err != nil {
		rw.WriteHeader(404)
	}
}
