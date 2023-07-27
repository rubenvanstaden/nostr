SHELL := /usr/bin/env bash

up:
	docker-compose up -d

relay:
	go run ./cmd/relay/main.go

build:
	go build -o ./bin/ncli ./cmd/cli/*

fmt:
	go mod tidy -compat=1.17
	gofmt -l -s -w .

install:
	cp -f ./bin/ncli $(HOME)/go/bin/

test:
	go test ./...
