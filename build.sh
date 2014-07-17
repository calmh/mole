#!/bin/bash

pkg=github.com/calmh/mole
buildstamp=$(date +%s)
buildver=$(git describe --always --dirty)
builduser="$(whoami)@$(hostname)"
ldflags="-w -X main.buildStamp '$buildstamp' -X main.buildVersion '$buildver' -X main.buildUser '$builduser'"

export GOBIN=$(pwd)/bin

buildClient() {
	rm -rf auto
	mkdir auto
	for arch in linux-386 linux-amd64 darwin-amd64 ; do
		echo "$arch"
		export GOOS=${arch%-*}
		export GOARCH=${arch#*-}
		rm -rf bin
		godep go install -ldflags "$ldflags" "$pkg/cmd/mole"
		tar zcf "mole-$arch.tar.gz" bin

		[ -f bin/mole ] && mv bin/mole "auto/mole-$arch"
		[ -f bin/*/mole ] && mv bin/*/mole "auto/mole-$arch"
		hash=$(sha1sum auto/mole-$arch | awk '{print $1}')
		echo "{\"buildstamp\":$buildstamp, \"version\":\"$buildver\", \"hash\":\"$hash\"}" >> "auto/mole-$arch.json"
		gzip -9 "auto/mole-$arch"
	done
	for arch in windows-386 windows-amd64 ; do
		echo "$arch"
		export GOOS=${arch%-*}
		export GOARCH=${arch#*-}
		rm -rf bin
		godep go install -ldflags "$ldflags" "$pkg/cmd/mole"
		zip -qr "mole-$arch.zip" bin
	done
	rm -rf bin
}

buildServer() {
	rm -rf srv bin
	mkdir srv
	source /usr/local/golang-crosscompile/crosscompile.bash
	for arch in linux-386 linux-amd64 darwin-amd64 ; do
		echo "$arch"
		"go-$arch" install -ldflags "$ldflags" "$pkg/cmd/molesrv"
		[ -f bin/molesrv ] && mv bin/molesrv "srv/molesrv-$arch"
		[ -f bin/*/molesrv ] && mv bin/*/molesrv "srv/molesrv-$arch"
	done
	rm -rf bin
	tar zcf molesrv-all.tar.gz srv
}

case $1 in
	all)
		rm -fr "$GOPATH"/pkg
		godep go test ./... 
		echo
		echo Client
		echo
		buildClient
		echo
		echo Server
		echo
		buildServer
		;;
	*)
		godep go install -ldflags "$ldflags" "$pkg/cmd/..."
		;;
esac
