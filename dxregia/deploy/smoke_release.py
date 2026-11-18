"""Test release download and checksum validation without a published release."""
import hashlib
import os
from pathlib import Path
import platform
import subprocess
import tarfile
import tempfile


def main():
    installer = Path(__file__).with_name("install-release.sh").resolve()
    arch = {"x86_64": "amd64", "aarch64": "arm64"}[platform.machine()]
    bundle_name = f"dxregia-linux-{arch}"
    with tempfile.TemporaryDirectory(prefix="dxregia-release-smoke-") as work:
        root = Path(work)
        bundle = root / bundle_name
        stub = bundle / "dxregia" / "deploy" / "install-native.sh"
        stub.parent.mkdir(parents=True)
        stub.write_text('printf "%s\\n" "installer invoked" > "$SMOKE_RESULT"\n')
        archive = root / f"{bundle_name}.tar.gz"
        with tarfile.open(archive, "w:gz") as output:
            output.add(bundle, arcname=bundle_name)
        checksum = archive.with_suffix(".gz.sha256")
        checksum.write_text(f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n")

        curl = root / "curl"
        curl.write_text("""#!/bin/sh
set -eu
while [ "$#" -gt 1 ]; do
    case "$1" in
        --output) output=$2; shift 2 ;;
        --proto|--connect-timeout|--retry) shift 2 ;;
        -fsSL) shift ;;
        *) exit 2 ;;
    esac
done
case "$1" in
    https://github.com/wifi75/dxregia/releases/download/v1.0.0/*) ;;
    *) exit 2 ;;
esac
cp -- "$SMOKE_ASSETS/${1##*/}" "$output"
""")
        curl.chmod(0o755)
        result = root / "result.txt"
        env = dict(os.environ, PATH=str(root) + os.pathsep + os.environ["PATH"],
                   SMOKE_ASSETS=str(root), SMOKE_RESULT=str(result))
        command = ["sh", str(installer), "v1.0.0"]
        subprocess.run(command, env=env, check=True)
        assert result.read_text().strip() == "installer invoked"
        result.unlink()

        checksum.write_text(f"{'0' * 64}  {archive.name}\n")
        failed = subprocess.run(command, env=env, capture_output=True, text=True)
        assert failed.returncode != 0 and "Verifica del pacchetto fallita" in failed.stderr
        assert not result.exists()


if __name__ == "__main__":
    main()
