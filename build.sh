#!/usr/bin/env bash

export GOPROXY=https://mirrors.aliyun.com/goproxy/
export GO112MODULE=on
export GO111MODULE=on

cd waf-gate && go install
cd ../waf-server && go install