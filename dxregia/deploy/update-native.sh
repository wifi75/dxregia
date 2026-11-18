#!/bin/sh
set -eu

fail() { printf '%s\n' "$*" >&2; exit 1; }
as_root() {
    if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi
}

[ "$(uname -s)" = Linux ] || fail 'Questo script richiede Linux.'
[ "$#" -eq 0 ] || { [ "$#" -eq 2 ] && [ "$1" = --source ]; } ||
    fail "Uso: sh $0 [--source CHECKOUT]"
runtime=/opt/dxregia
state=/var/lib/dxregia
engine=$state/mojo
unit=/etc/systemd/system/dxregia.service
service=dxregia.service

for tool in git go npm curl flock python3; do
    command -v "$tool" >/dev/null 2>&1 || fail "Dipendenza richiesta: $tool"
done
as_root test -f "$unit" && as_root test -d "$runtime" && as_root test -d "$engine" ||
    fail 'Installazione nativa DXRegia con systemd non trovata.'

work=$(mktemp -d)
rollback=
engine_rollback=
staged_engine=
installed=0
success=0
parts='cmd connect data filter msg perl COPYING LICENSE'

cleanup() {
    result=$?
    trap - EXIT
    if [ "$installed" -eq 1 ] && [ "$success" -eq 0 ]; then
        printf '%s\n' 'Aggiornamento interrotto. Ripristino della versione precedente.' >&2
        set +e
        as_root systemctl stop "$service"
        if as_root test -d "$rollback/runtime"; then
            if as_root test -d "$runtime"; then
                as_root mv -T "$runtime" "$rollback/failed-runtime"
            fi
            as_root mv -T "$rollback/runtime" "$runtime"
        fi
        for part in $parts; do
            if as_root test -e "$engine_rollback/$part"; then
                if as_root test -e "$engine/$part"; then
                    as_root mv -T "$engine/$part" "$engine_rollback/failed-$part"
                fi
                as_root mv -T "$engine_rollback/$part" "$engine/$part"
            fi
        done
        as_root cp -a "$rollback/dxregia.service" "$unit"
        if as_root test -f "$rollback/Local.pm"; then
            as_root cp -a "$rollback/Local.pm" "$engine/local/Local.pm"
        fi
        as_root systemctl daemon-reload
        as_root systemctl start "$service"
        printf 'File di ripristino: %s e %s\n' "$rollback" "$engine_rollback" >&2
    fi
    rm -rf -- "$work"
    if [ -n "$staged_engine" ]; then
        as_root rm -rf -- "$staged_engine"
    fi
    exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# The lock is shared by all invocations, including ones started from outside the checkout.
as_root touch /run/lock/dxregia-update.lock
exec 9< /run/lock/dxregia-update.lock
flock -n 9 || fail 'Un aggiornamento DXRegia è già in corso.'

if [ "$#" -eq 2 ]; then
    source=$(CDPATH='' cd -- "$2" && pwd)
else
    git clone --quiet --depth 1 --branch main https://github.com/wifi75/dxregia.git "$work/source"
    source=$work/source
fi
revision=$(git -C "$source" rev-parse --short HEAD)
printf 'Preparazione DXRegia %s. Il nodo resta attivo durante la compilazione.\n' "$revision"
(
    cd "$source/dxregia/web"
    npm ci
    npm run build
)
(
    cd "$source/dxregia/backend"
    go test ./...
    CGO_ENABLED=0 go build -trimpath -o "$work/dxregia" ./cmd/dxregia
)

rollback=$(as_root mktemp -d /opt/.dxregia-update-XXXXXX)
staged_engine=$(as_root mktemp -d "$state/.dxregia-engine-XXXXXX")
engine_rollback=$(as_root mktemp -d "$state/.dxregia-engine-previous-XXXXXX")
as_root install -d -m 0755 "$rollback/new-runtime/bin" "$rollback/new-runtime/web"
as_root install -m 0755 "$work/dxregia" "$rollback/new-runtime/bin/dxregia"
as_root install -m 0755 "$source/dxregia/deploy/backup_node.py" "$rollback/new-runtime/bin/dxregia-backup"
as_root cp -R "$source/dxregia/web/dist/." "$rollback/new-runtime/web/"
for part in $parts; do
    as_root cp -a "$source/$part" "$staged_engine/"
done
as_root chown -hR dxregia:dxregia "$staged_engine"
as_root cp -a "$unit" "$rollback/dxregia.service"
if as_root test -f "$engine/local/Local.pm"; then
    as_root cp -a "$engine/local/Local.pm" "$rollback/Local.pm"
fi

web_port=$(as_root awk -F= '/^DXREGIA_WEB_PORT=[0-9]+$/ {print $2}' /etc/dxregia.env | tail -n 1)
url="http://127.0.0.1:${web_port:-8080}"
configured=0
if as_root test -f "$engine/local/dxregia-initialized"; then configured=1; fi

installed=1
as_root systemctl stop "$service"
as_root mv -T "$runtime" "$rollback/runtime"
as_root mv -T "$rollback/new-runtime" "$runtime"
for part in $parts; do
    as_root mv -T "$engine/$part" "$engine_rollback/$part"
    as_root mv -T "$staged_engine/$part" "$engine/$part"
done
as_root install -m 0644 "$source/dxregia/deploy/dxregia.service" "$unit"
as_root systemctl daemon-reload
as_root systemctl start "$service"
for attempt in $(seq 1 25); do
    if curl -fsS --max-time 2 "$url/api/auth" >/dev/null 2>&1 &&
        { [ "$configured" -eq 0 ] ||
            curl -fsS --max-time 3 "$url/api/listener" | python3 -c 'import json, sys; sys.exit(json.load(sys.stdin).get("state") != "ready")' 2>/dev/null; }; then
        success=1
        printf 'DXRegia aggiornato a %s. Pannello: %s\nVersione precedente: %s e %s\n' "$revision" "$url" "$rollback" "$engine_rollback"
        exit 0
    fi
    sleep 1
done
fail 'Il pannello o il listener Telnet non risponde dopo l’aggiornamento.'
