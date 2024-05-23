LOCAL_BIN:=$(CURDIR)/bin

GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
GOLANGCI_TAG:=1.56.2
MOCKGEN_TAG:=latest
SMART_IMPORTS := ${LOCAL_BIN}/smartimports


# build app
.PHONY: build
build:
	go mod download && CGO_ENABLED=0  go build \
		-o ./bin/main$(shell go env GOEXE) ./cmd/main.go


.PHONY: run-dev
run-dev:
	CMD_TOPIC=frames CMD_RESPONSE_TOPIC=responses RESULT_TOPIC=frames-splitted KAFKA_BROKER=127.0.0.1:9091 go run cmd/main.go
