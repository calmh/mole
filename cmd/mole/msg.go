package main

const (
	msgDigShort      = `Dig tunnel`
	msgLsShort       = `List tunnels`
	msgPushShort     = `Push tunnel`
	msgShowShort     = `Show tunnel`
	msgTestShort     = `Test tunnel`
	msgUpgradeShort  = `Upgrade mole`
	msgVersionShort  = `Show version`
	msgRegisterShort = `Register with server`

	msgDigLong      = `"dig" connects the tunnel and sets up forwards`
	msgLsLong       = `"ls" lists tunnels, optionally filtering on a regular expression`
	msgPushLong     = `"push" sends a new or updated tunnel file to the server`
	msgShowLong     = `"show" show the tunnel configuration`
	msgTestLong     = `"dig" connects the tunnel and tests forwards`
	msgUpgradeLong  = `"upgrade" upgrades mole to the latest version`
	msgVersionLong  = `"version" shows current mole version`
	msgRegisterLong = `"register" sets mole up to talk to a server`

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

	msgErrPEMNoKey = `No ssh key found after PEM decode.`

	msgSshFirst = `ssh: Dial %s@%s`
	msgSshVia   = `ssh: Tunnel to %s@%s`

	msgVpncStarting  = `vpnc: Starting.`
	msgVpncStart     = `vpnc: Started (pid %d).`
	msgVpncStopping  = `vpnc: Stopping (pid %d).`
	msgVpncWait      = `vpnc: Waiting for connect...`
	msgVpncConnected = `vpnc: Connected.`
	msgVpncStopped   = `vpnc: Stopped.`

	msgOpncStarting  = `openconnect: Starting.`
	msgOpncStart     = `openconnect: Started (pid %d).`
	msgOpncStopping  = `openconnect: Stopping (pid %d).`
	msgOpncWait      = `openconnect: Waiting for connect...`
	msgOpncConnected = `openconnect: Connected.`
	msgOpncStopped   = `openconnect: Stopped.`

	msgWarnNoCert = `No server CA certificate present, cannot authenticate server.`

	msgDownloadingUpgrade = `Downloading upgrade...`
	msgUpgraded           = `Upgraded your mole to %s.`
	msgNoUpgrades         = `You are already running the latest version.`
	msgAutoUpgrades       = `
Mole uses automatic upgrades to keep your client up to date. To disable these
automatic upgrades (which is a bad idea for most users), add:

  [upgrades]
  automatic = no

to your ~/.mole/mole.ini file. To silence this message, you can instead add:

  [upgrades]
  automatic = yes
`

	msgExamples = `Examples:
  mole ls                # show all available tunnels
  mole ls foo            # show all available tunnels matching the non-achored regexp "foo"
  mole show foo          # show the hosts and forwards set up by the tunnel "foo"
  mole show -r foo       # show the raw config file for the tunnel "foo"
  sudo mole dig foo      # dig the tunnel "foo"
  sudo mole -d dig foo   # dig the tunnel "foo", while showing debug output
  mole push foo.ini      # create or update the "foo" tunnel from a local file
`

	msgFileNotInit = `File %q should have .ini extension`
	msgOkPushed    = `Pushed %q`

	msgErrNoTunModule = `Required tunnel module (kernel extension) not available and not loadable.`

	msgVpncUnavailable = `I cant't find a working "vpnc" on your system.

Tunnels marked as requiring "(vpnc)" will not work. Consider installing vpnc.
On Mac OS X, the recommended  way is to use homebrew (http://brew.sh/) to
install the "tuntap" and "vpnc" packages. On Linux, do whatever your
distribution recommends.`

	msgOpncUnavailable = `I cant't find a working "openconnect" on your system.

Tunnels marked as requiring "(opnc)" will not work. Consider installing
OpenConnect. On Mac OS X, the recommended  way is to use homebrew
(http://brew.sh/) to install the "tuntap" and "openconnect" packages.
On Linux, do whatever your distribution recommends.`

	msgNeedsAuth       = `Authentication required. Enter your LDAP credentials.`
	msgUsername        = `Username: `
	msgPassword        = `Password: `
	msgPasswordVisible = `Password will be visible when typed.`

	msgNoHost = `No server hostname is configured. Have you run "mole register"?`
)
