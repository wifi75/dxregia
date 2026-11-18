# Passaggio di consegne

## Ripresa del lavoro

- Repository condiviso: https://github.com/wifi75/dxregia, ramo main. Su un altro computer partire dal ramo main del repository e leggere TASK.md, questo file e AGENTS.md. Claude Code importa AGENTS.md tramite CLAUDE.md.
- Punto raggiunto: il wizard locale salva, inizializza e avvia Mojo. Diagnostica e riavvio con porta e utenti persistenti sono stati collaudati con Mojo reale su Windows. Una porta occupata viene rifiutata prima di scrivere.
- Per riprodurre, compilare il pannello e avviare il backend con una copia Mojo separata e Perl con le dipendenze richieste. I comandi e la baseline sono in dxregia/backend/README.md. Le istanze locali, i dati di prova e i codici temporanei non fanno parte del repository.
- Prossimo intervento: build del container e collaudo su Linux pulito. Il cambio obbligatorio della password sysop è implementato e verificato via API su Windows. Istruzioni della build locale in dxregia/deploy/README.md.
- Interfacce: solo tema chiaro, riferimento per il portale pubblico in docs/INTERFACES.md. Usare la skill design-taste-frontend quando disponibile. Dati della pagina pubblica e gestione dei partner ancora da integrare.

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

## 2026-11-17 — Mojo reale su Windows

Mojo 1.57 build 687 avviato in una copia di laboratorio separata, con Strawberry Perl portatile 5.42.3.1 e moduli Mojolicious, Data::Structure::Util, Math::Round e Digest::SHA1. Net::CIDR::Lite resta assente; la protezione CIDR è opzionale e il collaudo è ristretto al loopback. Il listener 127.0.0.1:7300 ha superato prompt, login con nominativo di prova e show/dx su archivio vuoto. L'API web sulla 8080 riporta ready. Prossimo passo: riprodurre su Linux con due nodi di prova e verificare scambio spot e riconnessione.

## 2026-11-17 — Collegamento fra due nodi di prova

Su Windows, Mojo Q1QQQ-2 (Telnet 127.0.0.1:7300) e Q3QQQ-2 (Telnet 127.0.0.1:7373) hanno completato il collegamento DXSP/PC92. Verificati links, disconnessione e riconnessione manuale. Uno spot di prova è arrivato come PC61 al ricevente ed è stato scartato dal controllo upstream di loopback/RFC1918. Nessun nodo pubblico coinvolto. La consegna dello spot al client e il collaudo Linux restano aperti.

Requisito confermato: il pannello si avvia prima di Mojo e tutta la prima configurazione si completa dal browser. La generazione dei file interni spetta al backend.

Direzione visiva aggiornata: solo tema chiaro per entrambe le aree. La diagnostica locale è stata ricompilata e il server HTTP serve la nuova build. Le schermate pubbliche e sysop complete restano da implementare.

Portale pubblico: direzione grafica approvata su anteprima statica con dati dimostrativi. Requisiti visivi in docs/INTERFACES.md. Feed reale e pannello sysop ancora da integrare.

## Prima configurazione via browser

Il modulo Svelte e l'endpoint Go generano la configurazione Mojo in una copia nuova. Build web e Go riuscite. Verificati codice errato, porta fuori intervallo, salvataggio con porta 7474, sintassi Perl e rifiuto del secondo invio. L'anteprima funziona sul loopback con codice temporaneo. Prossimo passo: inizializzare la banca utenti e avviare automaticamente il motore, quindi verificare la porta scelta.

## Avvio di Mojo dopo la configurazione

Con -setup-perl il backend inizializza una banca utenti nuova e avvia Mojo con log dedicato. La risposta del wizard attende il prompt Telnet sulla porta scelta. Build Go e Svelte riuscite. Su Windows, test HTTP con Mojo reale passato dopo una prima porta occupata e un retry sugli stessi dati. Il backend ha poi riportato ready sulla porta effettiva. Restano gestione del riavvio, correzione della configurazione già scritta, autenticazione sysop definitiva, installatore e collaudo Linux.

## Ripartenza del nodo configurato

Lo stato DXRegia viene salvato insieme ai file Mojo. Al riavvio il backend riprende la porta e avvia il motore senza ricreare il database. Se fallisce, il modulo ripropone i dati salvati per un nuovo tentativo. Build Go e Svelte riuscite. Collaudo Windows con Mojo reale: configurazione via HTTP, login Telnet di prova, spegnimento del nodo, riavvio del backend, stessa porta pronta e utente presente in SQLite. Restano autenticazione sysop definitiva, installatore e collaudo Linux.

Il wizard propone la porta del listener impostata per l'anteprima. Verifica che la porta sia libera prima di generare i file di Mojo. Collaudo Windows: 7300 occupata rifiutata senza salvataggio. 7474 accettata, Mojo avviato e prompt Telnet confermato.

## 2026-11-18 — Accesso sysop e pacchetto Linux

Build Go e Svelte riuscite. Verificati blocco del wizard prima del cambio password, invalidazione della password temporanea, logout e accesso con password salvata dopo riavvio. Aggiunti Containerfile e script di installazione da sorgente per Linux con Docker. Verificare build ed esecuzione Docker su Linux. Funzioni restanti in TASK.md.

## 2026-11-18 — Anagrafica partner

Anagrafica aggiungibile e modificabile nella sezione Nodi partner. Il backend conserva note datate e data di registrazione in .dxregia-partners.json nella directory Mojo. Build web e Go riuscite. Prova API Windows su accesso, validazione, duplicati, modifica e persistenza dopo riavvio. Prossimo passo collegare le schede a Mojo e registrare la prima connessione confermata.

## 2026-11-18 — Dashboard e autenticazione a due fattori

TOTP richiesto dopo la password per proteggere configurazione, anagrafica e presenza. Verificati via API iscrizione, persistenza, codice valido e rifiuto del riutilizzo. La chiave rimane sul server in .dxregia-totp. L'iscrizione è manuale nell'app autenticatore. Recupero con codici di emergenza e QR restano da integrare.

Per le installazioni gestite, il backend installa un hook Local.pm solo se assente. La dashboard mostra dati recenti sulle sessioni Mojo ogni cinque secondi. Verificato con Mojo reale Windows che Q7QQQ compare dopo login Telnet e scompare dopo disconnessione. Web Mojo e stato di partner non ancora collaudati in questa dashboard. Anagrafica e dashboard non modificano i collegamenti. Restano prima connessione persistente, visitatori del portale, contatore pubblico e analytics privato.

## 2026-11-18 — Wizard sysop e import iniziale

Wizard in due passaggi con rinvio e ripresa, dati sysop estesi e riepilogo privato. Anteprima da DXVars.pm, Listeners.pm con un listener e JSON a campi riconosciuti. Verificata l'importazione con il modello Mojo originale e con campioni JSON e Listeners.pm. Build Go e web riuscite. Modifica dei dati dopo l'avvio, cambio porta HTTP dal pannello e migrazione completa di utenti, filtri e partner restano aperti.

## 2026-11-18 — Portale spot reale e prima connessione dei partner

Il portale /spots interroga /api/spots ogni cinque secondi e legge fino a 100 spot recenti dall'archivio Mojo degli ultimi due giorni UTC. Verificati via HTTP la pagina, la bandiera locale e un record scritto dal vero Mojo nel laboratorio Windows. L'API restituisce solo i campi pubblici. Il modo appare quando esplicito nel commento. La mappa DXCC iniziale è parziale; PNG provenienti da Flagcdn. Il backend registra la prima sessione partner in stato normal, osservata nella fotografia privata di Mojo, e la conserva al ricollegamento. Test Go passato. Restano connessione dal pannello, analytics, migrazione completa, backup e collaudo Linux.

## 2026-11-18 — Contatore e statistiche del portale

Contatore pubblico di visualizzazioni su /spots e storico giornaliero privato degli ultimi 90 giorni in Statistiche del portale. Il referrer viene ridotto al dominio, senza path o query. Test HTTP passed: solo caricamenti GET del portale incrementano il totale, ricariche del feed e richieste HEAD escluse; verifica privata via TOTP e persistenza al riavvio. Restano visitatori attivi stimati e paesi di provenienza.

## 2026-11-18 — Contatto e richiesta peering

Modulo pubblico e coda privata disponibili. Scheda partner precompilata e nota iniziale generate dal pannello. Connessione SMTP con TLS implicito configurabile dall'ambiente. Test HTTP locale verificato per validazione, persistenza, accesso con TOTP, richieste messaggio e peering. Consegna SMTP con un provider reale, abilitazione automatica e messaggi della verifica tecnica restano da collaudare o implementare.

## 2026-11-18 — Abilitazione partner in ingresso

Il sysop autenticato può accodare la creazione dell'account Spider in Mojo. L'hook crea l'account con password casuale, scrive l'esito e rimuove la richiesta. Collaudato con Mojo reale su Windows: coda privata, creazione, richiesta della password via Telnet e assenza della password nei log debug di ingresso. Non verificato handshake nodo-nodo completo dopo la creazione. L'hook aggiorna solo la versione DXRegia nota e lascia intatti gli hook personalizzati. Per usare la modifica dei log ExtMsg.pm, avviare Mojo dalla base aggiornata. Mancano invio automatico delle istruzioni e dell'esito, connessione in uscita e collaudo Linux.

## 2026-11-18 — Collegamento in uscita e notifiche partner\

Collaudato il collegamento in uscita via API e due Mojo reali su Windows, con autenticazione del partner e porta remota dedicata. Entrambi i nodi hanno riportato Attivo; password assente dai log di debug del laboratorio. Le email di istruzioni e verifica sono attivabili per singolo partner. Sequenza testata con invio simulato; SMTP reale e Linux ancora da verificare. La riconnessione automatica e gli errori specifici di negoziazione sono ancora aperti.

## Riconnessione automatica per partner

Opzione esplicita nel pannello, piano persistente e tentativi distanziati di due minuti. Test Go verificato per sospensione al cambio di porta, link attivo, accodamento e protezione da tentativi ravvicinati. Build Svelte riuscita. Prossimo collaudo: caduta e ripresa su due nodi Mojo, poi backup e installatore Linux.

Collaudo su Windows completato con due Mojo reali: disconnessione via comando del sysop, rilevazione del link inattivo e riconnessione automatica confermata dal motore. Nell'istanza di laboratorio è stato anticipato l'orario del prossimo tentativo; il vincolo reale di due minuti è verificato nel test Go. La prova di spegnimento forzato non ha interrotto il processo Mojo nel laboratorio e non costituisce una verifica di caduta dell'host.

## Backup offline e ripristino

Tool Python backup_node.py con backup, verifica integrità e ripristino in una directory nuova. Collaudo Windows su un nodo Mojo inizializzato: conservati database utenti e anagrafica, rifiutato archivio modificato e sovrascrittura, login Telnet dopo il ripristino. Gli script pendenti non vengono riprodotti dal backup. Il pannello può esportare le impostazioni del nodo in JSON reimportabile. Restano dal pannello la pianificazione, il backup completo e la gestione del ripristino.

## Collaudo Linux e recupero sysop

Il [runner Linux](https://github.com/wifi75/dxregia/actions/runs/44267305597) ha passato build immagine, installazione da sorgente, password temporanea, TOTP, wizard, login Telnet, backup con verifica e ripristino in directory nuova, e avvio sulla stessa porta dopo riavvio. Sono state aggiunte le dipendenze Perl mancanti e la gestione degli alias Mojo nel backup. Il ramo remoto di prova è dxregia-linux-validation. QR locale e recupero tramite codici monouso: verificati via API il reset, l'invalidazione dei vecchi codici e il blocco del pannello fino a conferma del nuovo TOTP.

## Migrazione offline di un nodo Mojo

Il comando migrate_mojo.py prepara una copia gestita da un nodo Mojo SQLite esistente, con un singolo listener. Verificati su Windows utenti e impostazioni conservati, database invariato, avvio della copia sotto il backend DXRegia e login del sysop via Telnet. Aggiunta la prova di migrazione allo smoke test Linux con matrice amd64 e arm64. Importazione completa da browser e migrazioni di formati più vecchi di SQLite restano aperte.

Checkpoint Git richiesti circa ogni 20 minuti di lavoro attivo. Il ramo dxregia-linux-validation riceve i commit da collaudare. Aggiornare main dopo aver verificato i risultati.

## Modifica del profilo sysop a nodo avviato

Configurazione sysop ora consente di modificare nome, email, QTH, locator, coordinate e host pubblico. Il JSON salvato viene letto dall'hook gestito da DXRegia e applicato alle impostazioni Mojo e ai record del nodo e del sysop. Le modifiche a indicativi e porta Telnet richiedono una migrazione coordinata e sono rifiutate da questa API. Gli hook personalizzati vengono conservati. Smoke test Linux esteso con modifica, verifica del saluto Telnet e ripartenza.

L'installatore Docker ora accetta porte host web e Telnet distinte tramite ambiente. Il wizard in container usa il listener pubblicato e i tentativi di cambiare la porta solo dal wizard vengono rifiutati. La matrice smoke usa valori personalizzati su arm64 ed esercita la riesecuzione dello script. Il cambio delle porte di un container esistente resta un'operazione di migrazione.

## Installazione nativa systemd

Il percorso install-native.sh affianca Docker. Collaudo GitHub Actions 44272109256 passato su Ubuntu 24.04 con porte personalizzate, installazione da sorgente, setup 2FA, Telnet, backup, riavvio e riesecuzione senza sostituire il nodo. L'installatore nativo da release precompilata e la prova su LXC Proxmox restano aperti.
