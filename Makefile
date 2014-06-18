NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
BIN_NAME=goop

all: build

build:
	@mkdir -p build/
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	go get -d -v ./...
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	go install -x ./...
	cp $(GOPATH)/bin/$(BIN_NAME) build/$(BIN_NAME)

format:
	go fmt ./...

test:
	@echo "$(OK_COLOR)==> Testing...$(NO_COLOR)"
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d
	@ginkgo -r -trace -keepGoing

.PHONY: all build format test
