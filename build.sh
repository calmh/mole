#!/bin/bash
set -euo pipefail

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
		mkdir -p bin
		go build -i -v -ldflags "$ldflags" -o "bin/mole-$arch" ./cmd/mole
		tar zcf "mole-$arch.tar.gz" bin

		mv bin/* auto
		hash=$(shasum auto/mole-$arch | awk '{print $1}')
		echo "{\"buildstamp\":$buildstamp, \"version\":\"$buildver\", \"hash\":\"$hash\"}" >> "auto/mole-$arch.json"
		gzip -9 "auto/mole-$arch"
	done
	for arch in windows-386 windows-amd64 ; do
		echo "$arch"
		export GOOS=${arch%-*}
		export GOARCH=${arch#*-}
		rm -rf bin
		mkdir -p bin
		go build -i -v -ldflags "$ldflags" -o "bin/mole-$arch.exe" ./cmd/mole
		zip -qr "mole-$arch.zip" bin
	done
	rm -rf bin
}

buildServer() {
	rm -rf srv
	mkdir srv
	for arch in linux-386 linux-amd64 darwin-amd64 ; do
		echo "$arch"
		export GOOS=${arch%-*}
		export GOARCH=${arch#*-}
		go build -i -v -ldflags "$ldflags" -o "srv/molesrv-$arch" ./cmd/molesrv
	done
	tar zcf molesrv-all.tar.gz srv
}

case ${1:-default} in
	all)
		buildClient
		buildServer
		;;

	client)
		buildClient
		;;

	server)
		buildServer
		;;

	default)
		gb build -ldflags "$ldflags"
		;;
esac
