export GOPATH := $(shell dirname $(shell dirname $(PWD)))
PACKAGE := selenium

all:
	go build $(PACKAGE)

test:
	@echo WARNING: You probably want to use run-tests.sh
	go test -v $(PACKAGE)

test-docker:
	docker build -t go-selenium .
	docker run \
	    -v ${PWD}:/code \
	    -w /code \
	    go-selenium ./docker-test.sh

doc:
	go doc $(PACKAGE)

install:
	go install $(PACKAGE)

README.html: README.md
	markdown $< > $@

.PHONY: all test install fix doc test-docker
