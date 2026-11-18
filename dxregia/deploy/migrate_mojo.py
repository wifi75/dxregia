#!/usr/bin/env python3
"""Prepare a separate DXRegia-managed copy of an existing Mojo node."""

import argparse
from contextlib import closing
import json
import os
import re
import sqlite3
import sys
import tempfile
import zipfile
from pathlib import Path

from backup_node import backup, restore

ASSIGNMENT = re.compile(
    r"""^\s*(?:our\s+)?\$([a-zA-Z]\w*)\s*=\s*(?:"([^"$@\\\r\n]*)"|'((?:\\['\\]|[^'\\\r\n])*)'|([+-]?\d+(?:\.\d+)?))\s*;\s*(?:#[^\n]*)?$""",
    re.M,
)
LISTENER = re.compile(r"""\[\s*"([^"\r\n]+)"\s*,\s*(\d{1,5})\s*\]|\[\s*'([^'\r\n]+)'\s*,\s*(\d{1,5})\s*\]""")
CALL = re.compile(r"[A-Z0-9]+[0-9][A-Z0-9]+(?:-(?:[1-9]|1[0-5]))?")
GRID = re.compile(r"[A-R]{2}[0-9]{2}(?:[A-X]{2})?")
OPTIONAL = {"myname": "name", "myqth": "qth", "myemail": "email"}


def read_settings(root):
    config_path = root / "local/DXVars.pm"
    source = config_path.read_text(encoding="utf-8-sig")
    values = {}
    for key, quoted, single, number in ASSIGNMENT.findall(source):
        values[key] = re.sub(r"\\(['\\])", r"\1", single) if single else quoted or number
    required = {}
    for old, new in (("mycall", "node"), ("myalias", "sysop"), ("mylocator", "locator")):
        if not values.get(old):
            raise ValueError(f"Valore letterale mancante in DXVars.pm: {old}")
        required[new] = values[old].upper()
    if (
        not CALL.fullmatch(required["node"]) or len(required["node"]) > 12
        or not CALL.fullmatch(required["sysop"]) or len(required["sysop"]) > 12
        or required["node"] == required["sysop"] or not GRID.fullmatch(required["locator"])
    ):
        raise ValueError("Nominativi o locator incompatibili con il pannello DXRegia.")

    listeners_path = root / "local/Listeners.pm"
    if not listeners_path.is_file():
        listeners_path = root / "perl/Listeners.pm"
    listeners = listeners_path.read_text(encoding="utf-8-sig")
    active = "\n".join(line.split("#", 1)[0] for line in listeners.splitlines())
    matches = LISTENER.findall(active)
    if len(matches) != 1 or active.count("[") != 1:
        raise ValueError("Serve un singolo listener Telnet letterale per la migrazione gestita.")
    host, port, host_single, port_single = matches[0]
    port = int(port or port_single)
    if not 1 <= port <= 65535 or (host or host_single) not in ("0.0.0.0", "127.0.0.1", "::"):
        raise ValueError("Il listener deve essere raggiungibile anche dal loopback.")
    required["port"] = port
    for old, new in OPTIONAL.items():
        if value := values.get(old):
            required[new] = value
    if "mylatitude" in values and "mylongitude" in values:
        lat, lon = float(values["mylatitude"]), float(values["mylongitude"])
        if not -90 <= lat <= 90 or not -180 <= lon <= 180:
            raise ValueError("Coordinate in DXVars.pm fuori intervallo.")
        required.update(latitude=lat, longitude=lon)
    return required


def check_users(root, settings):
    path = root / "local_data/dxusers.db"
    if not path.is_file():
        raise ValueError("Serve la banca utenti SQLite di Mojo in local_data/dxusers.db.")
    with closing(sqlite3.connect(path.resolve().as_uri() + "?mode=ro", uri=True)) as database:
        if database.execute("PRAGMA quick_check").fetchall() != [("ok",)]:
            raise ValueError("Il controllo di integrità SQLite non è riuscito.")
        for key in ("node", "sysop"):
            if not database.execute("SELECT 1 FROM users WHERE call = ?", (settings[key],)).fetchone():
                raise ValueError(f"Account {key} assente nella banca utenti.")


def migrate(source, destination):
    source = source.resolve(strict=True)
    destination = destination.absolute()
    if destination.exists() or destination.is_symlink() or destination.resolve().is_relative_to(source):
        raise ValueError("Scegli una directory di destinazione nuova, fuori dal nodo sorgente.")
    if (source / "local/dxregia-node.json").exists():
        raise ValueError("Il nodo è già gestito da DXRegia. Usa backup e ripristino.")
    with tempfile.TemporaryDirectory(dir=destination.parent, prefix=".dxregia-migrate-") as folder:
        work = Path(folder)
        archive, staging = work / "migration.zip", work / "mojo"
        count = backup(source, archive)
        restore(archive, staging)
        config = read_settings(staging)
        check_users(staging, config)
        local = staging / "local"
        for filename, content in (
            ("dxregia-node.json", json.dumps(config, ensure_ascii=False) + "\n"),
            ("dxregia-initialized", ""),
        ):
            fd = os.open(local / filename, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
            with os.fdopen(fd, "w", encoding="utf-8") as file:
                file.write(content)
        if destination.exists() or destination.is_symlink():
            raise FileExistsError(destination)
        staging.rename(destination)
    return config, count


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("destination", type=Path)
    args = parser.parse_args()
    try:
        config, count = migrate(args.source, args.destination)
    except (OSError, ValueError, RuntimeError, KeyError, sqlite3.Error, zipfile.BadZipFile) as error:
        print(f"Errore: {error}", file=sys.stderr)
        return 1
    print(f"Copia preparata: {config['node']} · Telnet {config['port']} · {count} file")
    return 0


if __name__ == "__main__":
    sys.exit(main())
