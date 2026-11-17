"""Check the DXSpider login banner without logging in or sending commands."""

import argparse
import socket
import sys
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("host")
    parser.add_argument("port", type=int)
    args = parser.parse_args()
    if not 1 <= args.port <= 65535:
        parser.error("port must be between 1 and 65535")

    timeout = 5
    login_prompt = b"login: "
    try:
        with socket.create_connection((args.host, args.port), timeout) as connection:
            deadline = time.monotonic() + timeout
            banner = bytearray()
            while len(banner) < 65_536:
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    break
                connection.settimeout(remaining)
                chunk = connection.recv(min(4096, 65_536 - len(banner)))
                if not chunk:
                    break
                banner.extend(chunk)
                if login_prompt in banner:
                    print("OK: DXSpider login prompt received")
                    return 0
    except OSError as error:
        print(f"Telnet check failed: {error}", file=sys.stderr)
        return 1

    print("Telnet check failed: DXSpider login prompt not found", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
