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

func authenticate(c *Client) (string, error) {
	c.Ticket = serverIni.ticket
	br := bufio.NewReader(os.Stdin)
	for i := 0; i < retries; i++ {
		user, err := c.Ping()
		if err == nil {
			return user, nil
		}
		if !strings.HasPrefix(err.Error(), "403 Forbidden") {
			return "", err
		}

		infoln(msgNeedsAuth)
		fmt.Printf(msgUsername)
		bs, _, err := br.ReadLine()
		fatalErr(err)
		user = string(bs)
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

			return user, nil
		} else {
			warnln(err.Error())
		}
	}

	return "", nil
}
