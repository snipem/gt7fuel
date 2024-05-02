SHELL := /bin/bash
GIT_COMMIT := $(shell git rev-list -1 HEAD)

install:
	go build -ldflags "-X main.GitCommit=${GIT_COMMIT} -s -w" -o ${HOME}/bin gt7fuel.go

run:
	go run gt7fuel.go --race-time=60

test:
	go test ./lib

deps:
	git submodule update --init --recursive


run_dump:
	go run gt7fuel.go --dump-file testdata/gt7testdata/watkinsglen.gob.gz
