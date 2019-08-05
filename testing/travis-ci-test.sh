#!/bin/bash
# Run tests under Travis for continuous integration.

go test -coverprofile=coverage.txt -covermode=atomic -test.v -timeout=20m ./...
