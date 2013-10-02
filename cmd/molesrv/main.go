package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type handler struct {
	method  string
	pattern string
	fn      func(http.ResponseWriter, *http.Request)
	auth    bool
	ro      bool
}

var (
	// Pattern to handleFunc
	handlers = map[string][]handler{}

	// CLI options
	listenAddr = ":9443"
	storeDir   = "~/mole-store"
	certFile   = "cert.pem"
	keyFile    = "key.pem"
	auditFile  = "audit.log"
	auditIntv  = 86400 * time.Second
	noAuth     = false
	readOnly   = false
	disableGit = false
)

func addHandler(hnd handler) {
	handlers[hnd.pattern] = append(handlers[hnd.pattern], hnd)
	log.Printf("Added %s handler for %q (auth=%v, ro=%v)", hnd.method, hnd.pattern, hnd.auth, hnd.ro)
}

func main() {
	fs := flag.NewFlagSet("molesrv", flag.ExitOnError)
	fs.Usage = usageFor(fs, "molesrv [options]")
	fs.StringVar(&listenAddr, "listen", listenAddr, "HTTPS listen address")
	fs.StringVar(&storeDir, "store-dir", storeDir, "Mole store directory")
	fs.StringVar(&certFile, "cert-file", certFile, "Certificate file (relative to store directory)")
	fs.StringVar(&keyFile, "key-file", keyFile, "Key file (relative to store directory)")
	fs.StringVar(&auditFile, "audit-file", auditFile, "Audit file (relative to store directory)")
	fs.DurationVar(&auditIntv, "audit-intv", auditIntv, "Audit file creation interval")
	fs.BoolVar(&noAuth, "no-auth", noAuth, "Do not perform authentication")
	fs.BoolVar(&readOnly, "no-write", readOnly, "Disallow writable client operations (push, rm, etc)")
	fs.BoolVar(&disableGit, "no-git", disableGit, "Do not treat the store as a git repository")
	fs.Parse(os.Args[1:])

	if strings.HasPrefix(storeDir, "~/") {
		home := getHomeDir()
		storeDir = strings.Replace(storeDir, "~", home, 1)
	}

	err := loadKeys()
	if err != nil {
		log.Println("Warning: ", err)
	}
	if keys == nil {
		keys = make(map[string]string)
		log.Println("Initialized new key store")
	}

	for pattern, handlerList := range handlers {
		setupHandler(pattern, handlerList)
	}

	err = http.ListenAndServeTLS(listenAddr, path.Join(storeDir, certFile), path.Join(storeDir, keyFile), nil)
	if err != nil {
		log.Println("Error:", err)
	}
}

func setupHandler(p string, hs []handler) {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		for _, h := range hs {
			if h.method != req.Method {
				continue
			}

			if !h.ro && readOnly {
				audit(req, p+"; rejected (ro)")
				rw.WriteHeader(403)
				rw.Write([]byte("Server is in read-only mode"))
				return
			}

			if h.auth && !noAuth {
				if !authenticate(rw, req) {
					audit(req, p+"; rejected")
					rw.WriteHeader(401)
					return
				}
			}

			audit(req, p+"; accepted")
			h.fn(rw, req)
			return
		}

		rw.WriteHeader(405)
	}

	http.HandleFunc(p, fn)
}

func getHomeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		log.Println("Warning: no home directory, using /tmp")
		return "/tmp"
	}
	return home
}
