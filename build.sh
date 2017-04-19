#!/bin/sh

APP_NAME='git-coin'
APP_REPO="github.com/tokenshift/git-coin"

IFS='/'

set -x

mkdir release
go tool dist list | while read os arch; do
	env GOOS=$os GOARCH=$arch go build -o "./release/${APP_NAME}.${os}_${arch}" "$APP_REPO"
done
