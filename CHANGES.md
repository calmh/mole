Changes from 2.3 to 3.0
=======================

 - The client no longer keeps tunnel definitions in ~/.mole/tunnels on
   the client. Tunnel definitions are loaded on demand from the server.
   To access a tunnel definition for editing etc use the `export`
   command. The `pull` command is deprecated.

 - Tunnel definitions no longer contain credentials in cleartext.
   Cleartext credentials are converted to token equivalents when a
   tunnel definition is pushed. Tokens are resolved to their cleartext
   equivalents on demand by the server.

 - Forwarding directives can default the remote port and use ranges on
   the local side. Example: `127.0.0.1:42000-42009 = 10.1.2.3`

 - The configuration format has changed slightly to enable the use of a
   much improved `.ini` file parser. Specifically:

   * There is a new a mandatory `version` field in the general section
     that must be set to `3`.

   * The `host` and `forward` keywords has been changed to `hosts` and
     `forwards`.

   * The section headers use dots to separate keywords from names, i.e.
     `[hosts.foo]` instead of '[host foo]'. Any other dots need to be
     escaped: `[host.example\.com]`.

   The older configuration format is still understood by the client, but
   will be converted to version 3 format by the server when pushed.

 - Error checking of tunnel definitions is improved, catching duplicate
   forward definitions that would previously fail silently.

 - The v3 client is compatible with the v2 server but the `list` or `ls`
   commands are limited to displaying tunnel names only.

 - The v3 server requires a v3 client.

