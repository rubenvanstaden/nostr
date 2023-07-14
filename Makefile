SHELL := /usr/bin/env bash

up:
	docker-compose up -d

relay:
	go run ./cmd/relay/main.go

build:
	go build -o ./bin/ncli ./cmd/cli/* ./cli/*

fmt:
	go mod tidy -compat=1.17
	gofmt -l -s -w .

test:
	go test ./...
