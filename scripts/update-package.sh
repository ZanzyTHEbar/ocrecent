#!/usr/bin/env bash
set -euo pipefail

version=${1:?usage: update-package.sh VERSION [SOURCE_COMMIT]}
source_commit=${2:-$(git rev-parse HEAD)}
archive_url="https://github.com/ZanzyTHEbar/ocrecent/archive/${source_commit}.tar.gz"

checksum=$(curl --fail --location --silent --show-error "$archive_url" | sha256sum)
checksum=${checksum%% *}

tmp=$(mktemp packaging/PKGBUILD.XXXXXX)
trap 'rm -f "$tmp"' EXIT
sed \
  -e "s/^pkgver=.*/pkgver=$version/" \
  -e "s/^source_commit=.*/source_commit=$source_commit/" \
  -e "s/^sha256sums=.*/sha256sums=('$checksum')/" \
  packaging/PKGBUILD > "$tmp"
chmod 0644 "$tmp"
mv "$tmp" packaging/PKGBUILD
trap - EXIT
