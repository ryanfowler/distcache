#!/bin/sh
set -e

if ! docker info > /dev/null 2>&1; then
    echo 'Error: Docker is required to be running'
    exit 1
fi

docker build -f scripts/proto/Dockerfile -t distcache .
docker run -it --rm --name=distcache -v "$(pwd)":/go/src/github.com/ryanfowler/distcache distcache protoc -I . --go_out=:grpc/peerpb/v1 --go-grpc_out=grpc/peerpb/v1 grpc/peerpb/v1/peer.proto
