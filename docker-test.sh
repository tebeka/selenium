#!/bin/bash
# Run tests in golang docker container

export DISPLAY=:99
Xvfb ${DISPLAY}&
sleep 3  # Let xvfb time to start
./run-tests.sh
