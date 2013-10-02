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

	tun := req.URL.Path[7:]
	iniFile := path.Join(storeDir, "data", tun)
	if err := os.Rename(iniFile, iniFile+".deleted"); err != nil {
		rw.WriteHeader(404)
	}

	if !disableGit {
		// Commit
		dir := path.Join(storeDir, "data")
		user := req.Header.Get("X-Mole-Authenticated")
		commit(dir, "rm "+tun, user)
	}
}
