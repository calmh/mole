package main

import (
	"fmt"
	"github.com/mavricknz/ldap"
	"log"
)

var (
	ldapServer = "localhost"
	ldapPort   = 389
	ldapBind   = "uid=%s,cn=users"
)

func init() {
	authBackends["ldap"] = backendAuthenticateLDAP
	globalFlags.StringVar(&ldapServer, "ldap-host", ldapServer, "(for -auth=ldap) LDAP host")
	globalFlags.IntVar(&ldapPort, "ldap-port", ldapPort, "(for -auth=ldap) LDAP port")
	globalFlags.StringVar(&ldapBind, "ldap-bind", ldapBind, "(for -auth=ldap) LDAP bind template")
}

func backendAuthenticateLDAP(user, password string) bool {
	c := ldap.NewLDAPConnection(ldapServer, uint16(ldapPort))
	err := c.Connect()
	if err != nil {
		log.Println("ldap:", err)
		return false
	}

	err = c.Bind(fmt.Sprintf(ldapBind, user), password)
	if err != nil {
		log.Printf("ldap: %q: %s", user, err)
		return false
	}

	return true
}
