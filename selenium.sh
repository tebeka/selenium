#!/bin/bash
# Start/stop selenium server. Grab the server jar from $url


pidfile=/tmp/selenium.pid
log=/tmp/selenium.log
jar=selenium-server-standalone-2.39.0.jar
url=http://selenium.googlecode.com/files/$jar

start() {
    java -jar $jar > $log 2>&1 &
    echo $! > $pidfile
}

stop() {
    kill $(cat $pidfile)
}

download() {
    curl -LO $url
}

if [ $# -ne 1 ]; then
    echo "error: wrong number of arguments"
    $0 -h
    exit 1
fi

case $1 in
    -h | --help ) echo "usage: $(basename $0) start|stop|download"; exit;;
    start ) start;;
    stop ) stop;;
    download ) download;;
    jar ) echo $jar;;
    * ) echo "error: unknown command - $1"; exit 1;;
esac
