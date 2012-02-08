export GOPATH := $(PWD)

all:
	go build selenium

test:
	@echo WARNING: You probably want to use run-tests.sh
	go test -v selenium

install:
	go install selenium

.PHONY: all test install
