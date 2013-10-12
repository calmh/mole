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
	listenAddr        = ":9443"
	storeDir          = "~/mole-store"
	certFile          = "cert.pem"
	keyFile           = "key.pem"
	auditFile         = "audit.log"
	auditIntv         = 86400 * time.Second
	auth              = "none"
	readOnly          = false
	disableGit        = false
	canonicalHostname = ""
)

var buildVersion string

var globalFlags = flag.NewFlagSet("molesrv", flag.ExitOnError)

func init() {
	globalFlags.Usage = usageFor(globalFlags, "molesrv [options]")
	globalFlags.StringVar(&listenAddr, "listen", listenAddr, "HTTPS listen address")
	globalFlags.StringVar(&storeDir, "store-dir", storeDir, "Mole store directory")
	globalFlags.StringVar(&certFile, "cert-file", certFile, "Certificate file (relative to store directory)")
	globalFlags.StringVar(&keyFile, "key-file", keyFile, "Key file (relative to store directory)")
	globalFlags.StringVar(&auditFile, "audit-file", auditFile, "Audit file (relative to store directory)")
	globalFlags.DurationVar(&auditIntv, "audit-intv", auditIntv, "Audit file creation interval")
	globalFlags.StringVar(&auth, "auth", auth, "Authentication backend")
	globalFlags.BoolVar(&readOnly, "no-write", readOnly, "Disallow writable client operations (push, rm, etc)")
	globalFlags.BoolVar(&disableGit, "no-git", disableGit, "Do not treat the store as a git repository")
	globalFlags.StringVar(&canonicalHostname, "canonical-hostname", canonicalHostname, "Server hostname to advertise as canonical")
}

func addHandler(hnd handler) {
	handlers[hnd.pattern] = append(handlers[hnd.pattern], hnd)
}

func main() {
	globalFlags.Parse(os.Args[1:])

	if _, ok := authBackends[auth]; !ok {
		log.Fatalf("Unknown auth backend %q", auth)
	}

	if buildVersion == "" {
		buildVersion = "4.0-dev-unknown"
	}

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
		rw.Header().Set("X-Mole-Version", buildVersion)
		rw.Header().Set("X-Mole-Canonical-Hostname", canonicalHostname)

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

			if h.auth && auth != "none" {
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
