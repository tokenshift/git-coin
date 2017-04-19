#!/bin/sh

APP_NAME='git-coin'
APP_REPO="github.com/tokenshift/git-coin"

IFS='/'

set -x

go tool dist list | while read os arch; do
	mkdir -p "./release/${os}_${arch}/"
	env GOOS=$os GOARCH=$arch go build -o "./release/${os}_${arch}/$APP_NAME" "$APP_REPO"
done
