FROM golang:1.15
RUN apt-get update && apt-get install -y unzip

ENV PROTOBUF_VERSION 3.14.0
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOBUF_VERSION/protoc-$PROTOBUF_VERSION-linux-x86_64.zip && \
    unzip -o protoc-$PROTOBUF_VERSION-linux-x86_64.zip -d /usr/local bin/protoc && \
    unzip -o protoc-$PROTOBUF_VERSION-linux-x86_64.zip -d /usr/local 'include/*'

ENV PROTOC_GEN_GO_VERSION 1.25
ENV PROTOC_GEN_GO_GRPC_VERSION 1.0.1
RUN GO111MODULE=on go get \
    google.golang.org/protobuf/cmd/protoc-gen-go@v$PROTOC_GEN_GO_VERSION \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$PROTOC_GEN_GO_GRPC_VERSION

WORKDIR /go/src/github.com/ryanfowler/distcache
