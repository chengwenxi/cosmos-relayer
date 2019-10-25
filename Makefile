#!/usr/bin/make -f

export GO111MODULE = on

install:
	go install  ./relayer.go
