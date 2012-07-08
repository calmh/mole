                      ___
                     /\_ \
      ___ ___     ___\//\ \      __
    /' __` __`\  / __`\\ \ \   /'__`\
    /\ \/\ \/\ \/\ \L\ \\_\ \_/\  __/
    \ \_\ \_\ \_\ \____//\____\ \____\
     \/_/\/_/\/_/\/___/ \/____/\/____/

mole
====

[![build status](https://secure.travis-ci.org/calmh/mole.png)](http://travis-ci.org/calmh/mole)

What
----

Mole is an `ssh` and Cisco VPN tunnel manager with sweet collaboration features
for teams and a thick layer of pure awesome.

It's based around *tunnel definitions* which are self contained recipes that
describe how to connect to a customer or site.  A tunnel definition contains:

  - Possibly, a Cisco VPN configuration.

  - One or more host definitions (name, address, username).

  - A password or SSH key for the host. This is not optional, tunnel
    definitions should be able to connect without user interaction.

  - A description of how to chain the hosts together to reach the final
    destination, i.e. jump via host A to B, from B to C and from there to D.

  - A set of port forwarding descriptions to set up once the destination is
    reached, with commentary on what they're for.

The tunnel definitions live server side with a local cache and are pushed and
pulled similarly to how a DVCS works. If you don't know about that you don't
need to care, just know that `mole pull` will grab any new tunnel definitions
from the server and store them in the local cache.

All server communication is certificate authenticated and secured by TLS.

The end result of all this is that as long as you have mole installed and
someone has written a tunnel definition, you can just `mole dig foobar` to
connect all the way and get a nice list of available port forwardings presented
to you.

Quick start
-----------

Your admin or colleague gave you a hostname and a token? Do this:

 1. If you don't have it, get [Node.js](http://www.nodejs.org/#download).

 2. Install mole:

    $ sudo npm -g install mole

 3. Register using the credentials you got:

    $ mole register <hostname> <token>

 4. Check what tunnel definitions are available:

    $ mole list

 5. Connect to a tunnel:

    $ mole dig <tunnelname>

Write a new tunnel definition
-----------------------------

 1. Read through the [Configuration Reference](https://github.com/calmh/mole/blob/master/CONFIGURATION.md).

 2. Create a tunnel definition file with an `.ini` extension in your home directory.

 3. Test the tunnel definition:

    $ mole dig whateverfile.ini

 4. When you're happy, push it to the server:

    $ mole push whateverfile.ini

 5. Pull it yourself so it gets cached in ~/.mole/tunnels:

    $ mole pull

Set up a new server
-------------------

 1. Create and start a server:

    $ mole server

 2. Create an admin user. This works without authentication because the user
    database is currently empty and the default host for an unregistered mole
    is localhost. Mole will spit out a one-use token that can be used to bind a
    computer to the admin account.

    $ mole newuser admin

 3. Register the current computer to the admin account:

    $ mole register localhost <token-from-above>

 4. Use the admin user to create further users. Give the produced token to each
    user so they can register.

    $ mole newuser foo

License
-------

Simplified BSD

