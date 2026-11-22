<h1><picture><source media="(prefers-color-scheme: dark)" srcset="dxregia/web/public/brand/dxregia-logo-dark.svg"><img alt="DXRegia" src="dxregia/web/public/brand/dxregia-logo.svg" height="48"></picture></h1>

<p><a href="https://github.com/wifi75/dxregia/releases"><img src="https://img.shields.io/github/v/release/wifi75/dxregia?display_name=tag&amp;color=147d72" alt="Version"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/actions/workflows/linux-smoke.yml"><img src="https://img.shields.io/github/actions/workflow/status/wifi75/dxregia/linux-smoke.yml?branch=main&amp;label=Test%20Linux&amp;logo=githubactions&amp;logoColor=white" alt="Test Linux"></a>&thinsp;<a href="LICENSE"><img src="https://img.shields.io/badge/License-Artistic%202.0-0b6e99" alt="License: Artistic 2.0"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Status-in%20development-e0a23b" alt="Status: in development"></a>&thinsp;<a href="https://github.com/wifi75/dxregia/commits/main"><img src="https://img.shields.io/github/last-commit/wifi75/dxregia?color=475569" alt="Last commit"></a>&thinsp;<a href="TASK.md"><img src="https://img.shields.io/badge/Maintained-yes%202026-2ea44f" alt="Maintained: yes 2026"></a><br>
<a href="dxregia/backend/README.md"><img src="https://img.shields.io/badge/Backend-Go%201.22-00ADD8?logo=go&amp;logoColor=white" alt="Backend: Go 1.22"></a>&thinsp;<a href="dxregia/web/README.md"><img src="https://img.shields.io/badge/Panel-Svelte%205-FF3E00?logo=svelte&amp;logoColor=white" alt="Panel: Svelte 5"></a>&thinsp;<a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&amp;logoColor=white" alt="TypeScript 5.9"></a>&thinsp;<a href="https://vite.dev/"><img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite&amp;logoColor=white" alt="Vite 7"></a>&thinsp;<a href="https://github.com/EA3CV/dxspider/tree/mojo"><img src="https://img.shields.io/badge/Engine-DXSpider%20Mojo-147d72" alt="Engine: DXSpider Mojo"></a>&thinsp;<a href="https://www.perl.org/"><img src="https://img.shields.io/badge/Perl-5.38-39457E?logo=perl&amp;logoColor=white" alt="Perl 5.38"></a>&thinsp;<a href="docs/INTERFACES.md"><img src="https://img.shields.io/badge/Console-Telnet-334155" alt="Console: Telnet"></a><br>
<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Platform-Linux-334155?logo=linux&amp;logoColor=white" alt="Platform: Linux"></a>&thinsp;<a href="docs/INSTALLATION.md"><img src="https://img.shields.io/badge/Tested-Ubuntu%2024.04-E95420?logo=ubuntu&amp;logoColor=white" alt="Tested on Ubuntu 24.04, amd64 and arm64"></a>&thinsp;<a href="dxregia/deploy/README.md"><img src="https://img.shields.io/badge/Docker-supported-2496ED?logo=docker&amp;logoColor=white" alt="Docker supported"></a>&thinsp;<a href="dxregia/web/src/translations.ts"><img src="https://img.shields.io/badge/Languages-IT%20%C2%B7%20EN%20%C2%B7%20ES%20%C2%B7%20DE%20%C2%B7%20FR%20%C2%B7%20PT-0f766e" alt="Languages: IT EN ES DE FR PT"></a><br>
<a href="https://semver.org/"><img src="https://img.shields.io/badge/SemVer-2.0.0-3F4551?logo=semver&amp;logoColor=white" alt="SemVer 2.0.0"></a>&thinsp;<a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-Keep%20a%20Changelog-E05735?logo=keepachangelog&amp;logoColor=white" alt="Keep a Changelog"></a></p>

[English](README.md) · **Italiano** · [Español](README.es.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Português](README.pt.md)

**Un nodo DX Cluster per Linux con pannello web, basato su DXSpider Mojo.**

DXRegia rende semplice installare, gestire e condividere un nodo DX Cluster. Mantiene DXSpider Mojo come motore di rete (accesso Telnet classico, comandi e collegamenti tra nodi) e aggiunge un pannello web moderno e una pagina pubblica con gli spot in diretta.

## Portale e amministrazione

Schermate di DXRegia con dati dimostrativi. Clicca su un’immagine per ingrandirla.

### Spot in diretta · portale pubblico

Su `/spots`, i radioamatori seguono gli spot DX in tempo reale, filtrano per banda o modo e riconoscono le entità dalle bandiere DXCC.

[![Spot in diretta · portale pubblico](docs/screenshots/spots.png)](docs/screenshots/spots.png)

### Stato e sessioni · pannello sysop

Su `/admin`, il sysop vede lo stato di Telnet e del cluster, la direzione dei collegamenti, le sessioni attive e i grafici delle risorse del server.

[![Stato e sessioni · pannello sysop](docs/screenshots/admin.png)](docs/screenshots/admin.png)

## Cosa fa

- **Pagina spot in diretta** — pubblica e in tempo reale. Filtri per banda, modo e nominativo (anche più insieme), bandiere DXCC, evidenziazione gialla e un breve suono per ogni nuovo spot. Funziona su telefono, tablet e computer, con tema chiaro o scuro.
- **Spot più intelligenti** — il modo viene dedotto dalla frequenza quando il commento non lo indica, e le spedizioni DX annunciate (NG3K ADXO) compaiono accanto al nominativo.
- **Sei lingue** — italiano, inglese, spagnolo, tedesco, francese e portoghese; ogni altra lingua ripiega sull'inglese.
- **Pannello admin su `/admin`** — configurazione guidata, accesso con password e 2FA facoltativa, dashboard in diretta con LED di stato e centro di analytics.
- **Impostazioni del nodo nel pannello** — messaggio del giorno, comandi pianificati, script di avvio, liste di blocco (parole, nominativi, indirizzi IP), hop, filtri e opzioni Telnet, ognuna con un tooltip. Il pannello parla le sei lingue e mostra chi si è collegato al nodo.
- **Registrazione per inviare spot** — i visitatori si registrano dal portale (nominativo, email, password) e confermano via email; solo i registrati possono inviare spot via Telnet. Facoltativa, si attiva dal pannello.
- **Installabile come app** — il portale si installa sulla schermata Home dello smartphone e si apre anche senza connessione.
- **Nodi partner** — aggiungi, collega e riconnetti da solo; vedi chi ha avviato ogni collegamento; elimina un nodo per scollegarlo e revocarne l'accesso.
- **Richieste di collegamento in automatico** — altri sysop chiedono dalla tua pagina Contatti; DXRegia può creare il loro account, mandare per email i tuoi dati di collegamento e confermare quando si collegano (interruttore e limite giornaliero nel pannello). Per i partner CC Cluster, puoi autorizzare l’ingresso senza password con un IP pubblico o nome host DNS. La seconda email conferma una sessione in ingresso attiva.
- **Email in HTML** — istruzioni e conferme in sei lingue tramite il tuo server SMTP, configurato e provato dal pannello.
- **Accesso classico mantenuto** — Telnet e comandi DXSpider per logger e utenti, compatibili con i nodi esistenti.
- **Sicuro da gestire** — backup offline con verifica, aggiornamenti con un solo comando e ripristino automatico, statistiche dei visitatori senza cookie né indirizzi salvati.
- **Facile da installare** — un solo script su Debian o Ubuntu (systemd, amd64 o arm64), oppure Docker.

- Salva e confronta le viste degli spot, consulta DXCC e cronologia per nominativo e imposta allarmi radio che funzionano mentre il portale è aperto.
- **Propagazione e QSL** — consulta gli indici solari e le condizioni HF giorno/notte e VHF di HamQSL, la mappa MUF di KC2G, gli spot recenti per area e banda e i bollettini WWV/WCY. La tabella spot ha una colonna QSL per LoTW ed eQSL AG. La registrazione spiega il collegamento Telnet da sette logger documentati. Il sysop modifica le regole spot in Hop e filtri.

## Avvio rapido

Su Debian o Ubuntu con systemd installa il pacchetto precompilato per amd64 o arm64. Servono curl e sudo oppure una shell root.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.14.1/dxregia/deploy/install-release.sh
sh install-release.sh v0.14.1
```

1. Apri l’indirizzo LAN stampato dall’installatore, ad esempio `http://SERVER:8080/admin`. Una nuova installazione nativa ascolta su tutte le interfacce IPv4.
2. Accedi come `sysop` con la password temporanea stampata e scegli una nuova password. Il 2FA è facoltativo e disattivato di default. Per attivarlo vai in **Gestione → Password → Verifica in due passaggi**, conferma la password web, scansiona il QR e salva i codici di recupero.
3. Compila la configurazione guidata: nominativo del nodo, nominativo del sysop, locator e porta Telnet. Al termine la pagina spot è su `/spots` e Telnet risponde sulla porta 7300.

La guida descrive anche Docker e l’installazione da sorgente.

La guida passo passo spiega porte, posta, nodi partner, aggiornamenti, backup e risoluzione dei problemi: [**Guida all'installazione e configurazione**](docs/INSTALLATION.it.md).

## Altra documentazione

_I documenti tecnici qui sotto sono scritti in italiano._

- [Guida rapida, protocolli, Raspberry Pi e migrazione](docs/QUICKSTART.md)
- [Esempi di peering in sei lingue](docs/PEERING.md)
- [docs/OPERATIONS.md](docs/OPERATIONS.md) — Operazioni del sysop: nodi partner, log, backup
- [docs/INTERFACES.md](docs/INTERFACES.md) — Pagina pubblica, pannello admin, Telnet e marchio
- [dxregia/deploy/README.md](dxregia/deploy/README.md) — Riferimento dell'installatore Linux
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — Architettura
- [TASK.md](TASK.md) — Piano di lavoro
- [CHANGELOG.md](CHANGELOG.md) — Registro delle modifiche (in inglese)

**Struttura del repository:** `dxregia/` contiene il codice di DXRegia (backend, pannello, installatori), `engine/` il motore DXSpider Mojo con la sua struttura originale e `docs/` la documentazione.

## Stato

DXRegia è in sviluppo attivo (versione 0.14.1). Installazione, primo accesso, Telnet, backup e riavvio sono provati su Ubuntu 24.04 (amd64 e arm64). Restano aperti: la prova su Proxmox LXC, la consegna della posta con altri provider e un pacchetto di release firmato. Segnalazioni e idee sono benvenute nelle [Issue di GitHub](https://github.com/wifi75/dxregia/issues).

## Origine e licenza

DXRegia è un progetto indipendente basato su DXSpider Mojo di Dirk Koopman G1TLH e dei suoi contributori (il [mirror EA3CV](https://github.com/EA3CV/dxspider/tree/mojo)). Il repository conserva i file originali [COPYING](COPYING) e [LICENSE](LICENSE) (Artistic License 2.0). Le spedizioni DX annunciate vengono da [NG3K ADXO](https://www.ng3k.com/Misc/adxo.html); le bandiere da [flag-icons](https://github.com/lipis/flag-icons) (MIT).

[Licenze e avvisi di terze parti](docs/LICENSING.md)
