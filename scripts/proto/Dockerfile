FROM golang:1.22
RUN apt-get update && apt-get install -y unzip

ENV PROTOBUF_VERSION 26.1
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOBUF_VERSION/protoc-$PROTOBUF_VERSION-linux-x86_64.zip && \
    unzip -o protoc-$PROTOBUF_VERSION-linux-x86_64.zip -d /usr/local bin/protoc && \
    unzip -o protoc-$PROTOBUF_VERSION-linux-x86_64.zip -d /usr/local 'include/*'

ENV PROTOC_GEN_GO_VERSION 1.33.0
ENV PROTOC_GEN_GO_GRPC_VERSION 1.3.0
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v$PROTOC_GEN_GO_VERSION
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$PROTOC_GEN_GO_GRPC_VERSION

WORKDIR /go/src/github.com/ryanfowler/distcache
