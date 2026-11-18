#!/bin/sh
set -eu

root=/data/mojo
if [ ! -d "$root" ]; then
    stage=$(mktemp -d /data/.mojo-XXXXXX)
    cp -R /opt/mojo/. "$stage/"
    mv -T "$stage" "$root"
fi

exec /usr/local/bin/dxregia \
    -port 8080 -http-bind 0.0.0.0 -web /opt/dxregia/web \
    -telnet "127.0.0.1:${DXREGIA_TELNET_PORT:-7300}" \
    -bootstrap-password-file /data/sysop-initial -setup-root "$root" -setup-perl /usr/bin/perl -setup-telnet-bind 0.0.0.0
