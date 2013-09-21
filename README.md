                      ___
                     /\_ \
      ___ ___     ___\//\ \      __
    /' __` __`\  / __`\\ \ \   /'__`\
    /\ \/\ \/\ \/\ \L\ \\_\ \_/\  __/
    \ \_\ \_\ \_\ \____//\____\ \____\
     \/_/\/_/\/_/\/___/ \/____/\/____/

mole
====

Elevator Pitch
--------------

Like 1Password for ssh tunnels and VPN connections, plus sharing within a team.

What
----

Mole lets you seamlessly share ssh and Cisco VPN configurations so that all you
need to know is the name you want to connect to. Mole will sort out any needed
passwords, keys and tunnelings and set up a bunch of forwardings for you.

It'll tell you what forwardings are available and what they point to, so you
can get on with real work and not have to dig around for inrastructure
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

Quick start
-----------

Your admin or colleague gave you a hostname and a token? Do this:

 1. If you don't have it, get a mole binary for your platform.

 2. Register with the server.

        $ mole register <hostname>

 3. Check what tunnel definitions are available. This will require you to log
    in, which will grant you a ticket. The ticket is valid for requests from
    the same IP for a configurable time period (a week, by default).

        $ mole ls

 4. Connect to a tunnel:

        $ sudo mole dig <tunnelname>

Write a new tunnel definition
-----------------------------

 1. Read through the [Configuration Reference](https://github.com/calmh/mole/blob/master/CONFIGURATION.md).

 2. Create a tunnel definition file with an `.ini` extension in your home directory.

 3. Test the tunnel definition:

        $ mole dig -l whateverfile.ini

 4. When you're happy, push it to the server:

        $ mole push whateverfile.ini

Edit an existing tunnel definition
----------------------------------

 1. Export the existing tunnel config to a local file:

        $ mole show -r whatever > whatever.ini

 2. Edit the file `whatever.ini` to taste.

 3. Test the tunnel definition:

        $ mole dig -l whateverfile.ini

 4. When you're happy, push it to the server:

        $ mole push whateverfile.ini

License
-------

MIT

