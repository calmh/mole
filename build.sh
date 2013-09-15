#!/bin/bash

buildstamp=$(date +%s)
buildver=$(git describe --always)
builduser="$(whoami)@$(hostname)"
ldflags="-X main.buildStamp '$buildstamp' -X main.buildVersion '$buildver' -X main.buildUser '$builduser'"

export GOPATH=$(pwd):$GOPATH
export GOBIN=$(pwd)/bin

if [[ $1 == "all" ]] ; then
	source /usr/local/golang-crosscompile/crosscompile.bash
	for arch in linux-386 linux-amd64 darwin-amd64 windows-386 windows-amd64 ; do
		echo $arch
		rm -rf bin
		go-$arch install -ldflags "$ldflags" nym.se/mole/cmd/...
		tar zcf mole-$arch.tar.gz bin
	done
else
	go install -ldflags "$ldflags" nym.se/mole/cmd/...
fi

