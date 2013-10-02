package main

import (
	"bufio"
	"fmt"
	"github.com/calmh/mole/ini"
	"os"
	"path"
	"strings"
)

const retries = 3

type authenticatedRequest func() (interface{}, error)

// authenticated calls the request r and performs authentication if it fails
// with a 401 Unauthorized error. Any other error is returned to the caller.
func authenticated(c *Client, r authenticatedRequest) (interface{}, error) {
	c.Ticket = serverIni.ticket
	br := bufio.NewReader(os.Stdin)
	i := 0
	for {
		result, err := r()
		if err == nil || !strings.HasPrefix(err.Error(), "401 Unauthorized") {
			return result, err
		}

		if i >= retries {
			return nil, fmt.Errorf("Too many authentication failures")
		}

		infoln(msgNeedsAuth)
		fmt.Printf(msgUsername)
		bs, _, err := br.ReadLine()
		fatalErr(err)
		user := string(bs)
		pass := readpass(msgPassword)

		ticket, err := c.GetTicket(user, pass)
		if err == nil {
			c.Ticket = ticket

			configFile := path.Join(globalOpts.Home, "mole.ini")
			f, e := os.Open(configFile)
			fatalErr(e)
			cfg := ini.Parse(f)
			_ = f.Close()
			cfg.Set("server", "ticket", ticket)
			f, e = os.Create(configFile)
			fatalErr(e)
			err = cfg.Write(f)
			fatalErr(err)
			err = f.Close()
			fatalErr(err)
		} else {
			warnln(err.Error())
		}

		i++
	}
}
