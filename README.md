
====

Elevator Pitch
--------------

Like 1Password for ssh tunnels and VPN connections, plus sharing within a team.

Actual Documentation
--------------------

On the [GitHub Wiki](https://github.com/calmh/mole/wiki).

What
----

Mole lets you seamlessly share ssh and Cisco VPN configurations so that all you
need to know is the name you want to connect to. Mole will sort out any needed
passwords, keys and tunnelings and set up a bunch of forwardings for you.

It'll tell you what forwardings are available and what they point to, so you
can get on with real work and not have to dig around for infrastructure
information.

It's based around *tunnel definitions* which are self contained recipes that
describe how to connect to a customer or site.  A tunnel definition contains:

  - Possibly, a Cisco VPN configuration.

  - Probably, one or more host definitions (name, address, username).

  - A password or SSH key for the host. This is not optional, tunnel
    definitions should be able to connect without user interaction.

  - A description of how to chain the hosts together to reach the final
    destination, i.e. jump via host A to B, from B to C and from there to D.

  - A set of port forwarding descriptions to set up once the destination is
    reached, with commentary on what they're for.

The tunnel definitions live server side.

All server communication is ticket authenticated and secured by TLS.

The end result of all this is that as long as you have mole installed and
someone has written a tunnel definition, you can just `mole dig foobar` to
connect all the way and get a nice list of available port forwardings
presented to you.

Building
--------

Install `gb` (https://github.com/constabulary/gb)

Run `./build.sh` to create the binary.

Run `./build.sh all` to create the distribution packages.

License
-------

MIT
