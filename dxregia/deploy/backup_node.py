#!/usr/bin/env python3
"""Back up a stopped DXRegia/Mojo node or restore it to a new directory."""

import argparse
import hashlib
import json
import os
import re
import sqlite3
import sys
import tempfile
import zipfile
from pathlib import Path, PurePosixPath

MANIFEST = "dxregia-manifest.json"
ROOTS = {
    "perl", "cmd", "connect", "data", "filter", "msg", "sys", "local", "local_cmd",
    "local_data", "perl5lib", "COPYING", "LICENSE",
    ".dxregia-sysop", ".dxregia-totp", ".dxregia-partners.json",
    ".dxregia-contacts.json", ".dxregia-traffic.json",
}
VOLATILE = {"cluster.lck", "dxregia-presence.json", "dxregia-presence.json.tmp"}
JOBS = {"enable", "connect", "script", "result", "connect-result", "tmp", "mail.json"}
RESERVED = re.compile(r"^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)", re.I)
MAX_MANIFEST = 8 * 1024 * 1024
# Files SQLite keeps next to the database while the node runs: never part of a hot backup.
SQLITE_SIDE = ("-journal", "-wal", "-shm")
DATABASE = "local_data/dxusers.db"


def allowed(name):
    path = PurePosixPath(name)
    if str(path) != name or not path.parts or path.parts[0] not in ROOTS:
        return False
    if any(
        part in (".", "..") or RESERVED.match(part) or part.endswith((".", " "))
        or any(char in part for char in '\\:*?"<>|') or any(ord(char) < 32 for char in part)
        for part in path.parts
    ):
        return False
    if len(path.parts) == 2 and path.parts[0] == "local_data" and (
        path.name == ".dxregia-telnet-password" or path.name.startswith(".telnet-password-")
        or path.name in {"dxregia-telnet-password.result", "dxregia-telnet-password.result.tmp"}
    ):
        return False
    if path.parts[-1] in VOLATILE:
        return False
    if len(path.parts) == 3 and path.parts[:2] == ("local_data", "dxregia-partners"):
        return not any(path.name.endswith("." + ending) for ending in JOBS) and not path.name.startswith(".")
    return True


def paths(root):
    for name in sorted(ROOTS):
        current = root / name
        if current.is_symlink():
            raise ValueError(f"Collegamento simbolico non supportato: {current}")
        candidates = current.rglob("*") if current.is_dir() else (current,)
        for file in candidates:
            relative = file.relative_to(root).as_posix()
            if not allowed(relative):
                continue
            if file.is_symlink():
                target = file.resolve(strict=True)
                if not target.is_relative_to(root) or not target.is_file() or not allowed(target.relative_to(root).as_posix()):
                    raise ValueError(f"Collegamento simbolico non supportato: {file}")
                yield relative, target
            elif file.is_file():
                yield relative, file


def copy_hash(source, destination):
    digest = hashlib.sha256()
    count = 0
    while chunk := source.read(1024 * 1024):
        if destination is not None:
            destination.write(chunk)
        digest.update(chunk)
        count += len(chunk)
    return {"sha256": digest.hexdigest(), "size": count}


def snapshot_database(source, directory):
    """Consistent copy of the live SQLite database, checked before it goes into an archive."""
    target = Path(directory) / "dxusers.snapshot"
    origin = sqlite3.connect(f"file:{source}?mode=ro", uri=True, timeout=30)
    try:
        copy = sqlite3.connect(target)
        try:
            origin.backup(copy)
            if copy.execute("PRAGMA integrity_check").fetchone()[0] != "ok":
                raise ValueError("La banca utenti non ha superato il controllo di integrità.")
        finally:
            copy.close()
    finally:
        origin.close()
    return target


def backup(root, output, online=False):
    root = root.resolve(strict=True)
    if not (root / "perl/cluster.pl").is_file() or not (root / "local/DXVars.pm").is_file():
        raise ValueError("Indica la directory di un nodo Mojo configurato.")
    if not online and (root / "local_data/cluster.lck").exists():
        raise ValueError("Mojo ha un file di blocco. Arresta il nodo e il backend prima del backup, oppure usa --online.")
    output = output.resolve()
    if output.is_relative_to(root):
        raise ValueError("Salva l'archivio fuori dalla directory del nodo.")
    if output.exists():
        raise FileExistsError(output)
    with tempfile.NamedTemporaryFile(dir=output.parent, prefix=".dxregia-backup-", delete=False) as temp:
        staging = Path(temp.name)
    try:
        index = {}
        with tempfile.TemporaryDirectory(dir=output.parent, prefix=".dxregia-snapshot-") as scratch, \
                zipfile.ZipFile(staging, "w", compression=zipfile.ZIP_DEFLATED) as archive:
            for name, file in paths(root):
                if online and name.endswith(SQLITE_SIDE) and name.startswith("local_data/"):
                    continue
                if online and name == DATABASE:
                    file = snapshot_database(file, scratch)
                with file.open("rb") as source, archive.open(name, "w", force_zip64=True) as target:
                    index[name] = copy_hash(source, target)
            manifest = json.dumps({"format": "dxregia-backup-v1", "files": index}).encode()
            if len(manifest) > MAX_MANIFEST:
                raise ValueError("L'indice del backup supera la dimensione supportata.")
            archive.writestr(MANIFEST, manifest)
        if not online and (root / "local_data/cluster.lck").exists():
            raise ValueError("Mojo è partito durante il backup.")
        os.link(staging, output)  # publish without replacing an existing backup
    finally:
        staging.unlink(missing_ok=True)
    return len(index)


def read_index(archive):
    info = archive.getinfo(MANIFEST)
    if info.file_size > MAX_MANIFEST:
        raise ValueError("Manifest troppo grande.")
    manifest = json.loads(archive.read(info))
    if not isinstance(manifest, dict) or manifest.get("format") != "dxregia-backup-v1":
        raise ValueError("Formato di backup non supportato.")
    index = manifest.get("files")
    if not isinstance(index, dict) or not {"perl/cluster.pl", "local/DXVars.pm"}.issubset(index):
        raise ValueError("Manifest incompleto.")
    names = archive.namelist()
    if len(names) != len(index) + 1 or set(names) != set(index) | {MANIFEST}:
        raise ValueError("I file nell'archivio non corrispondono al manifest.")
    if len({name.casefold() for name in index}) != len(index):
        raise ValueError("Percorsi duplicati nel backup.")
    for name, expected in index.items():
        if not allowed(name) or archive.getinfo(name).is_dir() or not isinstance(expected, dict):
            raise ValueError(f"Percorso non valido: {name}")
        if archive.getinfo(name).file_size != expected.get("size"):
            raise ValueError(f"Dimensione non valida: {name}")
    return index


def restore(input_path, destination=None):
    if destination is not None:
        destination = destination.absolute()
        if destination.exists() or destination.is_symlink():
            raise FileExistsError(destination)
    with zipfile.ZipFile(input_path) as archive:
        index = read_index(archive)
        with tempfile.TemporaryDirectory(dir=destination.parent if destination else None, prefix=".dxregia-restore-") as folder:
            stage = Path(folder) / "mojo"
            for name, expected in index.items():
                with archive.open(name) as source:
                    if destination is None:
                        actual = copy_hash(source, None)
                    else:
                        path = stage.joinpath(*PurePosixPath(name).parts)
                        path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
                        fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
                        with os.fdopen(fd, "wb") as target:
                            actual = copy_hash(source, target)
                if actual != expected:
                    raise ValueError(f"Verifica fallita per {name}")
            if destination is not None:
                if destination.exists() or destination.is_symlink():
                    raise FileExistsError(destination)
                stage.rename(destination)
    return len(index)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    subcommands = parser.add_subparsers(dest="action", required=True)
    create = subcommands.add_parser("backup")
    create.add_argument("node", type=Path)
    create.add_argument("archive", type=Path)
    create.add_argument("--online", action="store_true", help="copia un nodo in funzione, con una copia coerente della banca utenti")
    verify = subcommands.add_parser("verify")
    verify.add_argument("archive", type=Path)
    recover = subcommands.add_parser("restore")
    recover.add_argument("archive", type=Path)
    recover.add_argument("destination", type=Path)
    args = parser.parse_args()
    try:
        if args.action == "backup":
            count = backup(args.node, args.archive, args.online)
        else:
            count = restore(args.archive, getattr(args, "destination", None))
    except (OSError, ValueError, KeyError, zipfile.BadZipFile, RuntimeError) as error:
        print(f"Errore: {error}", file=sys.stderr)
        return 1
    print(f"Operazione completata. File: {count}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
