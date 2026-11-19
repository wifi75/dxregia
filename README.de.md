<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · [Italiano](README.it.md) · [Español](README.es.md) · **Deutsch** · [Français](README.fr.md) · [Português](README.pt.md)

**Ein DX-Cluster-Knoten für Linux mit Web-Oberfläche, auf Basis von DXSpider Mojo.**

DXRegia macht es einfach, einen DX-Cluster-Knoten zu installieren, zu betreiben und zu teilen. Es behält DXSpider Mojo als Netzwerk-Engine (klassischer Telnet-Zugang, Befehle und Knotenverbindungen) und ergänzt eine moderne Web-Oberfläche und eine öffentliche Seite mit Live-Spots.

## Was es kann

- **Live-Spot-Seite** — öffentlich und in Echtzeit. Filter nach Band, Betriebsart und Rufzeichen (auch mehrere zugleich), DXCC-Flaggen, gelbe Hervorhebung und ein sanfter Signalton bei jedem neuen Spot. Läuft auf Smartphone, Tablet und Computer, im hellen oder dunklen Design.
- **Klügere Spots** — die Betriebsart wird aus der Frequenz abgeleitet, wenn der Kommentar sie nicht nennt, und angekündigte DX-Operationen (NG3K ADXO) erscheinen neben dem Rufzeichen.
- **Sechs Sprachen** — Italienisch, Englisch, Spanisch, Deutsch, Französisch und Portugiesisch; jede andere Sprache fällt auf Englisch zurück.
- **Admin-Bereich unter `/admin`** — geführte Ersteinrichtung, Anmeldung mit Benutzer, Passwort und Zwei-Faktor-Code, Live-Dashboard mit Status-LEDs und ein Analytics-Center.
- **Knoteneinstellungen im Panel** — Nachricht des Tages, geplante Befehle, Startskript, Sperrlisten (Wörter, Rufzeichen, IP-Adressen), Hops, Filter und Telnet-Optionen, jeweils mit Tooltip. Das Panel spricht die sechs Sprachen und zeigt, wer sich mit dem Knoten verbunden hat.
- **Registrierung zum Senden von Spots** — Besucher registrieren sich im Portal (Rufzeichen, E-Mail, Passwort) und bestätigen per E-Mail; nur Registrierte können per Telnet Spots senden. Optional, im Panel einzuschalten.
- **Als App installierbar** — das Portal lässt sich auf dem Startbildschirm des Smartphones installieren und öffnet sich auch ohne Verbindung.
- **Partnerknoten** — hinzufügen, verbinden und automatisch wiederverbinden; sehen, wer die Verbindung gestartet hat; einen Knoten löschen, um ihn zu trennen und seinen Zugang zu entziehen.
- **Verbindungsanfragen automatisch** — andere Sysops fragen über deine Kontaktseite an; DXRegia kann ihr Konto anlegen, ihnen deine Verbindungsdaten mailen und bestätigen, sobald sie verbunden sind (Schalter und Tageslimit im Panel).
- **HTML-E-Mails** — Anleitungen und Bestätigungen in sechs Sprachen über deinen eigenen SMTP-Server, im Panel eingerichtet und getestet.
- **Klassischer Zugang bleibt** — Telnet und DXSpider-Befehle für Logger und Benutzer, kompatibel mit bestehenden Knoten.
- **Sicher im Betrieb** — Offline-Backups mit Prüfung, Updates mit einem Befehl und automatischem Rollback, Besucherstatistik ohne Cookies und ohne gespeicherte Adressen.
- **Einfach zu installieren** — ein Skript auf Debian oder Ubuntu (systemd, amd64 oder arm64) oder Docker.

## Schnellstart

Auf Debian oder Ubuntu mit systemd. Du brauchst Git, Go 1.22+, Node.js 22.12+ mit npm, curl und sudo.

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

1. Öffne die vom Installationsskript angezeigte LAN-Adresse, zum Beispiel `http://SERVER:8080/admin`. Eine neue native Installation lauscht auf allen IPv4-Schnittstellen.
2. Melde dich als Benutzer `sysop` mit dem vom Installer angezeigten vorläufigen Passwort an, wähle ein neues Passwort und richte die Zwei-Faktor-Anmeldung mit einer Authenticator-App ein. Bewahre die Wiederherstellungscodes auf.
3. Fülle den Einrichtungsassistenten aus: Rufzeichen des Knotens, Rufzeichen des Sysops, Locator und Telnet-Port. Danach liegt die Spot-Seite unter `/spots` und Telnet antwortet auf Port 7300.

Lieber Docker? Führe stattdessen `sh dxregia/deploy/install-local.sh` aus.

Die Schritt-für-Schritt-Anleitung behandelt Ports, E-Mail, Partnerknoten, Updates, Backups und Fehlersuche: [**Anleitung zu Installation und Konfiguration**](docs/INSTALLATION.de.md).

## Weitere Dokumentation

_Die folgenden technischen Dokumente sind auf Italienisch verfasst._

- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Sysop-Betrieb: Partnerknoten, Logs, Backups
- [docs/INTERFACES.md](docs/INTERFACES.md) — Öffentliche Seite, Admin-Bereich, Telnet und Marke
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Referenz des Linux-Installers
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Architektur
- [TASK.md](TASK.md) — Arbeitsplan
- [CHANGELOG.md](CHANGELOG.md) — Änderungsprotokoll (auf Englisch)

**Aufbau des Repositorys:** `dxregia/` enthält den eigenen Code von DXRegia (Backend, Panel, Installer), `engine/` die DXSpider-Mojo-Engine mit ihrer ursprünglichen Struktur und `docs/` die Dokumentation.

## Stand

DXRegia befindet sich in aktiver Entwicklung (Version 0.4.0). Installation, erste Anmeldung, Telnet, Backup und Neustart sind auf Ubuntu 24.04 (amd64 und arm64) getestet. Offen sind noch: ein Test auf Proxmox-LXC, E-Mail-Zustellung bei weiteren Anbietern und ein signiertes Release-Paket. Fehlermeldungen und Ideen sind in den [GitHub Issues](https://github.com/wifi75/dxregia/issues) willkommen.

## Herkunft und Lizenz

DXRegia ist ein unabhängiges Projekt auf Basis von DXSpider Mojo von Dirk Koopman G1TLH und Mitwirkenden (der [EA3CV-Spiegel](https://github.com/EA3CV/dxspider/tree/mojo)). Das Repository enthält die originalen Dateien [COPYING](COPYING) und [LICENSE](LICENSE) (Artistic License 2.0). Angekündigte DX-Operationen stammen von [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); die Flaggen von [flag-icons](https://github.com/lipis/flag-icons) (MIT).
