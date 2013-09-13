#!/bin/bash

buildstamp=$(date +%s)
buildver=$(git describe --always)
builduser="$(whoami)@$(hostname)"
ldflags="-X main.buildStamp '$buildstamp' -X main.buildVersion '$buildver' -X main.buildUser '$builduser'"

go build -ldflags "$ldflags" nym.se/mole/cmd/mole \
	&& mv mole $GOPATH/bin/mole4
