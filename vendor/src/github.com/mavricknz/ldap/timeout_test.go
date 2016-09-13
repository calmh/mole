package ldap

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"
)

func TestSearchTimeout(t *testing.T) {
	fmt.Printf("TestSearchTimeout: starting...\n")
	l := NewLDAPConnection(ldap_server, ldap_port)
	l.NetworkConnectTimeout = 5000 * time.Millisecond
	l.ReadTimeout = 30 * time.Second
	l.AbandonMessageOnReadTimeout = true
	err := l.Connect()

	if err != nil {
		t.Error(err)
		return
	}
	if l == nil {
		t.Errorf("No Connection.")
		return
	}
	defer l.Close()

	search_request := NewSimpleSearchRequest(
		base_dn,
		ScopeWholeSubtree,
		filter[0],
		attributes,
	)

	sr, err := l.Search(search_request)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("TestSearchTimeout: %s -> num of entries = %d\n", search_request.Filter, len(sr.Entries))
}

func TestSearchTimeoutSSL(t *testing.T) {
	fmt.Printf("TestSearchTimeoutSSL: starting...\n")
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	l := &LDAPConnection{
		Addr:                  fmt.Sprintf("%s:%d", ldap_server, 636),
		IsSSL:                 true,
		TlsConfig:             config,
		NetworkConnectTimeout: 5000 * time.Millisecond,
		ReadTimeout:           30 * time.Second,
	}

	err := l.Connect()
	if err != nil {
		t.Error(err)
		return
	}
	defer l.Close()

	search_request := NewSimpleSearchRequest(
		base_dn,
		ScopeWholeSubtree,
		filter[0],
		attributes,
	)

	sr, err := l.Search(search_request)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("TestSearchTimeoutSSL: %s -> num of entries = %d\n", search_request.Filter, len(sr.Entries))
}
