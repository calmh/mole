package main

const (
	msgDebugEnabled = `Debug output enabled.`

	msgErrGainRoot = `Error: missing root privileges to execute "%s".

To give mole root access, execute it using sudo. Mole itself drops root
privilege on startup and executes as the non privileged user. However, child
processes such as ifconfig will inherit the saved user ID and have the
ability to become root as necessary.
`

	msgErrNoVPN = `No VPN provider for "%s" available.`

	msgErrIncorrectFwd     = `Badly formatted fwd command %q.`
	msgErrIncorrectFwdSrc  = `Badly formatted fwd source %q.`
	msgErrIncorrectFwdDst  = `Badly formatted fwd destination %q.`
	msgErrIncorrectFwdIP   = `Cannot forward from non-existent local IP %q.`
	msgErrIncorrectFwdPriv = `Cannot forward from privileged port %q (<1024).`
	msgErrNoSuchCommand    = `No such command %q. Try "help".`

	msgErrNoHome = `No home directory that I could find; cannot proceed.`

	msgWarnSetrlimit = `Warning: setrlimit: %v`

	msgErrPEMNoKey = `No ssh key found after PEM decode.`

	msgSshFirst = `ssh: Dial %s@%s`
	msgSshVia   = `ssh: Tunnel to %s@%s`

	msgVpncStart     = `vpnc: Started (pid %d).`
	msgVpncStopping  = `vpnc: Stopping (pid %d).`
	msgVpncWait      = `vpnc: Waiting for connect...`
	msgVpncConnected = `vpnc: Connected.`
	msgVpncStopped   = `vpnc: Stopped.`

	msgOpncStart     = `openconnect: Started (pid %d).`
	msgOpncStopping  = `openconnect: Stopping (pid %d).`
	msgOpncWait      = `openconnect: Waiting for connect...`
	msgOpncConnected = `openconnect: Connected.`
	msgOpncStopped   = `openconnect: Stopped.`

	msgStatistics = ` -- %d bytes in, %d bytes out`
)

var (
	msgOk = bold(green("\nOK"))
)
