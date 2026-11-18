#!/bin/sh
set -eu

fail() { printf '%s\n' "$*" >&2; exit 1; }
[ "$#" -le 1 ] || fail "Uso: sh $0 [vMAJOR.MINOR.PATCH]"
[ "$(uname -s)" = Linux ] || fail 'Questo script richiede Linux.'
case "$(uname -m)" in
    x86_64) arch=amd64 ;;
    aarch64) arch=arm64 ;;
    *) fail 'Architetture supportate: amd64 e arm64.' ;;
esac
for tool in curl tar sha256sum; do
    command -v "$tool" >/dev/null 2>&1 || fail "Dipendenza richiesta: $tool"
done

releases=https://github.com/wifi75/dxregia/releases
if [ "$#" -eq 0 ]; then
    url=$releases/latest/download
else
    printf '%s\n' "$1" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' ||
        fail 'Specifica una versione nel formato vMAJOR.MINOR.PATCH.'
    url=$releases/download/$1
fi

work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
cd "$work"
bundle=dxregia-linux-$arch
archive=$bundle.tar.gz
download() {
    curl -fsSL --proto '=https' --connect-timeout 15 --retry 2 --output "$2" "$1"
}
download "$url/$archive" "$archive" ||
    fail "Pacchetto non disponibile: $url/$archive"
download "$url/$archive.sha256" "$archive.sha256" ||
    fail "Checksum non disponibile: $url/$archive.sha256"
read -r digest filename extra < "$archive.sha256" || fail 'Checksum non valido.'
[ "$filename" = "$archive" ] && [ -z "$extra" ] &&
    printf '%s\n' "$digest" | grep -Eq '^[0-9a-f]{64}$' || fail 'Checksum non valido.'
printf '%s  %s\n' "$digest" "$archive" | sha256sum -c - ||
    fail 'Verifica del pacchetto fallita. Scaricalo di nuovo.'
tar -xzf "$archive"
sh "$bundle/dxregia/deploy/install-native.sh"
