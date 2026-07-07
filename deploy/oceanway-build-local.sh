#!/usr/bin/env bash
# Build an OceanWay/Sub2API linux/amd64 Docker image locally on macOS.
#
# This is faster and more reliable on Apple Silicon than running the full
# Node/Go build inside an amd64 Docker VM: the host builds frontend assets and
# cross-compiles the Go binary, then Docker only packages the runtime image.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${REPO_ROOT}/tmp/oceanway-release"

COMMIT="${COMMIT:-$(git -C "${REPO_ROOT}" rev-parse --short HEAD)}"
VERSION="${VERSION:-$(tr -d '[:space:]' < "${REPO_ROOT}/backend/cmd/server/VERSION")}"
TAG="${TAG:-oceanway-$(date +%Y%m%d-%H%M)-${COMMIT}}"
DOCKER_CONTEXT="${DOCKER_CONTEXT:-colima-amd64}"
DOCKER_PLATFORM="${DOCKER_PLATFORM:-linux/amd64}"
DATE_VALUE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

echo "VERSION=${VERSION}"
echo "COMMIT=${COMMIT}"
echo "TAG=${TAG}"
echo "DOCKER_CONTEXT=${DOCKER_CONTEXT}"
echo "DOCKER_PLATFORM=${DOCKER_PLATFORM}"

cd "${REPO_ROOT}"

corepack pnpm --dir frontend install --frozen-lockfile
corepack pnpm --dir frontend run build

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

(
  cd backend
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags embed \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.Date=${DATE_VALUE} -X main.BuildType=release" \
    -o "${OUT_DIR}/sub2api" \
    ./cmd/server
)

cp "${REPO_ROOT}/deploy/docker-entrypoint.sh" "${OUT_DIR}/docker-entrypoint.sh"

docker --context "${DOCKER_CONTEXT}" build \
  --platform "${DOCKER_PLATFORM}" \
  -f "${REPO_ROOT}/deploy/Dockerfile.prebuilt" \
  -t "sub2api:${TAG}" \
  "${OUT_DIR}"

docker --context "${DOCKER_CONTEXT}" image inspect "sub2api:${TAG}" \
  --format 'Image={{.RepoTags}} OS={{.Os}} Architecture={{.Architecture}}'

docker --context "${DOCKER_CONTEXT}" run --rm "sub2api:${TAG}" /app/sub2api --version

echo "${TAG}" > /tmp/sub2api-last-local-tag
echo "Built sub2api:${TAG}"
