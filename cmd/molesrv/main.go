package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/calmh/mole/ticket"
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
	auditFile         = "audit.log"
	auditIntv         = 86400 * time.Second
	auth              = "none"
	canonicalHostname = ""
	certFile          = "cert.pem"
	disableGit        = false
	initStore         = false
	keyFile           = "key.pem"
	listenAddr        = ":9443"
	readOnly          = false
	storeDir          = "~/mole-store"
	ticketKeyFile     = ""
)

var buildVersion string

var globalFlags = flag.NewFlagSet("molesrv", flag.ExitOnError)

func init() {
	globalFlags.Usage = usageFor(globalFlags, "molesrv [options]")
	globalFlags.StringVar(&auditFile, "audit-file", auditFile, "Audit file (relative to store directory)")
	globalFlags.DurationVar(&auditIntv, "audit-intv", auditIntv, "Audit file creation interval")
	globalFlags.StringVar(&auth, "auth", auth, "Authentication backend")
	globalFlags.StringVar(&canonicalHostname, "canonical-hostname", canonicalHostname, "Server hostname to advertise as canonical")
	globalFlags.StringVar(&certFile, "cert-file", certFile, "Certificate file (relative to store directory)")
	globalFlags.BoolVar(&initStore, "init-store", initStore, "Initialize store directory and certificates")
	globalFlags.StringVar(&keyFile, "key-file", keyFile, "Key file (relative to store directory)")
	globalFlags.StringVar(&listenAddr, "listen", listenAddr, "HTTPS listen address")
	globalFlags.BoolVar(&disableGit, "no-git", disableGit, "Do not treat the store as a git repository")
	globalFlags.BoolVar(&readOnly, "no-write", readOnly, "Disallow writable client operations (push, rm, etc)")
	globalFlags.StringVar(&storeDir, "store-dir", storeDir, "Mole store directory")
	globalFlags.StringVar(&ticketKeyFile, "ticket-file", ticketKeyFile, "Ticket key file. Leave blank to autogenerate key on startup.")
	globalFlags.StringVar(&buildVersion, "version", buildVersion, "Version string to advertise")
}

func addHandler(hnd handler) {
	handlers[hnd.pattern] = append(handlers[hnd.pattern], hnd)
}

func main() {
	globalFlags.Parse(os.Args[1:])
	if globalFlags.NArg() > 0 {
		log.Fatalf("Unrecognized extra arguments: %q", strings.Join(globalFlags.Args(), " "))
	}

	if strings.HasPrefix(storeDir, "~/") {
		home := getHomeDir()
		storeDir = strings.Replace(storeDir, "~", home, 1)
	}

	if initStore {
		dataDir := path.Join(storeDir, "data")
		err := os.MkdirAll(dataDir, 0700)
		if err != nil {
			log.Fatal(err)
		}

		keys = make(map[string]string)
		err = saveKeys()
		if err != nil {
			log.Fatal(err)
		}

		newCertificate()

		if !disableGit {
			gitInit(dataDir)
			gitCommit(dataDir, "Initial", "server")
		}

		extraDir := path.Join(storeDir, "extra")
		err = os.MkdirAll(extraDir, 0700)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path.Join(extraDir, "upgrades.json"), []byte("{}"), 0644)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path.Join(extraDir, "packages.json"), []byte("{}"), 0644)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path.Join(extraDir, "packages.json.example"), []byte(`{
    "darwin-amd64": [
        {"package":"tun", "description":"Tunnel driver (required by vpnc, openconnect)"},
        {"package":"openconnect", "description":"OpenConnect client support"},
        {"package":"vpnc", "description":"VPNC client support"}
    ]
}`), 0644)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("OK: Initialized store directory", storeDir)
		return
	}

	if _, ok := authBackends[auth]; !ok {
		log.Fatalf("Unknown auth backend %q", auth)
	}

	if buildVersion == "" {
		buildVersion = "4.0-dev-unknown"
	}

	log.Println("mole server", buildVersion)

	err := loadKeys()
	if err != nil {
		log.Println("Warning:", err)
	}
	if keys == nil {
		keys = make(map[string]string)
		log.Println("Initialized new key store")
	}

	if ticketKeyFile != "" {
		f, err := os.Open(ticketKeyFile)
		if err != nil {
			log.Fatal(err)
		}
		ticket.LoadKey(f)
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

func newCertificate() {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatal(err)
	}

	notBefore := time.Now()
	notAfter := time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC)

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName: "mole",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}

	certOut, err := os.Create(path.Join(storeDir, certFile))
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(path.Join(storeDir, keyFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
}
