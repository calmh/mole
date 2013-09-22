Automatic Upgrade Process
=========================

## Configuration

Mole looks for the `automatic` key in the `[upgrades]` section of
`~/.mole/mole.ini`. The presence and value of this configuration key results
in one of three behaviors:

 - Present, value `yes`: automatic upgrade is attempted. In case of success,
   a terse message informing the user about the upgrade is printed.

 - Present, any other value: no automatic upgrade is attempted.

 - Absent: automatic upgrade is attempted. In case of success, a message
   informing the user about the upgrade is printed together with an
   instruction about the relevant configuration setting to permit or disable
   automatic upgrades.

Errors that occur during automatic upgrade are not reported the user; the
upgrade attempt will silently fail and be retried at some point in the future.
The upgrade process when the user runs `mole upgrade` is the same as the one
below, except that errors *are* reported.

## Upgrade Process

### Get Upgrade Manifest

The file /extra/upgrades.json is requested from the server. Files in the
/extra directory does not require authentication so this will always succeed
if the file is present. If this returns an HTTP error or the manifest is
invalid the upgrade process terminates. The upgrade manifest is a JSON object
containing single `url` attribute:

    {"url": "http://some-server.example.com/whatever/directory"}

### Get Build Manifest

Mole takes the URL acquired above, a filename base formed by appending the
current operating system and architecture to the word mole with dashes and
requests a JSON file names as such. The OS/architecture tags can be seen by
running `mole version`. For example, on Mac OS X, the filename base will be
`mole-darwin-amd64` and the request will be for the URL:

    http://some-server.example.com/whatever/directory/mole-darwin-amd64.json

This is expected to return a build manifest, created by the build.sh script:

    {"buildstamp": 1379837190,
     "version": "v4.0.0-dev-57-g573c760",
     "hash": "58406302840fe53dcbaf4909eb956193071e6af4"}

The `buildstamp` is an epoch timestamp of when the build was created. This is
compared to the build stamp of the currently running binary. If the server
build stamp is higher, an upgrade is attempted. If the build manifest is
absent or invalid, the upgrade process terminates. The `version` field is
informative only. The `hash` field is the SHA1 hash of the corresponding
uncompressed binary.

### Get Binary

Given the URL and filename base as above, mole tries to fetch a gzipped
binary (as created by the build script). In this example, the url would be:

    http://some-server.example.com/whatever/directory/mole-darwin-amd64.gz

This file is decompressed and saved to the same location as the currently
running mole binary, with the filename `mole.part` (assuming the current
binary is called `mole`). If this location is not writable, the upgrade
process terminates.

### Deploy Binary

Once downloaded and saved, the file is opened and read again to compute it's
SHA1 hash. This is compared to the hash in the build manifest above. If the
hashes don't match the file is removed and the upgrade process terminates. If
the hashes match, the file `mole.part` is renamed to `mole` and the upgrade
process is complete.
