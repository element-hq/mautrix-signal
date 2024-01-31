#!/bin/sh
export MAUTRIX_VERSION=$(cat go.mod | grep 'github.com/element-hq/mautrix-go ' | awk '{ print $2 }')
export GO_LDFLAGS="-s -w -X main.Tag=$(git describe --exact-match --tags 2>/dev/null) -X main.Commit=$(git rev-parse HEAD) -X 'main.BuildTime=`date '+%b %_d %Y, %H:%M:%S'`' -X 'github.com/element-hq/mautrix-go.GoModVersion=$MAUTRIX_VERSION'"
go build -ldflags "$GO_LDFLAGS" -o mautrix-signal
