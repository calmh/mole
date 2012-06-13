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

Mole is an `ssh` tunnel manager with sweet features for teams and a thick layer
of pure awesome.

It's based around *tunnel definitions* which are self contained recipes that
describe how to connect to a customer or site.  A tunnel definition contains:

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

How
---

    Usage: mole [options] [command]
    
    Commands:

      dig <destination>
      dig a tunnel to the destination
      
      list 
      list available tunnel definitions
      
      pull 
      get tunnel definitions from the server
      
      push <file>
      send a tunnel definition to the server
      
      register [options] <server> <token>
      register with a mole server
      
      gettoken 
      generate a new registration token
      
      newuser [options] <username>
      create a new user
      
      export <tunnel> <outfile>
      export tunnel definition to a file
      
      view <tunnel>
      show tunnel definition
    
    Options:
    
      -h, --help   output usage information
      -d, --debug  display debug information
    
    Examples:
    
      Register with server "mole.example.com" and a token:
        mole register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2
    
      List available tunnels:
        mole list
    
      Dig a tunnel to "operator3":
        mole dig operator3
    
      Fetch new and updated tunnel specifications from the server:
        mole pull

