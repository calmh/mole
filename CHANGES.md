v3.2 to v4.0
============

## Improved installation

- Single binary to install. No longer depends on node, ssh, or expect.
  (Still depends on vpnc and openconnect for tunnels that require
  those.)

- Easy manual upgrades (`mole upgrade`).

- Automatic upgrades (with option to disable).

## Improved security

- Authenticates the user against LDAP (SSO). Successful authentication
  gets you a ticket that is valid for one week from the same IP number.
  Up to four IP numbers will be added to the same ticket to allow for
  some mobility. Requires server version 4.0.0 or higher.

- No longer writes any temporary files containing cleartext secrets.

- Requires explicit sudo invocation for reconfiguring interfaces or
  starting VPNs. Drops privileges on startup while keeping them
  available for ifconfig, vpnc, etc.

- Shows randomart rendering of the server certificate when registering.

## Improved reliability

- Uses internal SSH authentication, so does not need to understand SSH
  client password prompts.

- Does not set up a terminal session over SSH, so does not need to
  recognize or understand remote server shell prompt to confirm success.

- Does not attempt (and fail) to juggle sudo.

## Other new features and improvemens

- Add new port forwards on the fly using the mole shell (`fwd` command).

- Test all forwards for connection establishment in the mole shell
  (`test` command)

- Show number of connections and transferred data per forward in the
  mole shell (`stats` command).

- `mole ls <regexp>` to list a subset of tunnels based on regexp match
  of tunnel name, description or host names.

- `mole test <tunnel>` to automatically verify a tunnels connectivity.

- Bash completion. Add `eval "$(mole bashcomp)"` to your `.bash_profile`
  or similar to get tab completion of commands and tunnel names.

- Due to the LDAP authentication, there is no longer any need to
  maintain a separate user database on the server.

- Improved comment handling in tunnel files.

## Windows build

- No support for vpnc or openconnect.

- No support for extra localhost addresses; uses port remapping. Could
  be worked around with some PowerShell scripting if there's demand.

