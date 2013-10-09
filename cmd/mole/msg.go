package main

const (
	msgMainUsage     = "mole [options] <command> [command-options]"
	msgDigUsage      = "mole [global-options] dig [options] <tunnel>"
	msgInstallUsage  = "mole [global-options] install [package]"
	msgLsUsage       = "mole [global-options] ls [options] [regexp]"
	msgPushUsage     = "mole [global-options] push <tunnelfile>"
	msgRegisterUsage = "mole [global-options] register [options] <server>"
	msgShowUsage     = "mole [global-options] show [options] <tunnel>"
	msgTestUsage     = "mole [global-options] test [options] <tunnel>"
	msgUpgradeUsage  = "mole [global-options] upgrade [options]"
	msgVersionUsage  = "mole [global-options] version [options]"

	msgDigShort      = "Dig tunnel"
	msgInstallShort  = "Install package"
	msgLsShort       = "List tunnels"
	msgPushShort     = "Push tunnel"
	msgRegisterShort = "Register with server"
	msgRmShort       = "Delete tunnel"
	msgShowShort     = "Show tunnel"
	msgTestShort     = "Test tunnel"
	msgUpgradeShort  = "Upgrade mole"
	msgVersionShort  = "Show version"

	msgDebugEnabled = "Debug output enabled."

	msgErrGainRoot         = "Error: missing root privileges to execute %q.\nTo give mole root access, execute it using sudo. Mole itself drops root privilege on startup and executes as the non privileged user. However, child processes such as ifconfig will inherit the saved user ID and have the ability to become root as necessary."
	msgErrNoVPN            = "No VPN provider for %q available. Try 'mole install' to see what packages are available or use the packaging system native to your platform."
	msgErrIncorrectFwd     = "Badly formatted fwd command %q."
	msgErrIncorrectFwdSrc  = "Badly formatted fwd source %q."
	msgErrIncorrectFwdDst  = "Badly formatted fwd destination %q."
	msgErrIncorrectFwdIP   = "Cannot forward from non-existent local IP %q."
	msgErrIncorrectFwdPriv = "Cannot forward from privileged port %q (<1024)."
	msgErrNoSuchCommand    = "No such command %q. Try 'mole help'."
	msgErrNoHome           = "No home directory that I could find; cannot proceed."
	msgErrPEMNoKey         = "No ssh key found after PEM decode."

	msgVpncStart     = "vpnc: Started (pid %d)."
	msgVpncStopping  = "vpnc: Stopping (pid %d)."
	msgVpncWait      = "vpnc: Waiting for connect..."
	msgVpncConnected = "vpnc: Connected."
	msgVpncStopped   = "vpnc: Stopped."

	msgOpncStart     = "openconnect: Started (pid %d)."
	msgOpncStopping  = "openconnect: Stopping (pid %d)."
	msgOpncWait      = "openconnect: Waiting for connect..."
	msgOpncConnected = "openconnect: Connected."
	msgOpncStopped   = "openconnect: Stopped."

	msgDownloadingUpgrade = "Downloading upgrade..."
	msgUpgraded           = "Upgraded your mole to %s."

	msgFileNotInit    = "File %q should have .ini extension"
	msgOkPushed       = "Pushed %q"
	msgErrNoTunModule = "Required tunnel module (kernel extension) not available and not loadable."

	msgNeedsAuth       = "Authentication required. Enter your LDAP credentials."
	msgUsername        = "Username: "
	msgPassword        = "Password: "
	msgPasswordVisible = "Password will be visible when typed."

	msgNoHost = "No server hostname is configured. Have you run 'mole register'?"

	msgNoPackages = "There are no packages available for installation on your OS/architecture."

	msgRegistered = "Registered with %q. Consider running 'mole install' to see what extra packages, such as VPN providers, are available."

	msgOkDeleted = "Deleted %q."

	msg530 = "530 Version Unacceptable\nYour client is either too new or too old to talk to this server. Make sure you are in fact registered with the correct server and try 'mole upgrade' to get the newest client."

	msgLatest       = "You are running the latest version."
	msgAutoUpgrades = "Mole uses automatic upgrades to keep your client up to date. To disable these automatic upgrades (which is a bad idea for most users) or silence this message, see 'mole upgrade -help'."
	msgUpdatedHost  = "Updated configured server name to %q."

	msgLsFlags = `  o····  Requires OpenConnect
  v····  Requires vpnc
  ·k···  Uses SSH with key authentication
  ··p··  Uses SSH with password authentication
  ···l·  Uses local (non-SSH) forwards
  ···s·  Uses SOCKS proxy
  ····E  Parse or access error reading tunnel
  ····U  Unknown or unsupported features required
`
)
