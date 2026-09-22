#!/usr/bin/env bash
set -euo pipefail

version=${1:?usage: build-release.sh VERSION [DIST]}
dist=${2:-dist}

rm -rf "$dist"
mkdir -p "$dist"

for goarch in amd64 arm64; do
  stage=$(mktemp -d)
  trap 'rm -rf "$stage"' EXIT

  CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$version" \
    -o "$stage/ocrecent" \
    ./cmd/ocrecent
  cp LICENSE README.md "$stage/"
  tar -C "$stage" -czf "$dist/ocrecent_${version}_linux_${goarch}.tar.gz" \
    ocrecent LICENSE README.md

  rm -rf "$stage"
  trap - EXIT
done

(
  cd "$dist"
  sha256sum ocrecent_*.tar.gz > checksums.txt
)
