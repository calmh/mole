package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/calmh/ini"
	"github.com/calmh/mole/conf"
)

func init() {
	addHandler(handler{
		pattern: "/store/",
		method:  "PUT",
		fn:      putFile,
		auth:    true,
		ro:      false,
	})
}

var obfuscateKeys = []string{
	"key",
	"password",
	"IPSec_secret",
	"Xauth_password",
}

var filenamePattern = regexp.MustCompile(`^[a-z0-9_-]+\.ini$`)

func putFile(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		defer listCacheLock.Unlock()
		listCacheLock.Lock()
		listCache = nil
	}()

	tun := req.URL.Path[7:]
	if !filenamePattern.MatchString(tun) {
		rw.WriteHeader(403)
		rw.Write([]byte("filename not conformant to " + filenamePattern.String()))
		return
	}

	iniFile := path.Join(storeDir, "data", tun)
	// Read pushed data
	data, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	// Verify the configuration
	_, err = conf.Load(bytes.NewBuffer(data))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	// Get the raw INI
	inf := ini.Parse(bytes.NewBuffer(data))

	// Obfuscate
	shouldSaveKeys := false
	for _, section := range inf.Sections() {
		for _, option := range inf.Options(section) {
			for i := range obfuscateKeys {
				if option == obfuscateKeys[i] {
					val := inf.Get(section, option)
					if oval := obfuscate(val); oval != val {
						inf.Set(section, option, oval)
						shouldSaveKeys = true
					}
					break
				}
			}
		}
	}
	if shouldSaveKeys {
		saveKeys()
	}

	// Save
	outf, err := os.Create(iniFile)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	inf.Write(outf)
	outf.Close()

	if !disableGit {
		// Commit
		dir := path.Join(storeDir, "data")
		user := req.Header.Get("X-Mole-Authenticated")
		commit(dir, "push "+tun, user)
	}
}
