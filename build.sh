#!/bin/bash

builddate=$(TZ=UTC date +"%Y-%m-%d %H:%M:%S %Z")
buildver=$(git describe --always)
builduser="$(whoami)@$(hostname)"
ldflags="-X main.buildDate '$builddate' -X main.buildVersion '$buildver' -X main.buildUser '$builduser'"

go build -ldflags "$ldflags" nym.se/mole/cmd/mole \
	&& mv mole $GOPATH/bin/mole4
