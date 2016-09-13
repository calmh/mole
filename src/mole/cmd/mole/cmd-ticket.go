package main

import (
	"time"

	"mole/ansi"
)

func init() {
	addCommand(command{name: "ticket", fn: ticketCommand, descr: msgTicketShort})
}

func ticketCommand(args []string) {
	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	cl.Ticket = moleIni.Get("server", "ticket")
	tic, err := cl.ParseTicket()
	fatalErr(err)

	infof(msgTicketExplanation, ansi.Cyan(tic.User), ansi.Cyan(time.Time(tic.Validity).String()))
	for _, ip := range tic.IPs {
		infoln("  * ", ansi.Cyan(ip))
	}
}
