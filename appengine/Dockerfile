FROM gcr.io/google-appengine/golang
MAINTAINER Eric Garrido <eric@ericgar.com>

RUN echo 'deb http://httpredir.debian.org/debian jessie-backports main' > /etc/apt/sources.list.d/jessie-backports.list
RUN apt-get -q update -y && apt-get -y -q --no-install-recommends install \
      bzip2 \
      ca-certificates-java=20161107'*' \
      chromium \
      git \
      iceweasel \
      libgconf-2-4 \
      openjdk-8-jre-headless \
      unzip \
      xauth \
      xvfb

ENV PATH /usr/bin:$PATH
COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
