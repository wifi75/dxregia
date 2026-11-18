"""Smoke-test a fresh DXRegia Docker installation on a disposable Linux runner."""

import base64
import hmac
import http.cookiejar
import json
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
BASE = f"http://{HOST}:8080"
CONTAINER = "dxregia"


def docker(*args):
    return subprocess.run(["docker", *args], check=True, stdout=subprocess.PIPE, text=True).stdout


def request(client, path, body=None):
    req = urllib.request.Request(
        BASE + path,
        data=None if body is None else json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
    )
    try:
        with client.open(req, timeout=30) as response:
            return response.status, json.load(response)
    except urllib.error.HTTPError as error:
        return error.code, None


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


def telnet_login(callsign):
    with socket.create_connection((HOST, 7300), timeout=5) as conn:
        read_until(conn, b"login:")
        conn.sendall(callsign + b"\r\n")
        assert callsign in read_until(conn, b"dxspider >")
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
    virgin = docker("exec", CONTAINER, "cat", "/data/sysop-initial")
    config = {"node": "Q1QQQ-2", "sysop": "Q1QQQ", "locator": "JN55AA", "port": 7300}
    assert request(client, "/api/setup", config)[0] == 401
    status, login = request(client, "/api/auth/login", virgin)
    assert status == 200 and login["mustChange"]
    password = secrets.token_hex(24)
    status, enroll = request(client, "/api/auth/password", password)
    assert status == 200 and enroll["mustVerify"]
    assert request(client, "/api/setup", config)[0] == 401
    assert request(client, "/api/auth/totp", totp(enroll["enrollmentKey"]))[0] == 200
    assert request(client, "/api/setup", config) == (200, {"started": True, "port": 7300})
    telnet_login(b"Q7QQQ")
    assert request(client, "/api/spots")[1]["Available"]
    docker("stop", CONTAINER)
    with tempfile.TemporaryDirectory(prefix="dxregia-smoke-") as folder:
        base = Path(folder)
        source, restored, archive = base / "mojo", base / "restored", base / "node.zip"
        docker("cp", CONTAINER + ":/data/mojo", str(source))
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
    docker("start", CONTAINER)
    ready(client)
    telnet_login(b"Q7QQQ")
    print("Linux installer, setup with TOTP, Telnet, offline backup and persistent restart passed")


if __name__ == "__main__":
    main()
