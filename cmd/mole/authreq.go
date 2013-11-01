package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const retries = 3

type authenticatedRequest func() (interface{}, error)

// authenticated calls the request r and performs authentication if it fails
// with a 401 Unauthorized error. Any other error is returned to the caller.
func authenticated(c *Client, r authenticatedRequest) (interface{}, error) {
	c.Ticket = moleIni.Get("server", "ticket")
	br := bufio.NewReader(os.Stdin)

	result, err := r()
	if err == nil || !strings.HasPrefix(err.Error(), "401 Unauthorized") {
		return result, err
	}

	for i := 0; i < retries; i++ {
		infoln(msgNeedsAuth)
		user := moleIni.Get("server", "user")
		if user == "" || i != 0 {
			// Reuse known username on first attempt only
			fmt.Printf(msgUsername)
			bs, _, err := br.ReadLine()
			fatalErr(err)
			user = string(bs)
		}
		pass := readpass(fmt.Sprintf(msgPassword, user))

		ticket, err := c.GetTicket(user, pass)
		if err == nil {
			c.Ticket = ticket
			moleIni.Set("server", "ticket", ticket)
			moleIni.Set("server", "user", user)
			saveMoleIni()
		} else {
			warnln(err.Error())
		}

		result, err := r()
		if err == nil || !strings.HasPrefix(err.Error(), "401 Unauthorized") {
			return result, err
		}
	}

	return nil, fmt.Errorf("Too many authentication failures")
}
