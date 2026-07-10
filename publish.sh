#!/usr/bin/env sh
set -eu

VERSION=0.1.0
PACKAGE_NAME=devforth/letitcall
IMAGE=docker.io/$PACKAGE_NAME
PLATFORMS=${DOCKER_PLATFORMS:-linux/amd64,linux/arm64}

docker buildx build \
	--platform "$PLATFORMS" \
	--tag "$IMAGE:$VERSION" \
	--tag "$IMAGE:latest" \
	--push \
	.
