#!/bin/sh
set -eu

if [ "$(uname -s)" != Linux ]; then
    printf '%s\n' 'Questo script richiede Linux.' >&2
    exit 1
fi
for tool in docker curl; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        printf 'Dipendenza richiesta: %s\n' "$tool" >&2
        exit 1
    fi
done
docker info >/dev/null
root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
name=dxregia
image=dxregia:local
if docker container inspect "$name" >/dev/null 2>&1; then
    label=$(docker container inspect --format '{{ index .Config.Labels "org.dxregia.installation" }}' "$name")
    if [ "$label" != local ]; then
        printf 'Il nome del container %s è già in uso.\n' "$name" >&2
        exit 1
    fi
    docker start "$name" >/dev/null
else
    docker build -f "$root/dxregia/deploy/Containerfile" -t "$image" "$root"
    docker volume create dxregia-data >/dev/null
    docker run -d --name "$name" --restart unless-stopped \
        --label org.dxregia.installation=local \
        --mount type=volume,src=dxregia-data,dst=/data \
        -p 127.0.0.1:8080:8080 -p 127.0.0.1:7300:7300 "$image" >/dev/null
fi

url=http://127.0.0.1:8080
for attempt in $(seq 1 30); do
    if curl -fsS --max-time 2 "$url/api/auth" >/dev/null 2>&1; then
        printf 'Pannello sysop: %s\n' "$url"
        if docker exec "$name" test -f /data/mojo/.dxregia-sysop; then
            printf '%s\n' 'Usa la password sysop già configurata.'
        else
            printf '%s' 'Password temporanea sysop: '
            docker exec "$name" cat /data/sysop-initial
            printf '\n'
        fi
        exit 0
    fi
    sleep 1
done
printf 'Il pannello non risponde. Consulta docker logs %s.\n' "$name" >&2
exit 1
