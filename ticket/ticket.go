// Package ticket generates and verifies authentication tickets
package ticket

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/base64"
	"errors"
)

type ticket struct {
	Random   []byte
	User     string
	IP       string
	Validity int64
}

const nonceSize = 8

// Init (re)initializes the session that tickets are based on. Init is called
// automatically on package initialization but may be called manually to
// invalidate all currently granted tickets.
func Init() {
	initKeyAndIV()
}

// Grant generates a ticket for the given user, IP and validity stamp.
func Grant(user, ip string, validity int64) string {
	t := ticket{User: user, IP: ip, Validity: validity}
	t.Random = make([]byte, nonceSize)
	n, err := rand.Read(t.Random)
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

// Verify checks that a ticket is valid for the given IP and validity time,
// and returns the authenticated user name or an error.
func Verify(tic, ip string, validity int64) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(tic)
	if err != nil {
		return "", err
	}

	msg, err := decryptAndHash(bs)
	if err != nil {
		return "", err
	}

	var dec ticket
	_, err = asn1.Unmarshal(msg, &dec)
	if err != nil {
		return "", err
	}

	if dec.IP != ip {
		return "", errors.New("incorrect IP")
	}

	if dec.Validity < validity {
		return "", errors.New("expired")
	}

	return dec.User, nil
}
