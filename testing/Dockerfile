FROM golang:1.8-stretch
MAINTAINER Miki Tebeka <miki.tebeka@gmail.com>

RUN apt-get update
RUN apt-get install -y xvfb openjdk-8-jre unzip libgconf-2-4 chromium iceweasel bzip2
VOLUME /code
ENV GOPATH /code
