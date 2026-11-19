#!/bin/sh
set -eu

fail() { printf '%s\n' "$*" >&2; exit 1; }
as_root() {
    if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi
}
case "$(uname -s)" in Linux) ;; *) fail 'Questo script richiede Linux.';; esac
[ -r /etc/os-release ] || fail 'Distribuzione non identificata.'
. /etc/os-release
case "$ID" in debian|ubuntu) ;; *) fail 'Lo script nativo supporta Debian e Ubuntu.';; esac
[ -d /run/systemd/system ] || fail 'Avvia il sistema o la LXC con systemd.'
for tool in curl; do command -v "$tool" >/dev/null 2>&1 || fail "Dipendenza richiesta: $tool"; done
if [ "$(id -u)" -ne 0 ]; then command -v sudo >/dev/null 2>&1 || fail 'Serve sudo o un accesso root.'; fi

root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
state=/var/lib/dxregia
runtime=/opt/dxregia
unit=/etc/systemd/system/dxregia.service
web_port=${DXREGIA_WEB_PORT:-8080}
telnet_port=${DXREGIA_TELNET_PORT:-7300}
validate_port() {
    case "$1" in ''|0*|*[!0-9]*) return 1;; esac
    [ "${#1}" -le 5 ] && [ "$1" -le 65535 ]
}
if ! validate_port "$web_port" || ! validate_port "$telnet_port" ||
   [ "$web_port" = "$telnet_port" ] || [ "$telnet_port" = 27754 ]; then
    fail 'Scegli porte web e Telnet valide e distinte. La porta Telnet 27754 è riservata.'
fi

if [ ! -f "$unit" ]; then
    [ ! -e "$runtime" ] || fail "La destinazione $runtime esiste già."
    stage=$(mktemp -d)
    trap 'rm -rf -- "$stage"' EXIT
    binary="$root/dxregia/prebuilt/dxregia"
    if [ -d "$root/dxregia/prebuilt" ]; then
        [ -f "$binary" ] && [ -f "$root/dxregia/prebuilt/architecture" ] ||
            fail 'Il pacchetto precompilato è incompleto.'
        arch=$(cat "$root/dxregia/prebuilt/architecture")
        [ "$arch" = "$(dpkg --print-architecture)" ] ||
            fail 'Il pacchetto non corrisponde all’architettura del server.'
    else
        command -v go >/dev/null 2>&1 || fail 'Per compilare da sorgente serve Go 1.22 o successivo.'
        command -v npm >/dev/null 2>&1 || fail 'Per compilare da sorgente serve Node.js con npm.'
        binary="$stage/dxregia"
        (
            cd "$root/dxregia/web"
            npm ci
            npm run build
        )
        (
            cd "$root/dxregia/backend"
            CGO_ENABLED=0 go build -trimpath -o "$binary" ./cmd/dxregia
        )
    fi
    [ -f "$root/dxregia/web/dist/index.html" ] || fail 'Interfaccia web mancante dal pacchetto.'
    as_root apt-get update
    as_root apt-get install -y --no-install-recommends \
        ca-certificates perl python3 git procps libev-perl libmojolicious-perl libjson-perl \
        libjson-xs-perl libdata-structure-util-perl libmath-round-perl libnet-cidr-lite-perl \
        libdbd-sqlite3-perl libdbi-perl libtimedate-perl libio-socket-inet6-perl
    if ! id dxregia >/dev/null 2>&1; then
        as_root useradd --system --user-group --no-create-home --shell /usr/sbin/nologin dxregia
    fi
    as_root install -d -m 0700 -o dxregia -g dxregia "$state"
    if [ ! -d "$state/mojo" ]; then
        engine_stage=$(as_root mktemp -d "$state/.mojo-XXXXXX")
        for part in cmd connect data filter msg perl COPYING LICENSE; do
            case $part in
                COPYING|LICENSE) from="$root/$part" ;;
                *) from="$root/engine/$part" ;;
            esac
            as_root cp -a "$from" "$engine_stage/"
        done
        as_root chown -hR dxregia:dxregia "$engine_stage"
        as_root mv -T "$engine_stage" "$state/mojo"
    fi
    as_root install -d -m 0755 "$runtime/bin" "$runtime/web"
    if sh "$root/dxregia/deploy/download-country-db.sh" "$stage/dbip-country.mmdb" 2>/dev/null; then
        as_root install -m 0644 -o dxregia -g dxregia "$stage/dbip-country.mmdb" "$state/mojo/dbip-country.mmdb"
    fi
    as_root install -m 0755 "$binary" "$runtime/bin/dxregia"
    as_root install -m 0755 "$root/dxregia/deploy/backup_node.py" "$runtime/bin/dxregia-backup"
    as_root install -m 0755 "$root/dxregia/deploy/update-native.sh" /usr/local/bin/dxregia-update
    as_root cp -R "$root/dxregia/web/dist/." "$runtime/web/"
    if [ ! -f /etc/dxregia.env ]; then
        printf 'DXREGIA_WEB_PORT=%s\nDXREGIA_TELNET_PORT=%s\nDXREGIA_HTTP_BIND=0.0.0.0\n' "$web_port" "$telnet_port" > "$stage/environment"
        as_root install -m 0600 "$stage/environment" /etc/dxregia.env
    fi
    as_root install -m 0644 "$root/dxregia/deploy/dxregia.service" "$unit"
    as_root systemctl daemon-reload
    as_root systemctl enable --now dxregia.service
else
    as_root systemctl start dxregia.service
fi
configured_port=8080
configured_bind=127.0.0.1
if as_root test -f /etc/dxregia.env; then
    configured_port=$(as_root awk -F= '/^DXREGIA_WEB_PORT=[0-9]+$/ {print $2}' /etc/dxregia.env | tail -n 1)
    configured_bind=$(as_root awk -F= '/^DXREGIA_HTTP_BIND=[0-9.]+$/ {bind=$2} END {print bind}' /etc/dxregia.env)
fi
if as_root test -f "$state/web-port"; then
    configured_port=$(as_root cat "$state/web-port")
fi
probe_host=${configured_bind:-127.0.0.1}
[ "$probe_host" != 0.0.0.0 ] || probe_host=127.0.0.1
url="http://$probe_host:${configured_port:-8080}"
for attempt in $(seq 1 30); do
    if curl -fsS --max-time 2 "$url/api/auth" >/dev/null 2>&1; then
        printf 'Pannello sysop: %s/admin\n' "$url"
        if [ "$configured_bind" = 0.0.0.0 ]; then
            for ip in $(hostname -I 2>/dev/null || true); do
                case "$ip" in *:*) continue ;; esac
                printf 'Pannello dalla rete: http://%s:%s/admin\n' "$ip" "${configured_port:-8080}"
            done
        fi
        if as_root test -f "$state/sysop-initial"; then
            printf 'Utente iniziale: sysop\n'
            printf 'Password temporanea sysop: '
            as_root cat "$state/sysop-initial"
            printf '\n'
            printf 'Al primo accesso cambia la password. Puoi attivare il 2FA in Gestione > Password.\n'
        fi
        exit 0
    fi
    sleep 1
done
fail 'Il pannello non risponde. Consulta journalctl -u dxregia.service.'
