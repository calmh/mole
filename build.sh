#!/bin/sh

builddate=$(TZ=UTC date +"%Y-%m-%d %H:%M:%S %Z")
buildver=$(git describe --always)
builduser="$(whoami)@$(hostname)"
ldflags="-X main.buildDate '$builddate' -X main.buildVersion '$buildver' -X main.buildUser '$builduser'"

go build -ldflags "$ldflags" -tags debug \
	&& mv mole $GOPATH/bin/mole-dbg
go build -ldflags "$ldflags" \
	&& mv mole $GOPATH/bin/mole
