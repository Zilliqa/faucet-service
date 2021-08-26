GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/build

.PHONY: test cover deps build start

deps:
	go mod download 

test:
	go test ./... -cover -covermode=atomic -coverprofile=coverage.txt

cover:
	go tool cover -html=coverage.txt
	
build:
	CGO_ENABLED=0 GOOS=linux go build -o $(GOBIN)/faucet-service ./cmd

start:
	make build && docker build -t faucet-service . -f Dockerfile && docker run -p 8080:8080 --env-file ./.env faucet-service
