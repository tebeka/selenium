#!/bin/bash
# Run tests in golang docker container

go get -d -v
pushd vendor
go get -d -v
go run init.go --alsologtostderr --download_browsers
popd
go test -test.v --running_under_docker
