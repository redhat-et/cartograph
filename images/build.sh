#!/usr/bin/env bash
# images/build.sh — Build and optionally push multi-arch container images.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

usage() {
    cat <<EOF
Usage: $0 [--push] <target>

Targets:
  cartograph             Server image (Go binary on UBI9-minimal)
  crunchy-postgres-age   Database image (PostgreSQL 16 + AGE 1.6.0)
  all                    Both images

Environment:
  PLATFORMS   Architectures (default: linux/amd64,linux/arm64)

Override image name/tag per target:
  CARTOGRAPH_IMAGE   (default: quay.io/redhat-et/cartograph)
  CARTOGRAPH_TAG     (default: latest)
  CRUNCHY_IMAGE      (default: quay.io/redhat-et/crunchy-postgres-age)
  CRUNCHY_TAG        (default: 16-1.6.0)
EOF
    exit 1
}

build_image() {
    local image="$1" tag="$2" containerfile="$3" context="$4"
    local ref="${image}:${tag}"

    echo "==> Building ${ref} for ${PLATFORMS}"

    podman manifest rm "${ref}" 2>/dev/null || true
    podman manifest create "${ref}"

    for platform in ${PLATFORMS//,/ }; do
        echo "  ${platform}..."
        podman build \
            --platform "${platform}" \
            --manifest "${ref}" \
            -f "${containerfile}" \
            "${context}"
    done

    if [[ "${PUSH}" == true ]]; then
        echo "  Pushing ${ref}..."
        podman manifest push "${ref}" "docker://${ref}"
    fi

    echo "  Done: ${ref}"
}

build_cartograph() {
    local image="${CARTOGRAPH_IMAGE:-quay.io/redhat-et/cartograph}"
    local tag="${CARTOGRAPH_TAG:-latest}"
    build_image "${image}" "${tag}" \
        "${SCRIPT_DIR}/cartograph/Containerfile" "${REPO_ROOT}"
}

build_crunchy() {
    local image="${CRUNCHY_IMAGE:-quay.io/redhat-et/crunchy-postgres-age}"
    local tag="${CRUNCHY_TAG:-16-1.6.0}"
    build_image "${image}" "${tag}" \
        "${SCRIPT_DIR}/crunchy-postgres-age/Containerfile" \
        "${SCRIPT_DIR}/crunchy-postgres-age"
}

PUSH=false
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --push) PUSH=true; shift ;;
        --help|-h) usage ;;
        *)
            [[ -n "${TARGET}" ]] && usage
            TARGET="$1"; shift
            ;;
    esac
done

[[ -z "${TARGET}" ]] && usage

case "${TARGET}" in
    cartograph)            build_cartograph ;;
    crunchy-postgres-age)  build_crunchy ;;
    all)                   build_cartograph; build_crunchy ;;
    *)                     echo "Unknown target: ${TARGET}"; usage ;;
esac
