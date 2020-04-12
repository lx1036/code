#!/usr/bin/env bash
cd /go/src || exit
go mod init "k8s-lx1036"
go mod vendor
go run /go/src/main.go
