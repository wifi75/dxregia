<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

**English** · [Italiano](README.it.md) · [Español](README.es.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Português](README.pt.md)

**A DX Cluster node for Linux with a web panel, built on DXSpider Mojo.**

DXRegia makes it easy to install, run and share a DX Cluster node. It keeps DXSpider Mojo as the network engine (classic Telnet access, commands and node-to-node links) and adds a modern web panel and a public live spots page.

## See DXRegia in action

Screenshots from DXRegia using sample data. Select an image to enlarge it.

### Live spots · public portal

At `/spots`, radio amateurs can follow live DX spots, filter by band or mode, and identify stations by their DXCC flags.

[![Live spots · public portal](docs/screenshots/spots.png)](docs/screenshots/spots.png)

### Node status and sessions · sysop panel

At `/admin`, sysops can see Telnet and cluster status, partner link direction, active sessions, and server resource graphs.

[![Node status and sessions · sysop panel](docs/screenshots/admin.png)](docs/screenshots/admin.png)

## What it does

- **Live spots page** — public and in real time. Filter by band, mode and callsign (several at once), see DXCC flags, and get a yellow highlight and a soft chirp for every new spot. Works on phone, tablet and desktop, in light or dark theme.
- **Smarter spots** — the mode is inferred from the frequency when the comment does not say it, and announced DX operations (NG3K ADXO) appear next to the callsign.
- **Six languages** — Italian, English, Spanish, German, French and Portuguese; any other language falls back to English.
- **Admin panel at `/admin`** — guided first setup, sign-in with a password and optional two-factor authentication, a live dashboard with status LEDs and an analytics center.
- **Node settings in the panel** — message of the day, scheduled commands, startup script, block lists (words, callsigns, IP addresses), hop counts, filters and Telnet options, each with a tooltip. The panel itself speaks the six languages and shows who connected to the node.
- **Registration to send spots** — visitors register on the portal (callsign, email, password) and confirm by email; only registered users can send spots over Telnet. Optional, switched on in the panel.
- **Install as an app** — the portal installs on a phone's home screen and still opens without a connection.
- **Partner nodes** — add, connect and auto-reconnect; see who started each link; delete a node to disconnect it and revoke its access.
- **Link requests on autopilot** — other sysops ask from your Contact page; DXRegia can create their account, email your connection details and confirm when they connect (a switch and a daily limit in the panel). For CC Cluster peers, a public IP or DNS hostname can authorize inbound access without a password. The confirmation email follows an active inbound session.
- **HTML emails** — instructions and confirmations in six languages through your own SMTP server, set up and tested in the panel.
- **Classic access kept** — Telnet and DXSpider commands for loggers and users, compatible with existing nodes.
- **Safe to run** — offline backups with verification, one-command updates with automatic rollback, and visitor statistics without cookies or stored addresses.
- **Easy to install** — one script on Debian or Ubuntu (systemd, amd64 or arm64), or Docker.

- Save and compare spot views, explore DXCC and callsign history, and set local radio alerts that run while the portal is open.
- **Propagation and QSL** — see HamQSL solar indices and HF day/night and VHF conditions, the KC2G MUF map, recent cluster spots by region and band, and WWV/WCY bulletins. Live spots have a QSL column for LoTW and eQSL AG. The registration page explains how to connect by Telnet from seven documented loggers. Sysops can edit spot rules under Hop and filters.

## Quick start

On Debian or Ubuntu with systemd, install the prebuilt package for amd64 or arm64. You need curl and sudo, or a root shell.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.7.1/dxregia/deploy/install-release.sh
sh install-release.sh v0.7.1
```

1. Open the LAN address printed by the installer, for example `http://SERVER:8080/admin`. A fresh native installation listens on all IPv4 interfaces.
2. Sign in as `sysop` with the temporary password shown by the installer and choose a new password. Two-factor authentication is optional and off by default. To turn it on, go to **Management → Password → Two-step verification**, confirm your web password, scan the QR code and save the recovery codes.
3. Fill in the first-setup wizard: node callsign, sysop callsign, locator and Telnet port. When it finishes, the spots page is at `/spots` and Telnet answers on port 7300.

The guide also covers Docker and installation from source.

The step-by-step guide covers ports, mail, partner nodes, updates, backups and troubleshooting: [**Installation and configuration guide**](docs/INSTALLATION.md).

## More documentation

_The technical documents below are written in Italian._

- [Quick guides, protocol reference, Raspberry Pi and migration](docs/QUICKSTART.md) (Italian)
- [Peering examples in six languages](docs/PEERING.md)
- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Sysop operations: partner nodes, logs, backups
- [docs/INTERFACES.md](docs/INTERFACES.md) — Public page, admin panel, Telnet and brand
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Linux installer reference
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Architecture
- [TASK.md](TASK.md) — Work plan
- [CHANGELOG.md](CHANGELOG.md) — Changelog

**Repository layout:** `dxregia/` is DXRegia's own code (backend, panel, installers), `engine/` is the DXSpider Mojo engine with its original structure, and `docs/` holds the documentation.

## Status

DXRegia is in active development (version 0.7.1). Installation, first sign-in, Telnet, backup and restart are tested on Ubuntu 24.04 (amd64 and arm64). Still open: a Proxmox LXC test, delivery with more mail providers and a signed release package. Bug reports and ideas are welcome in [GitHub Issues](https://github.com/wifi75/dxregia/issues).

## Credits and licence

DXRegia is an independent project based on DXSpider Mojo by Dirk Koopman G1TLH and contributors (the [EA3CV mirror](https://github.com/EA3CV/dxspider/tree/mojo)). The repository keeps the original [COPYING](COPYING) and [LICENSE](LICENSE) (Artistic License 2.0). Announced DX operations come from [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); flags are from [flag-icons](https://github.com/lipis/flag-icons) (MIT).

[Licenses and third-party notices](docs/LICENSING.md)
