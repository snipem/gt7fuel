SHELL := /bin/bash
GIT_COMMIT := $(shell git rev-list -1 HEAD)

install:
	go build -ldflags "-X main.GitCommit=${GIT_COMMIT} -s -w" -o ${HOME}/bin gt7fuel.go
