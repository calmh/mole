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

Mole is an ssh tunnel manager with sweet features for teams and a thick layer of pure awesome.

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

