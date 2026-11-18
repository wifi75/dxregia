#!/bin/sh
set -eu

target=$1
current=$(date -u +%Y-%m)
previous=$(date -u -d "$(date -u +%Y-%m-01) -1 day" +%Y-%m)
for month in "$current" "$previous"; do
    if curl -fsSL --max-time 25 -o "$target.gz" "https://download.db-ip.com/free/dbip-country-lite-$month.mmdb.gz" &&
        gzip -dc "$target.gz" > "$target" && [ -s "$target" ]; then
        exit 0
    fi
done
exit 1
