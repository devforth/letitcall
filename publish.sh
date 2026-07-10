#!/usr/bin/env sh
set -eu

VERSION=${1:-}
IMAGE=${2:-${DOCKER_IMAGE:-ghcr.io/letitcall/letitcall}}
PLATFORMS=${DOCKER_PLATFORMS:-linux/amd64,linux/arm64}

if [ -z "$VERSION" ]; then
	echo "Usage: ./publish.sh <version> [image]" >&2
	echo "Example: ./publish.sh 0.1.0 ghcr.io/example/letitcall" >&2
	exit 2
fi

case "$VERSION" in
	latest)
		echo "Version must be a release version, not 'latest'." >&2
		exit 2
		;;
esac

docker buildx build \
	--platform "$PLATFORMS" \
	--tag "$IMAGE:$VERSION" \
	--tag "$IMAGE:latest" \
	--push \
	.
