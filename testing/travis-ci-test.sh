#!/bin/bash
# Run tests under Travis for continuous integration.

go get -d -v
pushd vendor
go get -d -v
go run init.go
popd
# Travis has firefox already installed.
go test -test.v -test.run=TestFirefox --start_frame_buffer=false
