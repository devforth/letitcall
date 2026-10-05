#!/usr/bin/env sh
set -eu

VERSION=1.0.2
PACKAGE_NAME=devforth/letitcall
IMAGE=docker.io/$PACKAGE_NAME
PLATFORMS=${DOCKER_PLATFORMS:-linux/amd64,linux/arm64}

docker buildx build \
	--platform "$PLATFORMS" \
	--tag "$IMAGE:$VERSION" \
	--tag "$IMAGE:latest" \
	--push \
	.
