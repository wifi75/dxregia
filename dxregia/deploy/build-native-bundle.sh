#!/bin/sh
set -eu
fail() { printf '%s\n' "$*" >&2; exit 1; }
[ "$(uname -s)" = Linux ] || fail 'Il pacchetto si compila da Linux.'
[ "$#" -eq 1 ] && [ -d "$1" ] || fail "Uso: sh $0 DIRECTORY_OUTPUT"
output=$(CDPATH='' cd -- "$1" && pwd)
root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
arch=${GOARCH:-$(go env GOHOSTARCH)}
case "$arch" in amd64|arm64) ;; *) fail 'Architetture supportate: amd64 e arm64.';; esac
archive="$output/dxregia-linux-$arch.tar.gz"
checksum="$archive.sha256"
[ ! -e "$archive" ] && [ ! -e "$checksum" ] || fail 'Un pacchetto con lo stesso nome esiste già.'
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
bundle="$work/dxregia-linux-$arch"
mkdir -p "$bundle/dxregia/prebuilt" "$bundle/dxregia/web/dist" "$bundle/dxregia/deploy"
(
    cd "$root/dxregia/web"
    npm ci
    npm run build
    cp -a dist/. "$bundle/dxregia/web/dist/"
)
(
    cd "$root/dxregia/backend"
    go test ./...
    GOOS=linux GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -o "$bundle/dxregia/prebuilt/dxregia" ./cmd/dxregia
)
printf '%s\n' "$arch" > "$bundle/dxregia/prebuilt/architecture"
for name in perl cmd connect data filter msg COPYING LICENSE; do
    cp -a "$root/$name" "$bundle/"
done
for name in install-native.sh update-native.sh dxregia.service backup_node.py migrate_mojo.py check_telnet.py README.md; do
    cp -a "$root/dxregia/deploy/$name" "$bundle/dxregia/deploy/"
done
stage=$(mktemp -d "$output/.dxregia-package-XXXXXX")
trap 'rm -rf -- "$work" "$stage"' EXIT
tar -czf "$stage/package.tar.gz" -C "$work" "$(basename "$bundle")"
digest=$(sha256sum "$stage/package.tar.gz")
printf '%s  %s\n' "${digest%% *}" "$(basename "$archive")" > "$stage/checksum"
ln "$stage/package.tar.gz" "$archive"
ln "$stage/checksum" "$checksum"
printf 'Pacchetto: %s\nChecksum: %s\n' "$archive" "$checksum"
