FROM golang:1.6
MAINTAINER Miki Tebeka <miki.tebeka@gmail.com>


ENV DISPLAY :99
RUN apt-get update
RUN apt-get install -y xvfb iceweasel openjdk-7-jre-headless
VOLUME /code
