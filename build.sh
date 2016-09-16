#!/bin/bash
set -euo pipefail

pkg=github.com/calmh/mole
buildstamp=$(date +%s)
buildver=$(git describe --always --dirty)
builduser="$(whoami)@$(hostname)"
ldflags="-w -X main.buildStamp=$buildstamp -X main.buildVersion=$buildver -X main.buildUser=$builduser"

export GOBIN=$(pwd)/bin

buildClient() {
	rm -rf auto
	mkdir auto
	for arch in linux-386 linux-amd64 darwin-amd64 ; do
		echo "$arch"
		export GOOS=${arch%-*}
		export GOARCH=${arch#*-}
		rm -rf bin
		mkdir bin
		godep go build -ldflags "$ldflags" -o bin/mole "$pkg/cmd/mole"
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
		mkdir bin
		godep go build -ldflags "$ldflags" -o bin/mole.exe "$pkg/cmd/mole"
		zip -qr "mole-$arch.zip" bin
	done
	rm -rf bin
}

case ${1:-default} in
	all)
		rm -fr "$GOPATH"/pkg
		godep go test ./... 
		echo
		echo Client
		echo
		buildClient
		;;
	default)
		godep go install -ldflags "$ldflags" "$pkg/cmd/..."
		;;
esac
