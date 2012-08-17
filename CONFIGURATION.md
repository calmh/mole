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
  - zero or more `host` sections,
  - zero or more `forward` sections,
  - an optional `vpnc` section,
  - an optional `vpn routes` section which must only be present in combination
    with a `vpnc section`.

You need to *either* have at least one `host` or at least one `localforward`.
You can't have `forward` without having at least one `host` to do them through,
but `localforward` doesn't need a host. You can't combine `host`/`forward` and
`localforward` -- either you ssh somewhere and use port forwards through there
or you do it locally.

Section `general`
------------------

The `general` section contains three mandatory elements;

  - `description` - A free text description of this configuration that is
    displayed by `mole list`.
  - `author` - Name and email of the configuration file author.
  - `main` - Name of the host to connect to when the tunnel definition is
    invoked.

### Example

    description="OperatorOne (UK, production network)"
    author="Jakob Borg <jakob@example.com>"
    main=op1prod

Section `host`
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

Of these, `addr` and `user` are mandatory. `port` is optional and defaults to
`22`. Either `password` or `key` must be specified so the login can be
completed noninteratively. In case `key` is used, it must contain a valid SSH
private key with newlines replaces by spaces. The key must not be locked by a
password.

### Example

    [host op1jump]
    addr=192.168.10.10
    user=admin
    password=3x4mpl3
    port=2222
    
    [host op1prod]
    addr=10.0.33.66
    user=admin
    key="-----BEGIN RSA PRIVATE KEY----- MIIEogIBAAKCAQEAxymzAVzTX6oJTlZ5uCkqjdrDb/ovLZ6VktH+i5h2wdJpyT3f s2Q23e ...etc"
    via=op1jump

Section `forward`
-----------------

The `forward` sections describes SSH port forwardings that will be set up when
the destination is reached. The description of the forward is set in the
section header after the `forward` keyword and may contains spaces and special
characters within reason. It's encourages to be as descriptive as possible so
that the tunnel definition is self documented and will be presented to the user
after connecting.

Each element withing the `forward` section is a pair on the form
`<local address> = <remote address>`. The local side can use addresses other than
127.0.0.1 but still in the 127.0.0.0/8 block; these will be added to the local
loopback interface if they don't already exist.

If there is no SSH configuration, but there is a VPN configuration, then the
forwards will be done from the local computer.  This can be used to provide the
user with the same usage pattern as in the SSH forward case and also keep the
tunnel definition self documenting.

### Example

    [forward The Globe units]
    127.0.0.1:22000=10.0.33.69.193:22000
    127.0.0.1:22001=10.0.33.69.193:22001
    127.0.0.1:22002=10.0.33.69.193:22002
    127.0.0.2:22000=10.0.33.70.194:22000
    127.0.0.2:22001=10.0.33.70.194:22001
    127.0.0.2:22002=10.0.33.70.194:22002

    [forward Albert Hall units]
    127.0.0.3:22000=10.2.34.91:22000
    127.0.0.3:22001=10.2.34.91:22001
    127.0.0.3:22002=10.2.34.91:22002
    127.0.0.4:22000=10.2.34.92:22000
    127.0.0.4:22001=10.2.34.92:22001
    127.0.0.4:22002=10.2.34.92:22002

Section `vpnc`
--------------

The `vpnc` section defines a configuration for the vpnc Cisco VPN command line
client. The elements are any configuration directives recognized by vpnc, with
spaces replaced by underscores. Element names cannot contain special characters
such as paranthesis, but since there is no equal sign or similar in a vpnc
configuration a line like

    DPD idle timeout (our side) 0

can be represented in the tunnel definition as

    DPD_idle_timeout="(our side) 0"

The configuration must contain Xauth username and password since it must be
able to connect noninteractively.

The `vpnc` section is optional and requires that vpnc be installed if present.
If present, the VPN will be connected before any attempts are made to connect
to hosts defined as above.

### Example

    [vpnc]
    IPSec_gateway=213.154.23.72
    IPSec_ID=IPSECGROUP
    IPSec_secret=abrakadabra
    Xauth_username=extuser
    Xauth_password=K0ssanmu7
    IKE_Authmode=psk
    DPD_idle_timeout="(our side) 0"
    NAT_Traversal_Mode=force-natt
    Local_Port=0
    Cisco_UDP_Encapsulation_Port=0

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
    10.200.0.0=16
    192.168.10.0=24
    192.168.20.0=24
