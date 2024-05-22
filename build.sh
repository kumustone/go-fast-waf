#!/usr/bin/env bash

export GOPROXY=https://mirrors.aliyun.com/goproxy/
export GO112MODULE=on
export GO111MODULE=on

go build waf-gate/main.go
go build waf-server/main.go

