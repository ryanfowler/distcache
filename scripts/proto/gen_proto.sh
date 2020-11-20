#!/bin/sh
set -e

if ! docker info &> /dev/null; then
    echo 'Error: Docker is required to be running'
    exit 1
fi

docker build -f scripts/proto/Dockerfile -t distcache . &> /dev/null
docker run -it --rm --name=distcache -v "$(pwd)":/go/src/github.com/ryanfowler/distcache distcache protoc -I . --go_out=:grpc/peerpb --go-grpc_out=grpc/peerpb grpc/peerpb/peer.proto &> /dev/null
