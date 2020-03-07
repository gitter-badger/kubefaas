FROM ubuntu:18.04

RUN apt update && \
    apt install -y software-properties-common && \
    add-apt-repository ppa:longsleep/golang-backports

RUN apt update && \
    apt install -y golang-1.14 \
                   parallel \
                   coreutils \
                   git \
                   apt-transport-https \
                   ca-certificates \
                   curl \
                   gnupg-agent \
                   curl \
                   wget \
                   unzip \
                   zip

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"

RUN apt-get update && apt-get install -y docker-ce docker-ce-cli containerd.io
RUN mkdir /go/

ENV GOROOT /usr/lib/go-1.14/
ENV GOPATH /go/
