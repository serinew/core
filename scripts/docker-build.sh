#!/usr/bin/env bash
# Jenkins / CI용 Docker 빌드. BuildKit 로컬 캐시로 go build 재컴파일 시간을 줄입니다.
#
# Jenkins Execute shell 예:
#   bash scripts/docker-build.sh
#
# 환경 변수(선택):
#   IMAGE_NAME=archive IMAGE_TAG=latest
#   DOCKER_BUILD_CACHE_DIR=/var/jenkins_home/caches/p-archive-buildkit
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export DOCKER_BUILDKIT=1

IMAGE_NAME="${IMAGE_NAME:-archive}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
CACHE_DIR="${DOCKER_BUILD_CACHE_DIR:-/var/jenkins_home/caches/p-archive-buildkit}"

mkdir -p "$CACHE_DIR"

BUILDER_NAME="p-archive-builder"
if ! docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
	docker buildx create --name "$BUILDER_NAME" --driver docker-container --use
else
	docker buildx use "$BUILDER_NAME"
fi

echo ">>>>> build image (buildx cache: ${CACHE_DIR})"
docker buildx build --load \
	--cache-from "type=local,src=${CACHE_DIR}" \
	--cache-to "type=local,dest=${CACHE_DIR},mode=max" \
	-t "${IMAGE_NAME}:${IMAGE_TAG}" \
	-f Dockerfile \
	.
