// Package ticket generates and verifies authentication tickets
package ticket

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"io"
)

type Ticket struct {
	Nonce    []byte
	User     string
	IP       []string
	Validity int64
}

const (
	nonceSize = 8
)

var (
	ErrExpired   = errors.New("expired")
	ErrInvalidIP = errors.New("invalid IP")
)

// Init (re)initializes the session that tickets are based on. Init is called
// automatically on package initialization but may be called manually to
// invalidate all currently granted tickets.
func Init() {
	initKeyAndIV(rand.Reader)
}

// LoadKey initializes the session key that tickets are based on from the
// Reader.
func LoadKey(r io.Reader) {
	initKeyAndIV(r)
}

// Grant generates a ticket for the given user, IP and validity stamp.
func Grant(user, ip string, validity int64) string {
	t := Ticket{User: user, IP: []string{ip}, Validity: validity}
	return t.String()
}

// Verify checks that a ticket is valid for the given IP and validity time,
// and returns the authenticated user name or an error.
func Verify(tic, ip string, validity int64) (string, error) {
	dec, err := Load(tic)
	if err != nil {
		return "", err
	}

	foundIp := false
	for _, dip := range dec.IP {
		if dip == ip {
			foundIp = true
			break
		}
	}
	if !foundIp {
		return "", ErrInvalidIP
	}

	if dec.Validity < validity {
		return "", ErrExpired
	}

	return dec.User, nil
}

func (t Ticket) String() string {
	t.Nonce = make([]byte, nonceSize)
	n, err := rand.Read(t.Nonce)
	if n != nonceSize || err != nil {
		panic(err)
	}

	bs, err := asn1.Marshal(t)
	if err != nil {
		panic(err)
	}

	enc := hashAndEncrypt(bs)
	return base64.StdEncoding.EncodeToString(enc)
}

func Load(tic string) (*Ticket, error) {
	bs, err := base64.StdEncoding.DecodeString(tic)
	if err != nil {
		return nil, err
	}

	msg, err := decryptAndHash(bs)
	if err != nil {
		return nil, err
	}

	var dec Ticket
	_, err = asn1.Unmarshal(msg, &dec)
	if err != nil {
		return nil, err
	}

	return &dec, nil
}
