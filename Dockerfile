FROM ubuntu:14.04

MAINTAINER Dmitry Pokidov <dooman87@gmail.com>

WORKDIR /

ENV GO15VENDOREXPERIMENT=0
ENV PATH=/opt/go/bin:/gohome/bin:$PATH
ENV GOPATH=/gohome
ENV GOROOT=/opt/go

RUN apt-get update && \
    apt-get -y install automake bison flex g++ git libboost1.54-all-dev libevent-dev libssl-dev libtool make \
    pkg-config wget \
    python-mapnik libmapnik-dev \
    protobuf-compiler libprotobuf-dev \
    jq cmake libncurses5-dev && \
    rm -rf /var/lib/apt/lists/*

#Installing apache thrift
RUN wget http://apache.uberglobalmirror.com/thrift/0.9.3/thrift-0.9.3.tar.gz && \
    tar -xzvf thrift-0.9.3.tar.gz && \
    rm ./thrift-0.9.3.tar.gz && \
    cd /thrift-0.9.3 && \
    ./configure --without-java && \
    make && \
    make install

#Install golang 1.6
WORKDIR /opt
RUN wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz && \
    tar -xzvf go1.6.linux-amd64.tar.gz && \
    rm go1.6.linux-amd64.tar.gz && \
    mkdir -p /gohome/bin && \
    mkdir -p /gohome/pkg && \
    mkdir /gohome/src

RUN go get github.com/mattn/gom

RUN mkdir /gopnik
ADD . /gopnik
WORKDIR /gopnik
RUN gom install && \
    gom exec ./bootstrap.bash && \
    gom exec ./build.bash

RUN mkdir /gopnik_data
COPY example/dockerconfig.json /gopnik_data/config.json
COPY sampledata/stylesheet.xml /gopnik_data/
COPY sampledata/world_merc.shp /gopnik_data/
COPY sampledata/world_merc.dbf /gopnik_data/
VOLUME /gopnik_data

EXPOSE 8080
EXPOSE 9090

ENTRYPOINT ./entrypoint.sh
