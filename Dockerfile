FROM ubuntu:14.04

WORKDIR /

#Installing apache thrift
RUN apt-get update
RUN apt-get -y install automake bison flex g++ git libboost1.55-all-dev libevent-dev libssl-dev libtool make pkg-config wget
RUN wget http://apache.uberglobalmirror.com/thrift/0.9.3/thrift-0.9.3.tar.gz
RUN tar -xzvf thrift-0.9.3.tar.gz
WORKDIR /thrift-0.9.3
RUN ./configure --without-java
RUN make
RUN make install

#Install mapnik
RUN apt-get -y install python-mapnik

#Install golang
RUN apt-get -y install golang

#RUN apk add --update git bash ncurses protobuf automake bison flex g++
#RUN go get github.com/mattn/gom
#VOLUME /gopnik/
VOLUME /go/src/