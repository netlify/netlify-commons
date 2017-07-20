#!/usr/bin/env bash

set -e
set -x

WORKSPACE=/go/src/github.com/netlify/netlify-commons

docker run \
	--volume $(pwd):$WORKSPACE \
	--workdir $WORKSPACE \
	--rm \
	calavera/go-glide:v0.12.3 script/test.sh
