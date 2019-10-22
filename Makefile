#!/usr/bin/make -f

export GO111MODULE = on

NetworkType := $(shell if [ -z ${NetworkType} ]; then echo "mainnet"; else echo ${NetworkType}; fi)

ldflags = -X github.com/chengwenxi/cosmos-relayer/chains/config.NetworkType=${NetworkType}

install:
	go install -ldflags '$(ldflags)' ./relayer.go
