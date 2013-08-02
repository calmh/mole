package auth

import (
	"fmt"
	"errors"
	"github.com/calmh/mole/github.com/mmitton/ldap"
)

var ServerError = errors.New("could not connect to LDAP server")
var AuthenticationFailed = errors.New("invalid credentials")

type Configuration struct {
	Server string
	Port uint16
	UseSSL bool
	BindTemplate string
}

func (cfg *Configuration) Authenticate(user, password string) error {
	// ssl seems to require known ca
	c, e := ldap.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Server, cfg.Port))
	if e != nil {
		return ServerError
	}

	e = c.Bind(fmt.Sprintf(cfg.BindTemplate, user), password)
	if e != nil {
		return AuthenticationFailed
	}

	return nil
}

