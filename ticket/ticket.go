// Package ticket generates and verifies authentication tickets
package ticket

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

type ticket struct {
	// Short JSON keys keep the ticket size down
	User     string `json:"u"`
	IP       string `json:"i"`
	Validity int64  `json:"v"`
}

// Init (re)initializes the session that tickets are based on. Init is called
// automatically on package initialization but may be called manually to
// invalidate all currently granted tickets.
func Init() {
	initKeyAndIV()
}

// Grant generates a ticket for the given user, IP and validity stamp.
func Grant(user, ip string, validity int64) string {
	bs, err := json.Marshal(ticket{user, ip, validity})
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
	err = json.Unmarshal(msg, &dec)
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
