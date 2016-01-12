#!/bin/bash
# Run the test suite

# Show commands
set -x

jar=$(./selenium.sh jar)
if [ ! -f $jar ]; then
    ./selenium.sh download
    if [ $? -ne 0 ]; then
        exit $?
    fi
fi

./selenium.sh start
# Wait for selenium to start
max_wait=20
start_time=$(date +%s)
while true;
do
	now=$(date +%s)
	wait_time=$((now - start_time))
	curl -s http://127.0.0.1:4444/wd/hub > /dev/null
	if [ $? -eq 0 ]; then
		echo "selenium server started after ${wait_time} seconds"
		break
	fi
	if [ $wait_time -gt $max_wait ]; then
		echo "error: selenium server didn't start after ${max_wait} seconds"
		./selenium.sh stop
		exit 1
	fi
done

go test -v $@
value=$!
./selenium.sh stop
exit $value
