"""Smoke-test a fresh DXRegia Docker installation on a disposable Linux runner."""

import base64
import hmac
import http.cookiejar
import json
import os
import secrets
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
import zipfile
from pathlib import Path

HOST = "127.0.0.1"
WEB_PORT = int(os.environ.get("DXREGIA_WEB_PORT", "8080"))
TELNET_PORT = int(os.environ.get("DXREGIA_TELNET_PORT", "7300"))
BASE = f"http://{HOST}:{WEB_PORT}"
CONTAINER = "dxregia"
NATIVE = os.environ.get("DXREGIA_SMOKE_NATIVE") == "1"


def docker(*args):
    return subprocess.run(["docker", *args], check=True, stdout=subprocess.PIPE, text=True).stdout


def privileged(*args):
    return subprocess.run(["sudo", *args], check=True, stdout=subprocess.PIPE, text=True).stdout


def service(action):
    if NATIVE:
        privileged("systemctl", action, "dxregia.service")
    else:
        docker(action, CONTAINER)


def copy_node(destination):
    if NATIVE:
        privileged("cp", "-a", "/var/lib/dxregia/mojo", str(destination))
        privileged("chown", "-hR", f"{os.getuid()}:{os.getgid()}", str(destination))
    else:
        docker("cp", CONTAINER + ":/data/mojo", str(destination))


def backup_command(work, *args):
    if NATIVE:
        subprocess.run(["/opt/dxregia/bin/dxregia-backup", *map(str, args)], check=True)
    else:
        paths = ["/smoke/" + str(path.relative_to(work)) for path in args[1:]]
        docker("run", "--rm", "--network", "none", "--user", f"{os.getuid()}:{os.getgid()}",
               "--mount", f"type=bind,src={work},dst=/smoke", "--entrypoint", "python3",
               "dxregia:local", "/opt/dxregia/backup_node.py", args[0], *paths)


def reject_corrupt_backup(work, original):
    damaged = work / "damaged.zip"
    destination = work / "damaged-restore"
    with zipfile.ZipFile(original) as source, zipfile.ZipFile(damaged, "w") as output:
        for name in source.namelist():
            content = source.read(name)
            if name == "local/DXVars.pm":
                assert content
                content = bytes([content[0] ^ 1]) + content[1:]
            output.writestr(name, content)
    try:
        backup_command(work, "restore", damaged, destination)
    except subprocess.CalledProcessError as error:
        assert error.returncode == 1
    else:
        raise AssertionError("Restore accepted a modified file with a stale manifest")
    assert not destination.exists(), "A failed restore published a partial node"


def request(client, path, body=None, method=None, base=BASE):
    req = urllib.request.Request(
        base + path,
        data=None if body is None else json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
        method=method,
    )
    try:
        with client.open(req, timeout=30) as response:
            data = response.read()
            return response.status, json.loads(data) if data else None
    except urllib.error.HTTPError as error:
        return error.code, None



def portal_traffic(client):
    public = urllib.request.build_opener()
    assert request(public, "/api/traffic")[0] == 401
    assert request(public, "/api/visits") == (200, {"Views": 0})
    for page in ("/spots", "/contatti"):
        visit = urllib.request.Request(BASE + page, headers={"Referer": "https://radio.example.org/path?private=1"})
        with public.open(visit, timeout=10) as response:
            assert response.status == 200
    token = secrets.token_hex(16)
    for _ in range(2):
        assert request(public, "/api/visits/heartbeat", token) == (204, None)
    assert request(public, "/api/visits") == (200, {"Views": 2})
    status, traffic = request(client, "/api/traffic")
    assert status == 200 and traffic["Active"] == 1
    for field, expected in [
        ("Pages", {"/spots": 1, "/contatti": 1}),
        ("Referrers", {"radio.example.org": 2}),
    ]:
        totals = {}
        for day in traffic["Days"].values():
            for label, count in day[field].items():
                totals[label] = totals.get(label, 0) + count
        assert totals == expected, (field, totals)


def totp(secret):
    key = base64.b32decode(secret)
    counter = (int(time.time()) // 30).to_bytes(8, "big")
    digest = hmac.digest(key, counter, "sha1")
    offset = digest[-1] & 15
    number = int.from_bytes(digest[offset:offset+4], "big") & 0x7fffffff
    return f"{number % 1000000:06d}"


def read_until(conn, marker):
    data = bytearray()
    while marker not in data:
        chunk = conn.recv(4096)
        if not chunk or len(data) >= 65536:
            raise RuntimeError("Expected Telnet prompt was absent")
        data.extend(chunk)
    return data


def telnet_login(callsign, location=None):
    with socket.create_connection((HOST, TELNET_PORT), timeout=5) as conn:
        read_until(conn, b"login:")
        conn.sendall(callsign + b"\r\n")
        welcome = read_until(conn, b"dxspider >")
        assert callsign in welcome
        if location is not None:
            assert location in welcome
        conn.sendall(b"bye\r\n")


def ready(client):
    end = time.monotonic() + 30
    while time.monotonic() < end:
        try:
            status, data = request(client, "/api/listener")
            if status == 200 and data["state"] == "ready":
                return
        except OSError:
            pass
        time.sleep(1)
    raise RuntimeError("Mojo did not become ready after restart")


def wait_for(description, check, timeout=40):
    until = time.monotonic() + timeout
    while time.monotonic() < until:
        try:
            result = check()
            if result:
                return result
        except OSError:
            pass
        time.sleep(1)
    raise RuntimeError(f"Timed out waiting for {description}")


def telnet_presence(client):
    callsign = "Q7QQQ"

    def visible():
        status, snapshot = request(client, "/api/presence")
        assert status == 200
        return any(
            channel["Call"] == callsign and channel["Kind"] == "telnet"
            for channel in (snapshot["Channels"] or [])
        )

    with socket.create_connection((HOST, TELNET_PORT), timeout=5) as conn:
        read_until(conn, b"login:")
        conn.sendall(callsign.encode() + b"\r\n")
        read_until(conn, b"dxspider >")
        wait_for("Telnet user on the sysop dashboard", visible)
        conn.sendall(b"bye\r\n")
    wait_for("Telnet user to leave the sysop dashboard", lambda: not visible())


def spot_delivery(primary, remote, peer, peer_port):
    # An isolated documentation subnet lets the real PC61 source pass the
    # existing RFC1918 check without changing any production filters.
    network = "dxregia-spot-smoke"
    sender_ip, receiver_ip = "192.0.2.10", "192.0.2.11"
    docker("network", "create", "--internal", "--subnet", "192.0.2.0/24", network)
    try:
        docker("network", "connect", "--ip", sender_ip, network, CONTAINER)
        docker("network", "connect", "--ip", receiver_ip, network, peer)
        marker = "DXREGIA-SMOKE"
        with (socket.create_connection((receiver_ip, peer_port), timeout=15) as logger,
              socket.create_connection((sender_ip, TELNET_PORT), timeout=15) as spotter):
            for connection, call in [(logger, b"Q9QQQ"), (spotter, b"Q7QQQ")]:
                read_until(connection, b"login:")
                connection.sendall(call + b"\r\n")
                read_until(connection, b"dxspider >")
            spotter.sendall(f"dx 14025 JA1ABC {marker}\r\n".encode())
            reply = read_until(spotter, b"dxspider >")
            # Mojo can hold a new user's spots until the next PC92 update
            # (60 seconds by default, plus a five-second grace period).
            logger.settimeout(80)
            try:
                delivered = read_until(logger, marker.encode())
            except (TimeoutError, RuntimeError):
                print(f"Spot command reply: {reply!r}", flush=True)
                for name in (CONTAINER, peer):
                    subprocess.run(["docker", "exec", name, "sh", "-c",
                                    "stat -c 'presence modified at %y' /data/mojo/local_data/dxregia-presence.json; "
                                "cat /data/mojo/local_data/dxregia-presence.json; echo; "
                                "ps -o pid,stat,pcpu,comm -C perl; "
                                "grep -aEi 'undefined|died at|redefined|failed at|recursion|DBD::|fatal' "
                                "/data/mojo/local_data/dxregia-startup.log | tail -20; "
                                "find /data/mojo/local_data/debug -type f -name '*.dat' "
                                    "-exec grep -aE 'DXREGIA-SMOKE|JA1ABC|PC61 .* dropped' {} +"],
                                   check=False)
                raise
            assert any(
                line.startswith(b"DX de Q7QQQ") and b"JA1ABC" in line and marker.encode() in line
                for line in delivered.splitlines()
            )

        def contains_spot(response):
            status, feed = response
            assert status == 200 and feed["Available"]
            return any(spot["Call"] == "JA1ABC" and spot["Comment"].strip() == marker and spot["Entity"] == "Japan"
                       for spot in feed["Spots"])

        wait_for("spot in both public feeds", lambda: (
            contains_spot(request(primary, "/api/spots"))
            and contains_spot(remote("/api/spots"))
        ))
    finally:
        for container in (peer, CONTAINER):
            subprocess.run(["docker", "network", "disconnect", network, container], check=False)
        docker("network", "rm", network)


def peer_link(primary):
    peer = "dxregia-peer-smoke"
    peer_port, peer_web = 7393, 18808
    peer_base = f"http://{HOST}:{peer_web}"
    network = "dxregia-peer-smoke"
    docker("network", "create", network)
    try:
        docker("network", "connect", network, CONTAINER)
        docker("run", "-d", "--name", peer, "--network", network,
               "-e", f"DXREGIA_TELNET_PORT={peer_port}", "-p", f"{HOST}:{peer_web}:8080", "dxregia:local")
        secondary = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

        def remote(path, body=None):
            return request(secondary, path, body, base=peer_base)

        wait_for("secondary web service", lambda: remote("/api/auth")[0] == 200)
        virgin = docker("exec", peer, "cat", "/data/sysop-initial")
        assert remote("/api/auth/login", virgin)[1]["mustChange"]
        status, enrollment = remote("/api/auth/password", secrets.token_hex(24))
        assert status == 200 and enrollment["mustVerify"]
        assert remote("/api/auth/totp", totp(enrollment["enrollmentKey"]))[0] == 200
        config = {"node": "Q3QQQ-2", "sysop": "Q3QQQ", "locator": "JN55AA", "port": peer_port}
        assert remote("/api/setup", config) == (200, {"started": True, "port": peer_port})

        # DNS names stay valid even when Docker assigns a different IP on restart.
        local_call, peer_call = "Q1QQQ-2", "Q3QQQ-2"
        assert remote("/api/partners", {"Call": local_call, "Host": CONTAINER, "Port": TELNET_PORT})[0] == 200
        assert request(primary, "/api/partners", {"Call": peer_call, "Host": peer, "Port": peer_port})[0] == 200
        status, incoming = remote(f"/api/partners/{local_call}/enable", {})
        assert status == 202 and incoming["Password"]
        assert request(primary, f"/api/partners/{peer_call}/enable", {})[0] == 202

        def local_partner():
            return request(primary, "/api/partners")[1][peer_call]

        wait_for("partner account provisioning", lambda: (
            local_partner()["Provisioning"] == "Abilitato in Mojo"
            and remote("/api/partners")[1][local_call]["Provisioning"] == "Abilitato in Mojo"
        ))
        assert request(primary, f"/api/partners/{peer_call}/connect?retry=true", incoming["Password"])[0] == 202
        wait_for("Mojo link on both dashboards", lambda: (
            local_partner()["Connection"] == "Attivo"
            and remote("/api/partners")[1][local_call]["Connection"] == "Attivo"
        ))
        wait_for("first connection timestamp", lambda: local_partner()["FirstConnectedAt"])
        spot_delivery(primary, remote, peer, peer_port)
        docker("stop", peer)
        wait_for("disconnected partner status", lambda: local_partner()["Connection"] != "Attivo")
        docker("start", peer)
        # Automatic reconnect is throttled, so allow more than the retry interval.
        try:
            wait_for("automatic peer reconnect", lambda: local_partner()["Connection"] == "Attivo", timeout=150)
        except RuntimeError:
            print(f"Primary partner state: {local_partner()!r}", flush=True)
            print(f"Primary listener: {request(primary, '/api/listener')!r}", flush=True)
            print(f"Peer listener: {remote('/api/listener')!r}", flush=True)
            for name in (CONTAINER, peer):
                subprocess.run(["docker", "exec", name, "sh", "-c",
                                "stat -c 'presence modified at %y' /data/mojo/local_data/dxregia-presence.json; "
                                "cat /data/mojo/local_data/dxregia-presence.json; echo; "
                                "ps -o pid,stat,pcpu,comm -C perl; "
                                "grep -aEi 'undefined|died at|redefined|failed at|recursion|DBD::|fatal' "
                                "/data/mojo/local_data/dxregia-startup.log | tail -20; "
                                "find /data/mojo/local_data/debug -type f -name '*.dat' "
                                "-exec grep -aE 'Connect.*(Failed|timed out)|Connected.* to |PC92.*disconn|login failed' {} + "
                                "| tail -40"], check=False)
            raise
    finally:
        subprocess.run(["docker", "rm", "-fv", peer], check=False)
        subprocess.run(["docker", "network", "disconnect", network, CONTAINER], check=False)
        docker("network", "rm", network)


def main():
    client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    virgin = (
        privileged("cat", "/var/lib/dxregia/sysop-initial")
        if NATIVE else docker("exec", CONTAINER, "cat", "/data/sysop-initial")
    )
    config = {"node": "Q1QQQ-2", "sysop": "Q1QQQ", "locator": "JN55AA", "port": TELNET_PORT}
    assert request(client, "/api/setup", config)[0] == 401
    status, login = request(client, "/api/auth/login", virgin)
    assert status == 200 and login["mustChange"]
    password = secrets.token_hex(24)
    status, enroll = request(client, "/api/auth/password", password)
    assert status == 200 and enroll["mustVerify"]
    assert request(client, "/api/setup", config)[0] == 401
    first_code = totp(enroll["enrollmentKey"])
    assert request(client, "/api/auth/totp", first_code)[0] == 200
    assert request(client, "/api/setup", config) == (200, {"started": True, "port": TELNET_PORT})
    portal_traffic(client)
    telnet_presence(client)
    if NATIVE and os.environ.get("DXREGIA_TELNET_BIND") == "0.0.0.0":
        # The loopback Telnet check alone also passes with a loopback-only listener.
        listeners = subprocess.check_output(["ss", "-H4ltn", f"sport = :{TELNET_PORT}"], text=True)
        assert any(row.split()[3] in {f"0.0.0.0:{TELNET_PORT}", f"*:{TELNET_PORT}"}
                   for row in listeners.splitlines()), listeners
    edited = dict(config, name="Lab sysop", qth="Profilo Linux", email="lab@example.org", advertisedHost="node.example.org")
    assert request(client, "/api/node-settings", dict(edited, port=TELNET_PORT + 1), method="PUT")[0] == 409
    assert request(client, "/api/node-settings", edited, method="PUT") == (200, edited)
    # Snapshot hook applies the updated profile during the next five-second cycle.
    time.sleep(6)
    telnet_login(b"Q8QQQ", b"Profilo Linux")
    assert request(client, "/api/spots")[1]["Available"]
    service("stop")
    with tempfile.TemporaryDirectory(prefix="dxregia-smoke-") as folder:
        base = Path(folder)
        source, restored, archive = base / "mojo", base / "restored", base / "node.zip"
        copy_node(source)
        for arguments in [("backup", source, archive), ("verify", archive), ("restore", archive, restored)]:
            backup_command(base, *arguments)
        reject_corrupt_backup(base, archive)
        assert (restored / "local/DXVars.pm").read_bytes() == (source / "local/DXVars.pm").read_bytes()
        assert (restored / "local_data/dxusers.db").read_bytes() == (source / "local_data/dxusers.db").read_bytes()
        # Emulate an existing Mojo installation on this disposable copy.
        for marker in ("dxregia-node.json", "dxregia-initialized"):
            (restored / "local" / marker).unlink()
        adopted = base / "adopted"
        migration = Path(__file__).with_name("migrate_mojo.py")
        subprocess.run([sys.executable, "-B", str(migration), str(restored), str(adopted)], check=True)
        assert (adopted / "local/dxregia-initialized").is_file()
        assert (adopted / "local_data/dxusers.db").read_bytes() == (source / "local_data/dxusers.db").read_bytes()
        service("start")
        ready(client)
        assert request(client, "/api/visits") == (200, {"Views": 2})
        time.sleep(6)
        telnet_login(b"Q7QQQ", b"Profilo Linux")
        # HTTP sessions are held in memory, so authenticate again after restarting.
        assert request(client, "/api/auth/login", password)[0] == 200

        def next_code():
            code = totp(enroll["enrollmentKey"])
            return code if code != first_code else None

        assert request(client, "/api/auth/totp", wait_for("next TOTP window", next_code))[0] == 200

        def verify_upload(opener, content):
            upload = urllib.request.Request(
                BASE + "/api/backups/verify", data=content,
                headers={"Content-Type": "application/zip"},
            )
            try:
                with opener.open(upload, timeout=30) as response:
                    return response.status
            except urllib.error.HTTPError as error:
                return error.code

        # The server can reject an unauthenticated request before receiving its body.
        assert verify_upload(urllib.request.build_opener(), b"zip") == 401
        assert verify_upload(client, archive.read_bytes()) == 200
        assert verify_upload(client, (base / "damaged.zip").read_bytes()) == 422
    print("Linux installer, setup with TOTP, Telnet, backup upload and persistent restart passed")
    if not NATIVE:
        peer_link(client)


if __name__ == "__main__":
    main()
