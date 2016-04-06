FROM ubuntu:14.04

WORKDIR /

#Installing apache thrift
RUN apt-get update
RUN apt-get -y install automake bison flex g++ git libboost1.54-all-dev libevent-dev libssl-dev libtool make pkg-config wget
RUN wget http://apache.uberglobalmirror.com/thrift/0.9.3/thrift-0.9.3.tar.gz
RUN tar -xzvf thrift-0.9.3.tar.gz
WORKDIR /thrift-0.9.3
RUN ./configure --without-java
RUN make
RUN make install

#Install mapnik
RUN apt-get -y install python-mapnik libmapnik-dev

#Install protobuf
RUN apt-get -y install protobuf-compiler

#Install golang
WORKDIR /opt
RUN wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz 
RUN tar -xzvf go1.6.linux-amd64.tar.gz
RUN mkdir /gohome
RUN mkdir /gohome/bin
RUN mkdir /gohome/pkg
RUN mkdir /gohome/src
ENV PATH=/opt/go/bin:/gohome/bin:$PATH
ENV GOPATH=/gohome
ENV GOROOT=/opt/go

RUN go get github.com/mattn/gom

RUN mkdir /gopnik
ADD . /gopnik
WORKDIR /gopnik
#RUN gom install
#RUN apt-get -y install golang

#RUN apk add --update git bash ncurses protobuf automake bison flex g++
#RUN go get github.com/mattn/gom
#VOLUME /gopnik/

#VOLUME /go/src/
