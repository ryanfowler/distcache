#!/bin/sh
set -ex

PREFIX="/usr/local" && \
VERSION="0.52.0" && \
  curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m).tar.gz" | \
    sudo tar -xvzf - -C "${PREFIX}" --strip-components 1
