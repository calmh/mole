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
	fn   func(http.ResponseWriter, *http.Request)
	auth bool
}

var (
	// Pattern to handleFunc
	handlers = map[string]handler{}

	// CLI options
	listenAddr = ":9443"
	storeDir   = "~/mole-store"
	certFile   = "crt/server-cert.pem"
	keyFile    = "crt/server-key.pem"
	auditFile  = "audit.log"
	auditIntv  = 86400 * time.Second
	noAuth     = false
)

func main() {
	flag.StringVar(&listenAddr, "listen", listenAddr, "HTTPS listen address")
	flag.StringVar(&storeDir, "store", storeDir, "Mole store directory")
	flag.StringVar(&certFile, "cert", certFile, "Certificate file, relative to store directory")
	flag.StringVar(&keyFile, "key", keyFile, "Key file, relative to store directory")
	flag.StringVar(&auditFile, "auditfile", auditFile, "Audit file, relative to store directory")
	flag.DurationVar(&auditIntv, "auditintv", auditIntv, "Audit file rotation interval")
	flag.BoolVar(&noAuth, "no-auth", noAuth, "Disable authentication requirements")
	flag.Parse()

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

	for pattern, handler := range handlers {
		setupHandler(pattern, handler)
		log.Printf("Registered handler for %q", pattern)
	}

	http.ListenAndServeTLS(listenAddr, path.Join(storeDir, certFile), path.Join(storeDir, keyFile), nil)
}

func setupHandler(p string, h handler) {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		if h.auth && !noAuth {
			if !authenticate(rw, req) {
				audit(req, "rejected")
				rw.WriteHeader(403)
				return
			}
		}
		audit(req, "accepted")
		h.fn(rw, req)
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
