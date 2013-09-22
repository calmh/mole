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
// with a 403 Forbidden error. Any other error is returned to the caller.
func authenticated(c *Client, r authenticatedRequest) (interface{}, error) {
	c.Ticket = serverIni.ticket
	br := bufio.NewReader(os.Stdin)
	for i := 0; i < retries; i++ {
		result, err := r()
		if err == nil || !strings.HasPrefix(err.Error(), "403 Forbidden") {
			return result, err
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
			cfg.Sections["server"]["ticket"] = ticket
			f, e = os.Create(configFile)
			fatalErr(e)
			err = cfg.Write(f)
			fatalErr(err)
			err = f.Close()
			fatalErr(err)
		} else {
			warnln(err.Error())
		}
	}

	return nil, fmt.Errorf("Too many authentication failures")
}
