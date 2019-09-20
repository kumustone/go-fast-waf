#!/usr/bin/env bash

export GOPROXY=https://mirrors.aliyun.com/goproxy/
export GO112MODULE=on
export GO111MODULE=on

go build waf-gate/waf_gate.go
go build waf-server/waf_server.go

