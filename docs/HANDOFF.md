# Passaggio di consegne

## 2026-11-17 — Avvio del progetto

- Directory: D:/progetti/dxregia.
- Branch di lavoro: dxregia/main.
- Baseline upstream: mojo, commit 9b55f03ff870d50a9d3e57090dc7408035a4a2b1.
- Documentazione e struttura iniziali create. Il motore e i file di licenza del clone sono conservati.
- Su questo host il controllo iniziale ha trovato Node, ma go, npm e docker non risultavano disponibili nel PATH. Nessuna verifica runtime Linux è stata eseguita.
- Requisiti visivi e flussi per pagina pubblica e area sysop: docs/INTERFACES.md.
- Installazione automatica, URL del pannello e cambio password al primo accesso specificati in dxregia/deploy/README.md. Implementazione e test Linux in sospeso.
- Primo passo operativo: predisporre un ambiente Linux riproducibile per avviare Mojo e verificare la console locale. I criteri di accettazione sono in TASK.md.

I requisiti di monitoraggio dei partner, log, backup e ripristino sono descritti in [Operazioni del sysop](OPERATIONS.md). Implementazione e prove operative in sospeso.

L'accesso esterno alla console classica via Telnet e i relativi test con un logger sono specificati in docs/INTERFACES.md.

## 2026-11-17 — Analisi delle priorità

Priorità e criteri di verifica in [Analisi prodotto](PRODUCT-ANALYSIS.md). Il controllo del prompt Telnet è in dxregia/deploy/check_telnet.py. Prossimo passo: eseguirlo contro Mojo nel banco Linux, quindi verificare login, show/dx e scambio di spot. Questo host non ha WSL o un runtime container disponibile.
Controllo verificato con socket locali per prompt frammentato, risposta estranea e porta chiusa. Prova su Mojo reale in sospeso.

Direzione visiva registrata in docs/INTERFACES.md. Le prime schermate di riferimento sono Spot pubblici e Nodi partner con un errore di connessione. La resa visiva resta da implementare e verificare.

La home GitHub e la guida docs/INSTALLATION.md descrivono il controllo Telnet disponibile e il flusso previsto per installatore e pannello. Il prossimo passo tecnico resta il collaudo con Mojo su Linux.

## 2026-11-17 — Diagnostica web locale

Build Go e Svelte riuscite. Server HTTP e tre stati (prompt ricevuto, prompt assente, porta chiusa) verificati con socket locali su Windows. Avvio documentato in dxregia/backend/README.md. Prossimo passo: prova con Mojo su Linux, poi adattatore per lo stato dei partner.
