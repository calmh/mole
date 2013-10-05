Mole's Security Posture
=======================

The primary goal of mole is to be *convenient*. The secondary goal is to
prevent accidental leakage of sensitive information.

In the environment were mole grew up, mole replaced a shared text file
with plain text usernames and passwords. It's not hard to improve on
that level of security.

When using the mole client, cleartext credentials are never comitted to
stable storage. They are sent encrypted from the server, decrypted
locally and used to authenticate for example SSH sessions. This means
accidental disclosure of sensitive information is unlikely. If your
laptop is stolen and unlocked, it should not be possible for the thief
to access information from the mole system.

This doesn't mean that a determined attacker with access to the system
can't extract sensitive information. The credentials must, due to the
nature of things, be available on the local computer to perform
authentication towards remote parties. This means it's possible to
extract the keys from memory or subvert the code that uses them. It's an
open source project, so the latter isn't even particularly difficult.

Communication
-------------

Communication between client and server is secured by HTTPS (TLS). Upon
registering, the client stores the server fingerprint in
`~/.mole/mole.ini`. On every subsequent request, the fingerprint is
verified against the stored value on connect. If the fingerprint doesn't
match (such as in a MITM or DNS spoofed situation) the connection is
closed and an error emitted before any information is sent over the
connection.

Authentication
--------------

Authentication is ticket based with an LDAP backend. Every client
request is authenticated by sending an opaque ticket in the
`X-Mole-Ticket` HTTP header. The ticket is verified by the server and if
it checks out the request is permitted. If it does not, a `403
Forbidden` is returned and the client will have to request a new ticket.

The ticket is a hashed (SHA1) and encrypted (AES256) blob containing the
authenticated username, the IP the client came from and a validity time
in the form of epoch seconds.

The ticket is considered valid by the server if it decrypts and hashes
correctly, the validity time is in the future and the client IP is as
set in the ticket. If any of these checks fail, authentication based on
the ticket is denied.

The session key used to encrypt the ticket is generated on server
startup and not saved. Restarting the server will generate a new session
key and thus invalidate all existing tickets.

To get a new ticket upon recieving a `403 Forbidden` response, the client
prompts the user for username and password and posts these (protected by
TLS as above) to the `/ticket` URL of the server. The server will attempt
a bind request with the specified credentials against a configured LDAP
server. If this succeeds, a ticket as above is generated and returned to
the client. If the bind request does not succeed, a `403 Forbidden`
response is returned.

The client stores it's current ticket in `~/.mole/mole.ini`. This file
should therefore have restrictive permissions. The client enforces this.

Key Obfuscation
---------------

Cleartext credentials in tunnel definitions are obfuscated when the
tunnel file is pushed to the server. The obfuscation is performed by
saving the actual credential in a separate index, keyed by a randomly
generated UUID, and replacing the credential in the file by this UUID.

The effect is that the tunnel definition, as can be seen and edited by
the user, no longer contains the cleartext credentials but contains a
reference that can be resolved to the actual credentials by the client.
When performing an action that requires the credentials (`dig`, for
example) the client will request the actual value for each obfuscated
credential from the server. The resulting credentials are not saved to
disk, but passed internally to the SSH library or over stdin to vpnc and
similar VPN providers.

While the standard mole client doesn't provide a method to show the
credentials, it is fairly trivial for an attacker to create such a
method themselves. Hence this entire scheme should be viewed as
obfuscation to prevent accidental disclosure rather than a secure way to
prevent an authorized user from gaining access to the credentials.

The configuration keys subject to obfuscation are `key` and `password`
in `[hosts]` sections, `IPSec_secret` and `Xauth_password` in `[vpnc]`
sections and `password` in `[openconnect]` sections.

