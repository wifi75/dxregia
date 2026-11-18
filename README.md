# DXRegia

[![Versione](https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&color=147d72)](https://github.com/wifi75/dxregia/releases)
[![Test Linux](https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&label=Test%20Linux&logo=githubactions&logoColor=white)](https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml)
[![Licenza: Artistic 2.0](https://img.shields.io/badge/Licenza-Artistic%202.0-0b6e99)](LICENSE)
[![Stato: in sviluppo](https://img.shields.io/badge/Stato-in%20sviluppo-e0a23b)](TASK.md)
[![Ultimo commit](https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569)](https://github.com/wifi75/dxregia/commits/main)
[![Manutenzione: attiva 2026](https://img.shields.io/badge/Manutenzione-attiva%202026-2ea44f)](TASK.md)

[![Backend: Go 1.22](https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&logoColor=white)](dxregia/backend/README.md)
[![Pannello: Svelte 5](https://img.shields.io/badge/Pannello-Svelte%205-FF3E00?logo=svelte&logoColor=white)](dxregia/web/README.md)
[![TypeScript 5.9](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Vite 7](https://img.shields.io/badge/Vite-7-646CFF?logo=vite&logoColor=white)](https://vite.dev/)
[![Motore: DXSpider Mojo](https://img.shields.io/badge/Motore-DXSpider%20Mojo-147d72)](https://github.com/EA3CV/dxspider/tree/mojo)
[![Perl 5.38](https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&logoColor=white)](https://www.perl.org/)
[![Console: Telnet](https://img.shields.io/badge/Console-Telnet-334155)](docs/INTERFACES.md)

[![Destinazione: Linux](https://img.shields.io/badge/Destinazione-Linux-334155?logo=linux&logoColor=white)](docs/INSTALLATION.md)
[![Collaudato: Ubuntu 24.04 amd64 e arm64](https://img.shields.io/badge/Collaudato-Ubuntu%2024.04%20%C2%B7%20amd64%20%C2%B7%20arm64-E95420?logo=ubuntu&logoColor=white)](docs/INSTALLATION.md)
[![Docker disponibile](https://img.shields.io/badge/Docker-disponibile-2496ED?logo=docker&logoColor=white)](dxregia/deploy/README.md)
[![Lingue: IT EN ES DE FR PT](https://img.shields.io/badge/Lingue-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e)](dxregia/web/src/translations.ts)
[![SemVer 2.0.0](https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&logoColor=white)](https://semver.org/lang/it/)
[![Keep a Changelog](https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&logoColor=white)](CHANGELOG.md)

**Un nodo DX Cluster per Linux con pannello web, basato su DXSpider Mojo.**

DXRegia nasce per semplificare l'installazione e la gestione di un nodo DX Cluster. Il progetto mantiene DXSpider Mojo come motore per la rete, i collegamenti Telnet e i comandi classici, e prevede un pannello web per amministrare il nodo e una pagina pubblica per seguire gli spot.

**Stato del progetto:** sviluppo iniziale. Il repository contiene la base DXSpider Mojo, le specifiche e un controllo Telnet da terminale. La [diagnostica Go/Svelte e il wizard locale](dxregia/backend/README.md) consentono di configurare e avviare Mojo dal browser. Avvio e ripartenza con dati persistenti sono stati verificati con Mojo reale su Windows. Installatore, pannello completo e feed pubblico sono in sviluppo. Il controllo Telnet è stato provato con socket locali. Il collaudo con Mojo su Linux e con i logger è in programma.

## Installazione e configurazione

- [Guida all'installazione e configurazione](docs/INSTALLATION.md) — come orientarsi oggi, come verificare un nodo Mojo già avviato e quale sarà il flusso di primo accesso a DXRegia.
- [Specifiche dell'installatore Linux](dxregia/deploy/README.md) — script unico, indirizzo del pannello, password temporanea, cambio al primo login e conservazione dei dati.
- [Console Telnet e interfacce web](docs/INTERFACES.md) — accesso classico, pagina pubblica e pannello sysop.

## Funzioni previste

| Area | Funzioni |
| --- | --- |
| **Spot pubblici** | Feed in diretta, ricerca per nominativo, filtri per banda e modo, orario UTC e host Telnet copiabile. |
| **Pannello sysop** | Configurazione guidata, gestione utenti e filtri, stato del servizio e console web con cronologia. |
| **Nodi partner** | Aggiunta con nominativo, host e porta effettiva. Verifica della raggiungibilità, connessione, riconnessione e cronologia degli errori. |
| **Registro eventi** | Ricerca e filtri per categoria, gravità, intervallo e nominativo. Esportazione con dati sensibili oscurati. |
| **Backup e ripristino** | Revisioni della configurazione, backup manuali e pianificati e verifica dell'archivio prima di ripristinare. |
| **Compatibilità** | Accesso da client Telnet e logger, comandi DXSpider e collegamenti con i nodi esistenti da verificare nel banco di prova. |

[Specifiche operative del sysop](docs/OPERATIONS.md) · [Analisi di DXSpider e priorità](docs/PRODUCT-ANALYSIS.md)

## Sviluppo

Il motore Perl resta nei percorsi upstream. Il codice specifico di DXRegia vive in [dxregia/](dxregia/README.md), con percorsi per il backend Go, il frontend Svelte/TypeScript e la distribuzione Linux.

- [Piano di lavoro](TASK.md) e [passaggio di consegne](docs/HANDOFF.md)
- [Architettura](docs/ARCHITECTURE.md) e [registro delle decisioni](docs/DECISIONS.md)
- [Istruzioni per Codex](AGENTS.md) e [Claude Code](CLAUDE.md)
- [Issue GitHub](https://github.com/wifi75/dxregia/issues) per segnalazioni e proposte

## Origine e licenze

DXRegia è un progetto indipendente basato su DXSpider Mojo, di Dirk Koopman G1TLH e dei suoi contributori. La base proviene dal [mirror EA3CV](https://github.com/EA3CV/dxspider/tree/mojo). Il repository conserva la cronologia e i file originali [COPYING](COPYING) e [LICENSE](LICENSE). Le licenze applicabili ai componenti della distribuzione DXRegia saranno verificate prima della prima release installabile.
