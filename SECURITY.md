Mole's Security Posture
=======================

The primary goal of mole is to be *convenient*. The secondary goal is to
prevent accidental leakage of sensitive information.

In the environment were mole grew up, mole replaced a shared text file with
plain text usernames and passwords. It's not hard to improve on that lever of
security.

In the current version of mole, credentials are never comitted to stable
storage. They are sent encrypted from the server, decrypted locally and used
to authenticate for example SSH sessions. This means accidental disclosure of
sensitive information is unlikely. If your laptop is stolen and unlocked, it
should not be possible for the thief to access information from the mole
system[1].

It does not mean that a determined attacker with access to the system can't
extract sensitive information. The credentials must, due to the nature of
things, be available on the local computer to perform authentication towards
remote parties. This means it's possible to extract the keys from memory or
subvert the code that uses them. It's an open source project, so the latter
isn't even particularly difficult. This is the "DRM problem" -- it's
impossible to prevent someone from copying a movie while permitting them to
see it.

[1] Unless there is an active session ticket and they can get the computer
online on the same IP (from the mole server's point of view) it had before.
In that situation convenience trumps security and we lose.
