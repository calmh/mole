package main

const (
	msgBashcompShort = "Bash completion"
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

	msgBashcompLong = "'mole bashcomp' outputs the bash completion script; use as 'eval \"$(mole bashcomp)\"' in your .bash_profile"
	msgDigLong      = "'mole dig' connects the tunnel and sets up forwards"
	msgInstallLong  = "'mole install' installs a binary package fetched from the server"
	msgLsLong       = "'mole ls' lists tunnels, optionally filtering on a regular expression"
	msgPushLong     = "'mole push' sends a new or updated tunnel file to the server"
	msgRegisterLong = "'mole register' sets mole up to talk to a server"
	msgRmLong       = "'mole rm' deletes the specified tunnel from the server"
	msgShowLong     = "'mole show' shows the tunnel configuration"
	msgTestLong     = "'mole dig' connects the tunnel and tests forwards"
	msgUpgradeLong  = "'mole upgrade' upgrades mole to the latest version"
	msgVersionLong  = "'mole version' shows current mole version"

	msgDebugEnabled = "Debug output enabled."

	msgErrGainRoot         = "Error: missing root privileges to execute %q.\nTo give mole root access, execute it using sudo. Mole itself drops root privilege on startup and executes as the non privileged user. However, child processes such as ifconfig will inherit the saved user ID and have the ability to become root as necessary."
	msgErrNoVPN            = "No VPN provider for %q available."
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
	msgNoUpgrades         = "You are already running the latest version."

	msgFileNotInit     = "File %q should have .ini extension"
	msgOkPushed        = "Pushed %q"
	msgErrNoTunModule  = "Required tunnel module (kernel extension) not available and not loadable."
	msgVpncUnavailable = "I cant't find a working vpnc on your system. Tunnels marked as requiring \"(vpnc)\" will not work. Consider installing vpnc. On Mac OS X, try 'mole install' to see available packages. On Linux, do whatever your distribution recommends."
	msgOpncUnavailable = "I cant't find a working OpenConnect on your system. Tunnels marked as requiring \"(opnc)\" will not work. Consider installing OpenConnect. On Mac OS X, try 'mole install' to see available packages. On Linux, do whatever your distribution recommends."

	msgNeedsAuth       = "Authentication required. Enter your LDAP credentials."
	msgUsername        = "Username: "
	msgPassword        = "Password: "
	msgPasswordVisible = "Password will be visible when typed."

	msgNoHost = "No server hostname is configured. Have you run 'mole register'?"

	msgNoPackages = "There are no packages available for installation on your OS/architecture."

	msgRegistered = "Registered with %q. Consider running 'mole install' to see what extra packages, such as VPN providers, are available."

	msgOkDeleted = "Deleted %q."

	msg530 = "530 Version Unacceptable\nYour client is either too new or too old to talk to this server. Make sure you are in fact registered with the correct server and try 'mole upgrade' to get the newest client."

	msgAutoUpgrades = `Mole uses automatic upgrades to keep your client up to date. To disable these automatic upgrades (which is a bad idea for most users), add:

      [upgrades]
      automatic = no

    to your ~/.mole/mole.ini file. To silence this message, you can set:

      [upgrades]
      automatic = yes
`

	msgExamples = `Examples:
  mole ls                # show all available tunnels
  mole ls foo            # show all available tunnels matching the regexp "foo"
  mole show foo          # show the hosts and forwards in the tunnel "foo"
  sudo mole dig foo      # dig the tunnel "foo"
  sudo mole -d dig foo   # dig the tunnel "foo", while showing debug output
  mole push foo.ini      # create or update the "foo" tunnel from a local file
  mole install           # list packages available for installation
  mole install vpnc      # install a package named vpnc
`
)
