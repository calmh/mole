package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/calmh/mole/auth"
	"os"
	"os/exec"
	//"github.com/calmh/mole/configuration"
	"crypto/sha1"
	"crypto/tls"
	"math/big"
	//"crypto/x509"
	//"github.com/cznic/zappy"
	"github.com/calmh/mole/proxy"
)

func sshTest() {
	print("Execing ssh\n")
	cmd := exec.Command("ssh", "anto")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	print("Done with ssh\n")
}

func authTest() {
	cfg := auth.Configuration{
		Server:       "sso.int.prnw.net",
		Port:         389,
		UseSSL:       false,
		BindTemplate: "uid=%s,cn=users,cn=accounts,dc=int,dc=prnw,dc=net",
	}

	e := cfg.Authenticate("jb", "kossnmu7")
	if e != nil {
		panic(e)
	}
	println("Auth success")
}

func cryptoTest() {
	c := elliptic.P256()
	pk, e := ecdsa.GenerateKey(c, rand.Reader)
	if e != nil {
		panic(e)
	}

	pkm := make(map[string]*big.Int)
	pkm["X"] = pk.X //.Bytes()
	pkm["Y"] = pk.Y //.Bytes()
	pkm["D"] = pk.D //.Bytes()
	v, e := json.Marshal(pkm)
	if e != nil {
		panic(e)
	}
	print(string(v) + "\n")

	msg := "Hey you"
	r, s, e := ecdsa.Sign(rand.Reader, pk, []byte(msg))
	if e != nil {
		panic(e)
	}
	m := map[string]interface{}{"msg": msg, "r": r, "s": s}
	v, e = json.Marshal(m)
	if e != nil {
		panic(e)
	}
	print(string(v) + "\n")
}

func Fingerprint(c *tls.Conn) string {
	state := c.ConnectionState()
	if len(state.PeerCertificates) != 1 {
		panic("incorrect certificate count")
	}
	s := sha1.New()
	s.Write(state.PeerCertificates[0].Raw)
	h := s.Sum(nil)
	return fmt.Sprintf("%x", h)
}

func main() {
}
