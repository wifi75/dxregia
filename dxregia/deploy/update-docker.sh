#!/bin/sh
set -eu

fail() { printf '%s\n' "$*" >&2; exit 1; }
[ "$(uname -s)" = Linux ] || fail 'Esegui lo script sull’host Linux che ospita Docker.'
[ "$#" -eq 0 ] || { [ "$#" -eq 2 ] && [ "$1" = --source ]; } ||
    fail "Uso: sh $0 [--source CHECKOUT]"
for tool in docker git curl python3; do
    command -v "$tool" >/dev/null 2>&1 || fail "Dipendenza richiesta: $tool"
done
docker info >/dev/null
name=dxregia
work=$(mktemp -d)
revision=
previous=
engine_backup=
image=
restore=0
stopped=0
created=0

engine_change() {
    docker run --rm --network none --mount "type=volume,src=$volume,dst=/data" \
        --entrypoint /bin/sh "$image" -eu -c '
        backup=/data/$1
        parts="cmd connect data filter msg perl COPYING LICENSE"
        case "$2" in
            install)
                mkdir -m 700 "$backup"
                if [ -f /data/mojo/local/Local.pm ]; then
                    cp -p /data/mojo/local/Local.pm "$backup/Local.pm"
                fi
                mkdir "$backup/stage"
                for part in $parts; do
                    test -e "/data/mojo/$part"
                    cp -R "/opt/mojo/$part" "$backup/stage/"
                done
                for part in $parts; do
                    mv "/data/mojo/$part" "$backup/$part"
                    mv "$backup/stage/$part" "/data/mojo/$part"
                done
                rmdir "$backup/stage"
                ;;
            restore)
                for part in $parts; do
                    if [ -e "$backup/$part" ]; then
                        if [ -e "/data/mojo/$part" ]; then
                            mv "/data/mojo/$part" "$backup/failed-$part"
                        fi
                        mv "$backup/$part" "/data/mojo/$part"
                    fi
                done
                if [ -f "$backup/Local.pm" ]; then
                    cp -p "$backup/Local.pm" /data/mojo/local/Local.pm
                fi
                ;;
        esac
    ' sh "$engine_backup" "$1"
}

cleanup() {
    result=$?
    trap - EXIT
    if [ "$restore" -eq 1 ]; then
        printf '%s\n' 'Aggiornamento interrotto. Ripristino della versione precedente.' >&2
        if { [ "$created" -eq 0 ] || docker rm -f "$name" >/dev/null; } &&
            engine_change restore &&
            docker rename "$previous" "$name" &&
            docker start "$name" >/dev/null; then
            printf '%s\n' 'Container precedente riavviato.' >&2
        else
            printf 'Ripristino non completato. Container precedente: %s. Copia Mojo nel volume: %s.\n' "$previous" "$engine_backup" >&2
        fi
    elif [ "$stopped" -eq 1 ]; then
        docker start "$name" >/dev/null || true
    fi
    rm -rf -- "$work"
    exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Inspect once. The temporary directory is private and may contain SMTP credentials.
docker inspect --type container "$name" > "$work/container.json" ||
    fail 'Container Docker dxregia non trovato.'
python3 - "$work" <<'PY'
import json
import os
from pathlib import Path
import re
import sys

work = Path(sys.argv[1])
container, = json.loads((work / "container.json").read_text())
config, host = container["Config"], container["HostConfig"]
if not container["State"]["Running"]:
    sys.exit("Avvia il container DXRegia prima di aggiornarlo.")
mounts = container["Mounts"]
ports = host["PortBindings"]
if (config.get("Labels", {}).get("org.dxregia.installation") != "local"
        or host["NetworkMode"] not in {"default", "bridge"}
        or len(container["NetworkSettings"]["Networks"]) > 1
        or len(mounts) != 1 or mounts[0]["Type"] != "volume"
        or mounts[0]["Destination"] != "/data" or not mounts[0]["RW"]
        or not ports or len(ports) != 2 or "8080/tcp" not in ports):
    sys.exit("Questo script aggiorna il container creato con install-local.sh.")
volume = mounts[0]["Name"]
if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]+", volume):
    sys.exit("Nome del volume non supportato.")
bindings = []
for port, values in sorted(ports.items(), key=lambda entry: entry[0] != "8080/tcp"):
    if (len(values) != 1 or not re.fullmatch(r"[0-9]+/tcp", port)
            or values[0]["HostIp"] not in {"127.0.0.1", "0.0.0.0"}
            or not re.fullmatch(r"[0-9]+", values[0]["HostPort"])):
        sys.exit("Configurazione delle porte non supportata.")
    bindings.append(f'{values[0]["HostIp"]}:{values[0]["HostPort"]}:{port.split("/")[0]}')
policy = host["RestartPolicy"]
restart = policy["Name"] or "no"
if restart == "on-failure" and policy["MaximumRetryCount"]:
    restart += f':{policy["MaximumRetryCount"]}'
environment = [value for value in config["Env"] if value.startswith("DXREGIA_")]
if any("\n" in value or "\r" in value for value in environment):
    sys.exit("Variabile DXREGIA con un valore multilinea non supportato.")
fd = os.open(work / "env", os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
with os.fdopen(fd, "w") as file:
    file.writelines(value + "\n" for value in environment)
(work / "params").write_text("\n".join([volume, *bindings, restart, container["Id"]]) + "\n")
PY
{
    read -r volume
    read -r web_binding
    read -r telnet_binding
    read -r restart
    read -r container_id
} < "$work/params"
web_port=${web_binding#*:}
web_port=${web_port%%:*}
url=http://127.0.0.1:$web_port

if [ "$#" -eq 2 ]; then
    source=$(CDPATH='' cd -- "$2" && pwd)
else
    git clone --quiet --depth 1 --branch main https://github.com/wifi75/dxregia.git "$work/source"
    source=$work/source
fi
revision=$(git -C "$source" rev-parse --short HEAD)
stamp=$(date -u '+%Y%m%dT%H%M%SZ')
image=dxregia:update-$revision
previous=dxregia-previous-$stamp-$$
engine_backup=.dxregia-engine-previous-$stamp-$$

printf 'Compilazione DXRegia %s. Il container resta in esecuzione.\n' "$revision"
docker build -f "$source/dxregia/deploy/Containerfile" -t "$image" "$source"
configured=0
if docker exec "$name" test -f /data/mojo/local/dxregia-initialized; then configured=1; fi

[ "$(docker inspect --type container --format '{{.Id}}' "$name")" = "$container_id" ] ||
    fail 'Il container e cambiato durante la compilazione. Riesegui lo script.'
docker stop "$container_id" >/dev/null
stopped=1
docker rename "$container_id" "$previous"
restore=1
engine_change install
docker create --name "$name" --restart "$restart" --label org.dxregia.installation=local \
    --mount "type=volume,src=$volume,dst=/data" --env-file "$work/env" \
    -p "$web_binding" -p "$telnet_binding" "$image" >/dev/null
created=1
docker start "$name" >/dev/null
for attempt in $(seq 1 25); do
    if curl -fsS --max-time 2 "$url/api/auth" >/dev/null 2>&1 &&
        { [ "$configured" -eq 0 ] ||
            curl -fsS --max-time 3 "$url/api/listener" | python3 -c 'import json, sys; sys.exit(json.load(sys.stdin).get("state") != "ready")' 2>/dev/null; }; then
        restore=0
        stopped=0
        printf 'DXRegia aggiornato a %s. Pannello: %s\nContainer precedente: %s\n' "$revision" "$url" "$previous"
        exit 0
    fi
    sleep 1
done
fail 'Il pannello o il listener Telnet non risponde dopo l’aggiornamento.'
