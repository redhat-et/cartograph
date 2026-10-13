#!/usr/bin/env bash
# image/build.sh — Build and push the multi-arch Crunchy+AGE image.
set -euo pipefail

IMAGE="${IMAGE:-quay.io/redhat-et/crunchy-postgres-age}"
TAG="${TAG:-16-1.6.0}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

echo "Building ${IMAGE}:${TAG} for ${PLATFORMS}"

podman manifest rm "${IMAGE}:${TAG}" 2>/dev/null || true
podman manifest create "${IMAGE}:${TAG}"

for platform in ${PLATFORMS//,/ }; do
    echo "Building for ${platform}..."
    podman build \
        --platform "${platform}" \
        --manifest "${IMAGE}:${TAG}" \
        -f Containerfile \
        .
done

echo "Pushing ${IMAGE}:${TAG}..."
podman manifest push "${IMAGE}:${TAG}" "docker://${IMAGE}:${TAG}"

echo "Done. Image: ${IMAGE}:${TAG}"
