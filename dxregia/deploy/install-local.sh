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
web_port=${DXREGIA_WEB_PORT:-8080}
telnet_port=${DXREGIA_TELNET_PORT:-7300}

validate_port() {
    case "$1" in
        ''|0*|*[!0-9]*) return 1 ;;
    esac
    [ "${#1}" -le 5 ] && [ "$1" -le 65535 ]
}
if ! validate_port "$web_port" || ! validate_port "$telnet_port" || [ "$web_port" = "$telnet_port" ] || [ "$telnet_port" = 8080 ] || [ "$telnet_port" = 27754 ]; then
    printf '%s\n' 'Scegli due porte valide e distinte. La porta Telnet non può essere 8080 o 27754.' >&2
    exit 1
fi
if docker container inspect "$name" >/dev/null 2>&1; then
    label=$(docker container inspect --format '{{ index .Config.Labels "org.dxregia.installation" }}' "$name")
    if [ "$label" != local ]; then
        printf 'Il nome del container %s è già in uso.\n' "$name" >&2
        exit 1
    fi
    current_web=$(docker inspect --format '{{(index (index .HostConfig.PortBindings "8080/tcp") 0).HostPort}}' "$name")
    current_telnet=$(docker inspect --format '{{range $port, $mapping := .HostConfig.PortBindings}}{{if ne $port "8080/tcp"}}{{(index $mapping 0).HostPort}}{{end}}{{end}}' "$name")
    if [ "$current_web" != "$web_port" ] || [ "$current_telnet" != "$telnet_port" ]; then
        printf 'Il container esistente usa web %s e Telnet %s. Riavvialo con gli stessi parametri.\n' "$current_web" "$current_telnet" >&2
        exit 1
    fi
    docker start "$name" >/dev/null
else
    docker build -f "$root/dxregia/deploy/Containerfile" -t "$image" "$root"
    docker volume create dxregia-data >/dev/null
    docker run -d --name "$name" --restart unless-stopped \
        --label org.dxregia.installation=local \
        --mount type=volume,src=dxregia-data,dst=/data \
        -e DXREGIA_TELNET_PORT="$telnet_port" \
        -p "127.0.0.1:$web_port:8080" -p "127.0.0.1:$telnet_port:$telnet_port" "$image" >/dev/null
fi

url="http://127.0.0.1:$web_port"
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
