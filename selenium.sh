#!/bin/bash
# Start/stop selenium server. Grab the server jar from $url


pidfile=/tmp/selenium.pid
log=/tmp/selenium.log
url=http://selenium.googlecode.com/files/selenium-server-standalone-2.5.0.jar

start() {
    java -jar selenium-server-standalone-2.5.0.jar > $log 2>&1 &
    echo $! > $pidfile
}

stop() {
    kill $(cat $pidfile)
}

download() {
    curl -LO $url
}

case $1 in
    -h | --help ) echo "usage: $(basename $0) start|stop|download"; exit;;
    start ) start;;
    stop ) stop;;
    download ) download;;
    * ) echo "error: unknown command - $1"; exit 1;;
esac
