FROM --platform=linux/amd64 golang:latest
RUN apt -y update
RUN apt -y upgrade
RUN apt install -y protobuf-compiler
RUN apt install -y vim
RUN apt install -y unzip

WORKDIR /root
ENV PROTOC_VERSION 21.6
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
RUN unzip -o protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local bin/protoc
RUN unzip -o protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local 'include/*'
RUN rm -f protoc-${PROTOC_VERSION}-linux-x86_64.zip
RUN git clone https://github.com/googleapis/googleapis.git
RUN git clone https://github.com/grpc-ecosystem/grpc-gateway.git
COPY Makefile /root
RUN make install-golibs
ENV TERM=xterm-256color