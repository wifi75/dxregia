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


def request(client, path, body=None, method=None):
    req = urllib.request.Request(
        BASE + path,
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
    assert request(client, "/api/auth/totp", totp(enroll["enrollmentKey"]))[0] == 200
    assert request(client, "/api/setup", config) == (200, {"started": True, "port": TELNET_PORT})
    portal_traffic(client)
    telnet_login(b"Q7QQQ")
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
        tool = Path(__file__).with_name("backup_node.py")
        for arguments in [("backup", source, archive), ("verify", archive), ("restore", archive, restored)]:
            subprocess.run([sys.executable, str(tool), *map(str, arguments)], check=True)
        assert (restored / "local/DXVars.pm").read_bytes() == (source / "local/DXVars.pm").read_bytes()
        assert (restored / "local_data/dxusers.db").read_bytes() == (source / "local_data/dxusers.db").read_bytes()
        # Emulate an existing Mojo installation on this disposable copy.
        for marker in ("dxregia-node.json", "dxregia-initialized"):
            (restored / "local" / marker).unlink()
        adopted = base / "adopted"
        migration = tool.with_name("migrate_mojo.py")
        subprocess.run([sys.executable, "-B", str(migration), str(restored), str(adopted)], check=True)
        assert (adopted / "local/dxregia-initialized").is_file()
        assert (adopted / "local_data/dxusers.db").read_bytes() == (source / "local_data/dxusers.db").read_bytes()
    service("start")
    ready(client)
    assert request(client, "/api/visits") == (200, {"Views": 2})
    time.sleep(6)
    telnet_login(b"Q7QQQ", b"Profilo Linux")
    print("Linux installer, setup with TOTP, Telnet, offline backup and persistent restart passed")


if __name__ == "__main__":
    main()
