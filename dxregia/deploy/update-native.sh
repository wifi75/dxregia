#!/bin/sh
set -eu

indicator() {
    if [ -t 1 ] && [ "${TERM:-dumb}" != dumb ] && [ -z "${NO_COLOR:-}" ]; then
        printf '\033[%sm++++++++++\033[0m  %s\n' "$1" "$2"
    else
        printf '++++++++++  %s\n' "$2"
    fi
}
fail() { indicator 31 "Errore: $*" >&2; exit 1; }
steps=5
step() { printf '\n[%s/%s] %s\n' "$1" "$steps" "$2"; }
as_root() {
    if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi
}

detect_url() {
    web_port=8080
    web_host=127.0.0.1
    if as_root test -f /etc/dxregia.env; then
        web_port=$(as_root awk -F= '/^DXREGIA_WEB_PORT=[0-9]+$/ { port=$2 } END { print port }' /etc/dxregia.env)
        web_host=$(as_root awk -F= '/^DXREGIA_HTTP_BIND=[0-9.]+$/ { host=$2 } END { print host }' /etc/dxregia.env)
        [ "${web_host:-127.0.0.1}" != 0.0.0.0 ] || web_host=127.0.0.1
    fi
    if as_root test -f "$state/web-port"; then
        web_port=$(as_root cat "$state/web-port")
    fi
    url="http://${web_host:-127.0.0.1}:${web_port:-8080}"
}

[ "$(uname -s)" = Linux ] || fail 'Questo script richiede Linux.'
mode=update
case "$#:${1:-}" in
    0:) ;;
    1:--check) mode=check ;;
    2:--source) mode=source ;;
    *) fail "Uso: sh $0 [--check | --source CHECKOUT]" ;;
esac
runtime=/opt/dxregia
state=/var/lib/dxregia
engine=$state/mojo
unit=/etc/systemd/system/dxregia.service
service=dxregia.service

command -v git >/dev/null 2>&1 || fail "Dipendenza richiesta: git"
as_root test -f "$unit" && as_root test -d "$runtime" && as_root test -d "$engine" ||
    fail 'Installazione nativa DXRegia con systemd non trovata.'

remote=https://github.com/wifi75/dxregia.git
if [ "$mode" = source ]; then
    source=$(CDPATH='' cd -- "$2" && pwd)
    available=$(git -C "$source" rev-parse HEAD)
else
    printf '%s\n' 'Controllo aggiornamenti DXRegia su GitHub...'
    remote_head=$(git ls-remote --exit-code "$remote" refs/heads/${DXREGIA_BRANCH:-main}) ||
        fail 'Impossibile controllare gli aggiornamenti su GitHub.'
    available=$(printf '%s\n' "$remote_head" | awk '{print $1}')
fi
current=
if as_root test -f "$runtime/revision"; then
    current=$(as_root cat "$runtime/revision")
elif command -v go >/dev/null 2>&1; then
    current=$(go version -m "$runtime/bin/dxregia" 2>/dev/null |
        awk '$2 ~ /^vcs.revision=/ {sub(/^vcs.revision=/, "", $2); print $2; exit}')
fi
printf 'Installata: %.12s\nDisponibile: %.12s\n' "${current:-sconosciuta}" "$available"
if [ "$current" = "$available" ] && [ "$mode" != source ]; then
    indicator 32 'DXRegia è già aggiornato.'
    exit 0
fi
if [ -n "$current" ] && [ "$current" != "$available" ]; then
    indicator 32 'Aggiornamento disponibile.'
elif [ -z "$current" ]; then
    printf '%s\n' 'Questa installazione non registra ancora il commit. Dopo l’aggiornamento sarà possibile confrontarlo.'
fi
[ "$mode" != check ] || exit 0
for tool in go npm curl flock python3; do
    command -v "$tool" >/dev/null 2>&1 || fail "Dipendenza richiesta: $tool"
done

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
        indicator 31 'Aggiornamento interrotto. Ripristino della versione precedente.' >&2
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
        for config in Local.pm Listeners.pm; do
            if as_root test -f "$rollback/$config"; then
                as_root cp -a "$rollback/$config" "$engine/local/$config"
            fi
        done
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

step 1 'Scarico il sorgente. Il nodo rimane online.'
if [ "$mode" != source ]; then
    git clone --quiet --depth 1 --branch "${DXREGIA_BRANCH:-main}" "$remote" "$work/source" ||
        fail 'Download del sorgente non riuscito.'
    source=$work/source
fi
revision=$(git -C "$source" rev-parse HEAD)
if [ "$revision" = "$current" ] && [ "$mode" != source ]; then
    indicator 32 'DXRegia è già aggiornato.'
    exit 0
fi
indicator 32 'Sorgente pronto.'
# Se sono cambiati solo la grafica del pannello e la documentazione, i file si sostituiscono senza riavviare
# il servizio: Telnet e i collegamenti con i nodi partner restano attivi. DXREGIA_FULL=1 forza l'aggiornamento completo.
light=0
if [ "$mode" != source ] && [ "${DXREGIA_FULL:-0}" != 1 ] && [ -n "$current" ] &&
    git -C "$source" fetch --quiet --depth 1 origin "$current" 2>/dev/null; then
    changed=$(git -C "$source" diff --name-only "$current" "$revision" 2>/dev/null || true)
    if [ -n "$changed" ] && ! printf '%s\n' "$changed" | grep -qvE '^(dxregia/web/|docs/|[^/]+\.md$|dxregia/deploy/README\.md$)'; then
        light=1
        steps=3
        printf '%s\n' 'Sono cambiati solo la grafica e la documentazione: aggiornamento senza riavvio.'
    fi
fi
# Se non sono cambiati né il motore né il suo hook (engine/, COPYING, LICENSE, presence.pm), il servizio web si riavvia
# lasciando attivo il motore: Telnet e i collegamenti con i nodi partner non si interrompono.
keep_engine=0
if [ "$light" -eq 0 ] && [ "$mode" != source ] && [ "${DXREGIA_FULL:-0}" != 1 ] && [ -n "$current" ] &&
    git -C "$source" cat-file -e "$current" 2>/dev/null; then
    if engine_changed=$(git -C "$source" diff --name-only "$current" "$revision" -- engine COPYING LICENSE \
        dxregia/backend/cmd/dxregia/presence.pm 2>/dev/null) && [ -z "$engine_changed" ]; then
        keep_engine=1
        parts=''
        printf '%s\n' 'Il motore non è cambiato: resta attivo durante il riavvio del pannello.'
    fi
fi
step 2 "Compilo il pannello web ($(printf '%.12s' "$revision"))."
(
    cd "$source/dxregia/web" && npm ci && npm run build
) > "$work/web-build.log" 2>&1 || {
    cat "$work/web-build.log" >&2
    fail 'Compilazione del pannello web non riuscita.'
}
if [ "${DXREGIA_VERBOSE:-0}" = 1 ]; then
    cat "$work/web-build.log"
fi
indicator 32 'Pannello web compilato.'
if [ "$light" -eq 1 ]; then
    step 3 'Sostituisco i file del pannello. Nessun riavvio: Telnet e collegamenti restano attivi.'
    detect_url
    printf '%s\n' "$revision" > "$work/revision"
    rollback=$(as_root mktemp -d /opt/.dxregia-update-XXXXXX)
    as_root install -d -m 0755 "$runtime/web.new"
    as_root cp -R "$source/dxregia/web/dist/." "$runtime/web.new/"
    as_root cp -a "$runtime/revision" "$rollback/revision"
    as_root mv -T "$runtime/web" "$rollback/web"
    as_root mv -T "$runtime/web.new" "$runtime/web"
    as_root install -m 0644 "$work/revision" "$runtime/revision"
    for attempt in $(seq 1 10); do
        if curl -fsS --max-time 3 "$url/api/auth" >/dev/null 2>&1 && curl -fsS --max-time 3 "$url/" >/dev/null 2>&1; then
            for old in $(ls -1dt /opt/.dxregia-update-* 2>/dev/null | tail -n +4); do as_root rm -rf -- "$old"; done
            indicator 32 "DXRegia aggiornato a $(printf '%.12s' "$revision") (senza riavvio)."
            exit 0
        fi
        sleep 1
    done
    as_root mv -T "$runtime/web" "$rollback/failed-web"
    as_root mv -T "$rollback/web" "$runtime/web"
    as_root cp -a "$rollback/revision" "$runtime/revision"
    indicator 31 'Ripristino della versione precedente.' >&2
    fail 'Il pannello non risponde dopo la sostituzione dei file.'
fi
step 3 'Compilo il backend Go. Questa fase può restare silenziosa per qualche minuto.'
(
    cd "$source/dxregia/backend" &&
    CGO_ENABLED=0 go build -trimpath -o "$work/dxregia" ./cmd/dxregia
) || fail 'Compilazione del backend Go non riuscita.'
indicator 32 'Backend compilato.'
if sh "$source/dxregia/deploy/download-country-db.sh" "$work/dbip-country.mmdb" 2>/dev/null; then
    as_root install -m 0644 -o dxregia -g dxregia "$work/dbip-country.mmdb" "$engine/dbip-country.mmdb.new"
    as_root mv -fT "$engine/dbip-country.mmdb.new" "$engine/dbip-country.mmdb"
fi
printf '%s\n' "$revision" > "$work/revision"

rollback=$(as_root mktemp -d /opt/.dxregia-update-XXXXXX)
staged_engine=$(as_root mktemp -d "$state/.dxregia-engine-XXXXXX")
engine_rollback=$(as_root mktemp -d "$state/.dxregia-engine-previous-XXXXXX")
as_root install -d -m 0755 "$rollback/new-runtime/bin" "$rollback/new-runtime/web"
as_root install -m 0644 "$work/revision" "$rollback/new-runtime/revision"
as_root install -m 0755 "$work/dxregia" "$rollback/new-runtime/bin/dxregia"
as_root install -m 0755 "$source/dxregia/deploy/backup_node.py" "$rollback/new-runtime/bin/dxregia-backup"
as_root cp -R "$source/dxregia/web/dist/." "$rollback/new-runtime/web/"
for part in $parts; do
    case $part in
        COPYING|LICENSE) from="$source/$part" ;;
        *) from="$source/engine/$part" ;;
    esac
    as_root cp -a "$from" "$staged_engine/"
done
as_root chown -hR dxregia:dxregia "$staged_engine"
as_root cp -a "$unit" "$rollback/dxregia.service"
for config in Local.pm Listeners.pm; do
    if as_root test -f "$engine/local/$config"; then
        as_root cp -a "$engine/local/$config" "$rollback/$config"
    fi
done

detect_url
configured=0
if as_root test -f "$engine/local/dxregia-initialized"; then configured=1; fi

step 4 'Applico i file e riavvio il servizio.'
installed=1
if [ "$keep_engine" -eq 1 ]; then as_root touch "$state/keep-engine"; fi
as_root systemctl stop "$service"
as_root mv -T "$runtime" "$rollback/runtime"
as_root mv -T "$rollback/new-runtime" "$runtime"
for part in $parts; do
    as_root mv -T "$engine/$part" "$engine_rollback/$part"
    as_root mv -T "$staged_engine/$part" "$engine/$part"
done
as_root install -m 0644 "$source/dxregia/deploy/dxregia.service" "$unit"
# L'aggiornamento dal pannello usa questo stesso script e due unità systemd.
as_root install -m 0755 "$source/dxregia/deploy/update-native.sh" /usr/local/bin/dxregia-update
for update_unit in dxregia-update.service dxregia-update.path; do
    as_root install -m 0644 "$source/dxregia/deploy/$update_unit" "/etc/systemd/system/$update_unit"
done
as_root systemctl daemon-reload
as_root systemctl enable --now dxregia-update.path || true
as_root systemctl start "$service"
as_root rm -f "$state/keep-engine"
step 5 'Verifico la risposta del pannello e, se configurato, di Telnet.'
for attempt in $(seq 1 25); do
    if curl -fsS --max-time 2 "$url/api/auth" >/dev/null 2>&1 &&
        { [ "$configured" -eq 0 ] ||
            curl -fsS --max-time 3 "$url/api/listener" | python3 -c 'import json, sys; sys.exit(json.load(sys.stdin).get("state") != "ready")' 2>/dev/null; }; then
        success=1
        # Le versioni precedenti servono solo per tornare indietro: si tengono le ultime tre (10–15 MB ciascuna).
        as_root rmdir "$staged_engine" 2>/dev/null || true
        for old in $(ls -1dt /opt/.dxregia-update-* 2>/dev/null | tail -n +4); do as_root rm -rf -- "$old"; done
        for old in $(ls -1dt "$state"/.dxregia-engine-previous-* 2>/dev/null | tail -n +4); do as_root rm -rf -- "$old"; done
        indicator 32 "DXRegia aggiornato a $(printf '%.12s' "$revision")."
        printf 'Pannello (verifica locale): %s/admin\nVersione precedente: %s e %s\n' "$url" "$rollback" "$engine_rollback"
        exit 0
    fi
    sleep 1
done
fail 'Il pannello o il listener Telnet non risponde dopo l’aggiornamento.'
