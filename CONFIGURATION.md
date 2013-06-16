Configuration Reference
=======================

This document describes the file format used for tunnel definitions. It follows
the common "INI" format with section headers in `[brackets]` followed by
directives on the form `element="a value"`.

The purpose of mole is to allow easy setup of SSH/VPN connections and related
tunnels. All tunnel definition files must be entirely self contained, that is
it must never be necessary for the user to interact during the login process to
give a password or similar. Tunnel definitions should also be self documented,
so make sure to use clear and descriptive names for hosts and forwards.

A tunnel definition consists of:

  - exactly one `general` section,
  - zero or more `hosts` sections,
  - zero or more `forwards` sections,
  - an optional `vpnc` section,
  - an optional `vpn routes` section which must only be present in combination
    with a `vpnc section`.

You need to *either* have at least one host or at least one forward.

Section `general`
------------------

The `general` section contains four mandatory elements;

  - `description` - A free text description of this configuration that is
    displayed by `mole list`.
  - `author` - Name and email of the configuration file author.
  - `main` - Name of the host to connect to when the tunnel definition is
    invoked.
  - `version` - Configuration format version. Must be set to `3` currently.
  - `aliases[]` - An array of extra name to IP mappings (white space separated)
    to insert into the hosts file. (Optional)

### Example

    description = OperatorOne (UK, production network)
    author = Jakob Borg <jakob@example.com>
    main = op1prod
    version = 3
    aliases[] = srv1 127.0.0.1
    aliases[] = srv2 127.0.0.2

Section `hosts`
---------------

There can be any number of host sections. Each describes a host that is
reachable via SSH, either directly or via another host. The name of the host is
set in the section header, after the `host` keyword. The host name cannot
contain spaces. The following elements can be set for each host:

  - `addr` - IP address or DNS name of the host.
  - `port` - Port number where an SSH daemon is listening.
  - `user` - The username to use when authenticating.
  - `password` - Password to use when authenticating.
  - `key` - SSH key to use when authenticating.
  - `via` - Name of another host to bounce via in order to reach this host.
    Must be the name of host defined elsewhere in the same tunnel definition
    file.
  - `prompt` - Override the regular expression that recognizes the destination
    host prompt. The default is usually fine, but if there's some unusual stuff
    on the other side an override might be necessary. This is only relevant for
    the `main` host.
  - `keepalive` - SSH keep alive interval (seconds). If the server is
    unresponsive for longer than this time, the connection will be terminated.
    Default is 180, minimum 15.
  - `socks` - Address and port of a SOCKS proxy to connect to this host
    via. Cannot be used together with `via` (although another host can
    of course connect via this one).

Of these, `addr` and `user` are mandatory. `port` is optional and
defaults to `22`. Either `password` or `key` must be specified so the
login can be completed noninteratively. In case `key` is used, it be a
quoted string containing a valid SSH private key with newlines replaces
by `\n`. The key must not be protected by a password.

### Example

    [hosts.op1jump]
    addr = 192.168.10.10
    user = admin
    password = 3x4mpl3
    port = 2222
    prompt = "~>"
    socks = 192.168.56.101:1080

    [hosts.op1prod]
    addr = 10.0.33.66
    user = admin
    key = "-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEAxymzAVzTX6oJTlZ5uCkqjdrDb/ovLZ6VktH+i5h2wdJpyT3f\ns2Q23e ...etc"
    via = op1jump

Section `forwards`
-----------------

The forward sections describes SSH port forwardings that will be set up
when the destination is reached. The name of the forward is set in the
section header after the `forwards` keyword and may contains spaces and
special characters within reason. The first part (i.e. up to any
whitespace) will be used as host name and inserted into the local hosts
file upon tunnel establishment. Any dots will need to be escaped, i.e.
`[hosts.example\.com]` to name the forward `example.com`.

Each element withing the forward section is a pair on the form
`<local address>:<port> = <remote address>:<port>`. The local side can use
addresses other than 127.0.0.1 but still in the 127.0.0.0/8 block; these
will be added to the local loopback interface if they don't already
exist.

The remote port can be left out if it is the same as the local port. If
the remote port is left out, the local port can be a dash separate
range. Local ports need to be higher than 1024.

If there is no SSH configuration, but there is a VPN configuration, then
the forwards will be done from the local computer.  This can be used to
provide the user with the same usage pattern as in the SSH forward case
and also keep the tunnel definition self documenting.

### Example

    [forwards.hostA]
    127.0.0.1:8443        = 10.0.33.69.193:443
    127.0.0.1:22001-22005 = 10.0.33.69.193

    [forwards.hostB]
    127.0.0.2:22001-22005 = 10.0.33.70.194


Section `vpnc`
--------------

The `vpnc` section defines a configuration for the vpnc Cisco VPN command line
client. The elements are any configuration directives recognized by vpnc, with
spaces replaced by underscores. Element names cannot contain special characters
such as paranthesis, but since there is no equal sign or similar in a vpnc
configuration a line like

    DPD idle timeout (our side) 0

can be represented in the tunnel definition as

    DPD_idle_timeout = "(our side) 0"

The configuration must contain Xauth username and password since it must be
able to connect noninteractively.

The `vpnc` section is optional and requires that vpnc be installed if present.
If present, the VPN will be connected before any attempts are made to connect
to hosts defined as above.

### Example

    [vpnc]
    IPSec_gateway = 213.154.23.72
    IPSec_ID = IPSECGROUP
    IPSec_secret = abrakadabra
    Xauth_username = extuser
    Xauth_password = K0ssanmu7
    IKE_Authmode = psk
    DPD_idle_timeout = "(our side) 0"
    NAT_Traversal_Mode = force-natt
    Local_Port = 0
    Cisco_UDP_Encapsulation_Port = 0

Section `vpn routes`
--------------------

The `vpn routes` section is optional and can be present when there is a `vpnc`
section as above. If present, any "split VPN" routes sent by the VPN server
will be discarded and the routes mentioned in this section will be used
instead. Routes for specific local IP:s sent by the VPN server (such as a DNS
server) will be allowed regardless. The format of elements in this section is
`<network> = <mask bits>`, so to allow 192.0.2.0/24 add an element
`192.0.2.0=24`. The purpose of this section is to avoid installing unwanted
routes such as a default route or routes that may conflict with the local
topology.

### Example

    [vpn routes]
    10.200.0.0 = 16
    192.168.10.0 = 24
    192.168.20.0 = 24
