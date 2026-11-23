# Passaggio di consegne

## Ripresa del lavoro

- Le schermate in `docs/screenshots/` mostrano `/spots` e `/admin` dalla build web locale su Edge. Per aggiornare le immagini dei README, usare un’istanza di prova con dati dimostrativi.
- Repository condiviso: https://github.com/wifi75/dxregia, ramo main. Su un altro computer partire dal ramo main del repository e leggere TASK.md, questo file e AGENTS.md.
- Punto raggiunto (2026-11-23, versione 0.15.4): pannello completo con i dieci punti del piano (TASK.md), motore DXSpider Mojo build 931 in `engine/`, collaudo GitHub verde su Linux (nativo e Docker, amd64 e arm64), server di produzione in aggiornamento dal pannello. Il motore sopravvive al riavvio del solo pannello e l'hook `local/Local.pm` si aggiorna da un marcatore (docs/DECISIONS.md).
- Per riprodurre, compilare il pannello e avviare il backend con una copia Mojo separata e Perl con le dipendenze richieste. I comandi e la baseline sono in dxregia/backend/README.md. Le istanze locali, i dati di prova e i codici temporanei non fanno parte del repository.
- Prossimo intervento: verifiche sul server dopo l'aggiornamento (elenco in TASK.md, «Da verificare»), poi le proposte di docs/PROPOSTE.md. Rilascio: ramo di lavoro, collaudo GitHub (`gh workflow run linux-smoke.yml --ref <ramo>`), unione su main, tag, pacchetti costruiti sul server e `gh release create`. Dopo ogni release che cambia `presence.pm`, controllare che `Local.pm` sul server sia cambiato.
- Interfacce: solo tema chiaro, riferimento per il portale pubblico in docs/INTERFACES.md. Usare la skill design-taste-frontend quando disponibile. Dati della pagina pubblica e gestione dei partner ancora da integrare.

## 2026-11-17 — Avvio del progetto

- Directory: D:/progetti/dxregia.
- Branch di lavoro: dxregia/main.
- Baseline upstream: mojo, commit 30e57466864c369787542b786a867a0216d85f31 (DXSpider 1.57 build 931, allineata il 2026-11-23, con due commit applicati sopra la base 5e4139b5 build 929; la base precedente era 9b55f03f, build 688).
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

Il percorso install-native.sh affianca Docker. Collaudo GitHub Actions 44272109256 passato su Ubuntu 24.04 con porte personalizzate, installazione da sorgente, setup 2FA, Telnet, backup, riavvio e riesecuzione senza sostituire il nodo. Il collaudo 44272892467 ha verificato anche l'installazione da archivio nativo precompilato su amd64 e arm64. Restano la pubblicazione degli archivi di release e la prova su LXC Proxmox.

L'analytics sysop mostra le visualizzazioni per pagina e una stima delle schede attive del portale. Una scheda visibile invia un segnale ogni 30 secondi e scade dopo 90 secondi dall'ultimo segnale. Gli identificativi casuali delle schede restano in memoria e il contatore pubblico mostra solo il totale delle visualizzazioni.

Il test Linux controlla anche il contatore pubblico, l'accesso riservato agli analytics, il conteggio per pagina e dominio, i segnali duplicati della stessa scheda e la persistenza delle visualizzazioni dopo il riavvio.

Il workflow Linux carica gli archivi nativi e i checksum come artefatti GitHub Actions al termine del collaudo dell'architettura corrispondente. Gli artefatti di prova durano 14 giorni e permettono il collaudo su server e LXC prima della release.

La matrice Docker ora prova due nodi Mojo isolati, ciascuno configurato tramite le API sysop con TOTP. Il test verifica abilitazione, collegamento, prima connessione e riconnessione. La propagazione di uno spot a un logger resta da verificare: la validazione upstream esclude gli spot PC61 con sorgente privata o loopback.

Lo script install-release.sh seleziona amd64/arm64, scarica i due asset nativi da GitHub Releases, verifica SHA-256 e delega l'installazione a install-native.sh. Il download reale resta da collaudare con una release che esponga gli asset.

Il collaudo dell'installer da release usa asset locali su ciascuna architettura: l'archivio valido deve avviare l'installer, un checksum errato deve fermarlo prima dell'estrazione. La prova contro gli URL reali resta aperta.

Il run Linux 44274300277 ha verificato su amd64 e arm64 provisioning, stato su entrambi i nodi, data di prima connessione e riconnessione dopo il riavvio del partner. La versione successiva del test controlla anche che una sessione Telnet compaia nella fotografia sysop e sparisca dopo la chiusura.

L'installatore nativo accetta DXREGIA_TELNET_BIND=0.0.0.0 prima del primo avvio per consentire logger e partner esterni. Il pannello resta sul loopback. I listener dei nodi gia configurati restano da modificare mediante migrazione coordinata.

Le notifiche SMTP supportano TLS implicito e STARTTLS obbligatorio tramite DXREGIA_SMTP_TLS. Entrambe le modalita verificano il certificato prima dell'invio. Resta da collaudare la consegna con un provider reale.

L'ascolto Telnet nativo su tutte le interfacce ha superato il controllo del listener su Ubuntu amd64 e arm64 nel run 44275996268. I job Docker di quel run si sono fermati al timeout del test spot. Il test successivo attende la finestra PC92 di Mojo per i nuovi utenti.

Il run 44276565981 ha ricevuto lo spot su Telnet in entrambe le architetture, ma il feed del nodo mittente non lo mostrava. DXCommandmode inoltrava gli spot ritardati senza salvarli sul nodo di origine. Ora usa Spot::add_local prima dell'inoltro, come il percorso immediato. Il test continua a richiedere la presenza in entrambi i feed.

L'installatore Docker accetta DXREGIA_TELNET_BIND come quello nativo. La matrice verifica l'endpoint host pubblicato, su loopback per amd64 e su tutte le interfacce IPv4 per arm64. Le riesecuzioni richiedono le impostazioni del container esistente.

Gli ambienti installati su Linux includono Python e lo strumento di backup verificabile: dxregia-backup nel runtime nativo e backup_node.py nell'immagine Docker. Il collaudo esegue backup, verifica e ripristino su copia usando lo strumento distribuito con ciascuna installazione. Rimane aperto il coordinamento dal pannello web.

Il run 44277402955 ha consegnato lo spot al logger e ai feed di entrambi i nodi su amd64 e arm64. Il test si e fermato al successivo riavvio del partner. I due container ora usano una rete Docker dedicata e il nome DNS stabile del partner per il collaudo della riconnessione.

Il collaudo del backup distribuito comprende anche un archivio ZIP valido con un file alterato e un checksum del manifest non corrispondente. Il ripristino deve rifiutarlo senza pubblicare una directory parziale.

Il pannello sysop ora verifica un archivio ZIP caricato fino a 512 MiB con lo stesso strumento del backup offline. Il file temporaneo viene eliminato a fine richiesta. Il nuovo collaudo controlla accesso protetto da TOTP, archivio valido e archivio alterato. Creazione, pianificazione e ripristino dal pannello restano aperti.

Il run 44278251212 ha riprodotto il timeout di riconnessione anche con nomi DNS stabili. Il prossimo collaudo raccoglie stato del partner, salute del listener e messaggi di connessione filtrati, senza stampare lo script di login o la password.

Il test del rifiuto senza autenticazione invia ora un corpo minimo: il backend risponde prima di leggere il corpo, e l'invio di un archivio intero poteva chiudere la connessione prima che il client leggesse il 401. I due casi autenticati usano ancora archivi completi.

La scheda del partner con riconnessione automatica mostra l'ultimo tentativo e l'orario UTC a partire dal quale e consentito riprovare, se l'endpoint e invariato e Mojo conferma uno stato gestibile. Il collaudo di riconnessione registrera gli stessi dati in caso di errore.

Il run 44279134901 mostra che il TCP della riconnessione raggiunge il partner e il suo listener risponde; la fotografia Mojo del nodo primario non si aggiorna. La diagnosi ora raccoglie anche listener primario, mtime della fotografia e stato del processo. Il test di verifica ZIP corretto e passato sui due job nativi del run 44279401543.

Il feed aggiunge il nome dell'entita DXCC ricavato dalla stessa tabella dei prefissi di Mojo, caricata all'avvio. Il collaudo di propagazione verifica il nome Japan per JA1ABC. Le bandiere e le traduzioni gia presenti rimangono disponibili per le rispettive entita.

Nel run 44280015487 entrambi i listener rispondono e la fotografia locale continua a essere scritta, ma il backend non la espone. L'hook ora serializza esplicitamente indicativo e stato come stringhe e orari come interi, anche quando gli scalari Perl hanno cambiato tipo durante la sessione. Il nuovo collaudo stampera la fotografia sintetica se la riconnessione fallisce ancora.

Per l'assistenza 2FA il backend distingue i codici già usati da quelli che non corrispondono alla chiave e all'ora corrente. Accetta gli spazi inseriti dall'app quando si copia un codice. La finestra di tolleranza e il blocco del riutilizzo restano attivi. Su Proxmox LXC controllare l'orologio sul nodo host.

Il run 44295449257 ha passato l'intera matrice Linux nativa e Docker amd64/arm64 con la modifica alla diagnosi TOTP. Il commit e3591bab e stato promosso su main. La procedura per l'orologio host Proxmox e il recupero 2FA e nella guida di installazione.

Aggiornamento nativo da qualsiasi directory tramite update-native.sh. Compila prima di arrestare il servizio, conserva configurazioni e credenziali, sostituisce runtime e parti upstream, controlla HTTP e Telnet e ripristina i componenti precedenti se il controllo fallisce. Il workflow lo esercita da una directory esterna al checkout e verifica che il contatore persistente resti invariato.

L'installatore nativo distribuisce lo script come /usr/local/bin/dxregia-update. Il collaudo lo invoca da fuori checkout. Sui server esistenti lo stesso file .sh puo essere eseguito direttamente o installato una volta nel PATH.

Aggiornamento Docker tramite update-docker.sh sull'host che esegue il container standard. Conserva il volume e riusa mapping di porte, politica di riavvio e variabili DXRegia. Il collaudo esegue l'aggiornamento da una directory esterna e verifica contatore e porte.

Navigazione sysop con menu e sottomenu a sinistra per Nodo, Collegamenti e Gestione. A destra si monta la sola pagina attiva, con link hash per cronologia e accesso diretto. La preselezione da una richiesta apre la pagina partner. Portale spot a larghezza piena, pulsanti condivisi con stati di clic, focus e movimento ridotto. Build Svelte verificata; il collaudo grafico sul server resta al sysop.

L'aggiornamento nativo indica commit installato e disponibile, supporta --check e salta riavvio e compilazione se coincidono. Il runtime conserva il commit all'aggiornamento e lo recupera dai metadati Go delle versioni precedenti quando disponibili. Cinque fasi, indicatori verdi e rossi su terminale, controllo HTTP con impostazioni di rete locali e senza avvisi per il file env opzionale assente.

La sezione Nodo → Accesso web consente di cambiare la porta HTTP nelle installazioni native gestite. Apre il nuovo listener prima di sostituire il precedente, salva la porta in /var/lib/dxregia/web-port e indica il nuovo URL. Installazione e aggiornamento usano la porta persistente per il controllo di salute. Le impostazioni di pubblicazione Docker rimangono gestite dall'host.

Il listener pubblico Telnet Mojo e standard su 0.0.0.0 per ogni nodo gestito, senza opzione di bind. All'avvio viene normalizzato Listeners.pm, con copia della configurazione precedente. L'installatore e l'aggiornamento Docker pubblicano Telnet su tutte le interfacce IPv4. Pannello HTTP e listener interno Mojo restano separati.

Il portale Spot usa una colonna sinistra con filtri per banda e modo e un registro a destra con ricerca, conteggio, orario dell'ultimo aggiornamento e controllo pausa. Su schermi stretti i filtri passano sopra il registro. Il sysop mantiene menu categorizzato a sinistra e area di lavoro a destra. Anteprima desktop renderizzata in Edge a 1440 px con spot di esempio.

Nodo → Console Telnet nel pannello sysop. Usa il listener locale configurato, richiede 2FA, consente interrogazioni di diagnostica con una lista di comandi e conserva la cronologia della vista. Ogni richiesta apre una sessione breve con limite di tempo e output. Il portale aggiunge superfici traslucide e segnale live animato con rispetto di prefers-reduced-motion.

Gestione → Password aggiunge il cambio separato di password web e sysop Telnet, entrambi con un nuovo TOTP oltre alla sessione autenticata. Il cambio web chiude la sessione; quello Telnet accoda una richiesta privata temporanea al motore, esclusa dai backup, e attende conferma. La console puo rispondere anche alla richiesta di password Telnet.

Nodo → Opzioni Telnet legge da Mojo password richiesta, nominativi con barra, registrazione e limiti di sessione. Le modifiche sono persistenti e il pannello attende riscontro dalla fotografia del motore. Le personalizzazioni senza hook gestito non sono editabili da questo menu. Le altre famiglie di opzioni Mojo rimangono da censire e integrare.

Scelta grafica sysop: Vetro chiaro. Portale e amministrazione condividono sfondo azzurro e menta, pannelli traslucidi, ombre e movimento riducibile. Menu sysop a sinistra e pagina selezionata a destra. Build Svelte e anteprime Edge desktop del feed e dello stato sysop con dati di esempio.

Il cambio della password sysop Telnet genera bcrypt prima di scrivere la richiesta al motore. Mojo conserva il valore bcrypt e verifica il login Telnet senza salvare il testo originale. Il collaudo nativo richiede un nuovo TOTP, attende la conferma e accede con il nominativo sysop. Credenziali peer legacy, chiavi TOTP recuperabili e backup storici richiedono una gestione separata. Corretto anche il pattern browser del codice 2FA nella pagina Password.

Aggiunta nodo: Salva e connetti persiste il piano di riconnessione prima di registrare l'anagrafica. Il worker prepara il profilo Mojo per la sola uscita, poi accoda il connect anche dopo riavvio. Salva solo anagrafica resta disponibile per le richieste da valutare. Il primo tentativo non invia una password; se il partner la richiede, va inserita nella scheda, lasciando attiva la riconnessione. L'abilitazione in ingresso resta un'azione separata.

Lo script Mojo generato per il peering ora specifica in modo esplicito timeout, host e porta, nominativo del nodo locale con SSID, password se richiesta e riga client con indicativo remoto. Il collaudo Linux con due nodi usa un profilo locale per sola uscita e il login con password al nodo ricevente. Nessun accesso al nodo N2KI durante il collaudo.

L'elenco partner aggiorna ogni cinque secondi la colonna Stato con LED e testo, anche senza aprire una scheda. L'anteprima dello script mostra campo per campo la configurazione del collegamento con password oscurata. Nodo → Log DXSpider mostra eventi di connessione e avvio selezionati, usando solo messaggi ricostruiti dal backend e senza esporre log di login grezzi.

Il form Aggiungi nodo accetta la password remota quando si seleziona Con password. Salva e connetti la passa al piano di connessione persistente. La scheda amministrativa mostra la password salvata tramite API autenticata con no-store, separata dall'elenco partner e dal log. Il piano legacy conserva le credenziali remote in un file privato del nodo, poiché Mojo deve reinviarle alla riconnessione.

Il provisioning dal pannello applica anche registered(1), equivalente a set/register, oltre al tipo Spider e alla configurazione della password per l'ingresso. Per gli account esistenti il flusso verifica tipo e blocco prima di registrarli e preserva password non corrispondenti.

La password remota è ora opzionale al salvataggio di un partner: si può impostare o aggiornare anche senza connettere subito, e se il campo manca resta quella già salvata. Se il salvataggio dell'anagrafica fallisce, il piano di riconnessione precedente viene ripristinato. Disattivare la riconnessione imposta il flag Disabled sul piano invece di cancellarlo, così la password resta disponibile per una riattivazione. Una password salvata senza connessione crea un piano disabilitato. Limite: build e test Go non eseguiti su questo host perché go non è nel PATH; la modifica non è collaudata a runtime.

## 2026-11-18 — Release 0.1.0

Pubblicata la prima versione 0.1.0 (tag v0.1.0, CHANGELOG.md, dxregia/VERSION). Gate Go non eseguito: go assente su questo host. Prossimo passo: installare Go, eseguire go build e go test in dxregia/backend e proseguire i collaudi a due nodi.

## 2026-11-18 — Release 0.1.0 e collaudo del collegamento a N2KI

Pubblicata la 0.1.0 (tag v0.1.0, CHANGELOG.md, dxregia/VERSION). Il server di prova è un'installazione nativa su 192.168.1.124 (servizio dxregia, dati in /var/lib/dxregia/mojo), aggiornato con update-native.sh scaricato da main perché non ha ancora il comando dxregia-update. Ultima revisione installata c16a2dee.

Il collegamento in uscita verso N2KI (dxc.n2ki.com:7300) è attivo e scambia dati. Tre difetti trovati e corretti nel collaudo:
- Salvare l'anagrafica di un nodo esistente azzerava la password remota: ora il pannello invia la password solo se scelta «Con password» e la rimuove solo se si passa esplicitamente a «Senza password».
- L'account Mojo creato con `registered` ma senza password bloccava «Prepara il collegamento» con «Account Mojo esistente da verificare». presence.pm ora accetta un account tipo S non bloccato senza password per il solo collegamento in uscita e gli assegna il marcatore «solo uscita». Poiché Local.pm è una copia di presence.pm, l'hash della versione precedente è stato aggiunto all'elenco degli aggiornabili in presence.go: ogni modifica futura di presence.pm richiede di aggiungere l'hash della versione precedente.
- Il passo di accesso prima del 2FA e il recupero del secondo fattore durano 20 minuti invece di 5. La sessione completa resta di 8 ore, ma è in memoria: ogni riavvio del servizio richiede un nuovo login con 2FA. Salvarla su disco non è stato fatto.

Novità del pannello: la pagina Stato e sessioni è una dashboard con LED (listener, cluster, partner, spot), contatori, spot per banda e ultimi spot, costruita sulle API esistenti. La pagina Nodi partner ha due schede, «Il tuo nodo si collega a…» e «… si collega al tuo nodo», con «Abilita e invia l'email» e «Reinvia le istruzioni via email» (POST /api/partners/{call}/resend). La password in ingresso generata all'abilitazione è conservata in local_data/dxregia-partners/<call>.inbound (0600), quindi il reinvio funziona solo per i nodi abilitati da questa versione in poi. Aggiunto il pannello Server di posta (API /api/smtp-settings, vedi OPERATIONS.md). Le bandiere coprono tutte le entità DXCC con un paese: 252 SVG di flag-icons (MIT) in dxregia/web/public/flags e tabella DXCC→paese in web/src/dxccFlags.ts.

Verifiche: go vet e test del pacchetto sul server (Go non è installato sul PC di sviluppo; si copia backend/ sul server e si lancia go test con -skip TestSpotEntitiesFromBundledMojoData, che legge dati Mojo non presenti nella copia). Le pagine nuove compilano ma non sono state viste nel browser. La consegna email con un SMTP reale non è collaudata.

Prossimo passo: form pubblico Contatti per la richiesta di collegamento con i soli campi utili, configurazione automatica del partner, invio dei dati di collegamento e email di conferma alla prima connessione. Decisione aperta: abilitazione completamente automatica o con un click del sysop (l'automatica permette di registrare un indicativo non proprio con un'email qualsiasi). Da fare anche: correggere l'IP di docker-vps nella skill SSH (ora 192.168.1.124).

Versione 0.2.0 (tag v0.2.0): raccoglie dashboard, pagina Nodi partner, pannello Server di posta, bandiere DXCC e le correzioni del collaudo N2KI. Il server di prova 192.168.1.124 esegue c16a2dee, precedente al solo bump di versione.

## 2026-11-18 — Pagina spot, analytics, logo e documentazione in sei lingue

Stato: il server di prova 192.168.1.124 esegue 5708af81 (o successivo) e il repository è su main. Novità rispetto alla 0.2.0, tutte nel CHANGELOG alla voce «Unreleased».

Pagina pubblica: filtri a selezione multipla su una riga, ricerca con lente sui nominativi, nuovi spot gialli per 10 secondi con un breve suono (interruttore nell'intestazione), modo dedotto dalla frequenza quando il commento non lo indica (piano IARU regione 1; il bordo tratteggiato lo distingue dal modo dichiarato), tema scuro, tabella a schede su telefono, footer con visite e link al progetto. Le lingue sono sei (src/translations.ts, src/i18n.svelte.ts): il catalogo non si chiama `locales` perché il `.gitignore` upstream ignora `local*` e il file non finiva su GitHub. Le spedizioni DX annunciate arrivano dal feed pubblico NG3K ADXO (backend/cmd/dxregia/adxo.go, ogni 6 ore, `DXREGIA_ADXO=off` per spegnerle) con popup al passaggio del mouse. Il tema scuro inverte i colori dell'interfaccia e riporta all'aspetto originale bandiere, riquadri di banda e modo, LED e menu scuro (src/theme.css): ogni nuovo elemento colorato va aggiunto a quell'elenco.

Pannello: ora su `/admin`, non collegato dalle pagine pubbliche; `/` rimanda a `/spots`. L'accesso chiede utente (indicativo del sysop, `sysop` prima della configurazione), password e codice a due fattori. Menu laterale scuro con gruppi apribili, LED dei partner e contatore delle richieste non lette. Dashboard live: flusso SSE su /api/live che avvisa quando cambiano presenza o spot. Centro di analytics (analytics.go, traffic.go): i visitatori unici si contano con un'impronta solo in memoria, con sale che cambia a ogni riavvio, senza cookie né indirizzi salvati. Richieste e nodi partner si eliminano dal pannello; l'eliminazione di un nodo scrive un file `.remove` che presence.pm usa per scollegarlo e cancellarne l'account Mojo. Configurazione automatica dei partner (partner_auto.go): crea la scheda, abilita l'ingresso e manda le istruzioni; non tocca nodi o account già presenti.

Promemoria tecnici: presence.pm viene copiato in Local.pm e si aggiorna solo se l'hash della versione precedente è nell'elenco di presence.go; ogni modifica di presence.pm richiede di aggiungere l'hash della versione che sostituisce. Il PC di sviluppo non ha Go: si copia backend/ sul server e si lancia `go vet` e `go test -skip TestSpotEntitiesFromBundledMojoData ./cmd/dxregia/`. Gli aggiornamenti del server si fanno con update-native.sh scaricato da main; ogni aggiornamento riavvia il servizio e interrompe per circa un minuto i collegamenti, che si ristabiliscono da soli.

Logo: scritta DXREGIA disegnata a mano in stile «Blocchi pesanti», senza font, con X ambra (dxregia/web/public/brand). L'icona DX resta per favicon e icone del telefono. «Blocchi netti» (icona e Regia) è conservata solo in docs/brand/varianti. Dettagli in docs/INTERFACES.md.

Documentazione: README e guida di installazione (docs/INSTALLATION*.md) in inglese per prima e poi italiano, spagnolo, tedesco, francese e portoghese; i sei file di ciascun gruppo vanno tenuti allineati a mano. CHANGELOG in inglese. Operazioni, interfacce, architettura e riferimento dell'installatore restano in italiano.

Verifiche: go vet e test del pacchetto sul server; build Svelte; controlli a schermo nel browser su pagina spot, accesso, analytics (con dati finti) e README su GitHub. Non verificati: suono e giallo dal vivo, pagine pubbliche nelle lingue diverse dall'italiano, consegna dell'email di conferma al richiedente, collegamento in ingresso reale di un partner, esito della CI sugli ultimi commit.

Limiti: la configurazione automatica non verifica l'indirizzo email, quindi un indicativo non proprio può ottenere l'accesso con un'email qualsiasi; il modo dedotto è una stima; il pannello è solo in italiano.

Prossimi passi: decidere il bordo del modo dedotto; verificare l'email del richiedente prima dell'accesso automatico; tradurre il pannello; preparare la 0.3.0 con le novità; correggere l'IP di docker-vps nella skill SSH (192.168.1.124).

Versione 0.3.0 (tag v0.3.0, 2026-11-18): raccoglie tutto il lavoro della sezione precedente. Il logo definitivo è il disegno «Blocchi pesanti» scelto dall'utente (lettere molto spesse, X ambra, A appuntita): non sostituirlo con una versione più leggera. Il server di prova esegue la stessa revisione dopo l'aggiornamento.

## 2026-11-18 — Motore spostato in engine/

Il motore DXSpider Mojo (cmd, connect, data, filter, msg, perl e le cartelle upstream non usate a runtime) è ora in `engine/` con i percorsi interni invariati; nella radice restano dxregia, docs, engine, COPYING, LICENSE e i file del progetto. Il nodo installato non cambia: `/var/lib/dxregia/mojo` ha la stessa struttura di prima, perché gli installatori ne copiano solo sei cartelle più COPYING e LICENSE, ora prese da `engine/`. Verifiche: CI verde su installazione nativa e Docker, amd64 e arm64 (bundle, installazione, riavvio, backup, aggiornamento), test del backend compreso quello sui dati del motore, aggiornamento reale del server di prova dal nuovo schema (`DXREGIA_BRANCH=engine-reorg sh update-native.sh`). Il tag `pre-engine-move` conserva lo stato precedente. Per unire in futuro novità del motore originale: `git merge upstream/mojo` rileva gli spostamenti; se serve, `-X find-renames`.

Nodo → Messaggio e pianificazione scrive `motd` e `local_cmd/crontab` (API /api/motd e /api/crontab, solo sysop con TOTP). DXSpider valuta la colonna comando del crontab come Perl: il pannello accetta solo `run_cmd("testo semplice")` e conserva intatte le altre righe. Il crontab viene riletto da DXSpider quando cambia. Pannello admin: alla scadenza torna all'accesso; la durata si imposta in Password (local/dxregia-session.json, predefinita 8 ore). Non verificato nel browser né su un nodo avviato. Prossimo passo: startup, liste di blocco, hop e filtri.

Nodo → Script di avvio, Liste di blocco, Hop e filtri: API /api/startup, /api/badlists, /api/hops, /api/filters (solo sysop con TOTP), descritte in docs/NODE-SETTINGS.md. presence.pm applica gli hop (local/dxregia-hops.json) e ricarica le liste quando il file cambia: l'hash precedente di presence.pm è stato aggiunto in presence.go. Ogni campo ha un tooltip (Hint.svelte). Le liste nodx/noann/nowwv di DXProtVars non sono lette da questo motore e non sono esposte. Non verificato su un nodo avviato. Prossimo passo: traduzione del pannello admin.

Pannello admin tradotto: i testi restano scritti in italiano nei componenti e adminI18n.ts li sostituisce nel DOM (nodi di testo e title/placeholder/aria-label) cercandoli in adminTranslations.ts (riga = it, en, es, de, fr, pt; «{}» = valore variabile). Per i testi costruiti dal codice (confirm) si usa tr(). Un testo senza riga resta in italiano: `python scripts/missing_admin_strings.py` dalla cartella web/ elenca quelli mancanti. Date e numeri restano formattati it-IT. Verificato in browser su pagine con API simulate (tedesco); non su un nodo reale. Scheda partner: nuova riga sullo stato in ingresso (FirstInboundAt). Piè di pagina del portale: versione DXSpider da /api/engine.

## Banner Telnet prima del login

Il pannello legge il testo salvato dalla vecchia versione in `motd` se manca `issue`. Salvando una volta lo scrive in `local_data/issue`. Mojo lo legge a ogni nuovo collegamento prima del prompt `login:`. Salvare il campo vuoto disattiva il banner.
Verifica locale: test Go e build Vite riusciti. Prova in produzione: aggiornare, risalvare il banner dal pannello e aprire una nuova sessione PuTTY.

Registrazione utenti dal portale (backend user_registration.go, register_mail.go; frontend Register.svelte, RegisterConfirm.svelte, NodeRegistrations.svelte). Flusso: POST /api/register (nominativo nella forma DXSpider senza SSID, email, password 8–72 ASCII stampabili) → hash bcrypt e token (solo l'hash SHA-256 del token si conserva) in .dxregia-registrations.json → email con link PUBLIC_URL/registrazione/conferma?token=… → la pagina fa POST /api/register/confirm → il backend scrive local_data/dxregia-users/NOMINATIVO.register (solo l'hash) → presence.pm (provision_users) crea o aggiorna l'utente Mojo (sort U, priv 0, password {DXREGIA-BCRYPT}hash, registered 1) e scrive NOMINATIVO.result. La revoca scrive NOMINATIVO.revoke. Impostazioni in local/dxregia-register.json (Enabled, PublicURL): l'indirizzo pubblico è scelto dal sysop e non si ricava mai dalla richiesta, per evitare link verso siti altrui nelle email. Per far inviare spot solo ai registrati serve Opzioni Telnet → Registrazione «Solo utenti registrati dal sysop inviano spot» (reqreg=1): con reqreg=2 il flag registered di Mojo significa «de-registrato». Limiti: 40 richieste al giorno, 3 per nominativo/email. Il nodo di prova ha un utente di collaudo TE5TA senza password. Prossimo passo: provare l'intero percorso con una vera email.

Versione 0.4.0 (tag v0.4.0, 2026-11-18): impostazioni di DXSpider nel pannello (messaggio del giorno, crontab, script di avvio, liste di blocco, hop, filtri, opzioni Telnet con tooltip), pannello admin tradotto in sei lingue con durata della sessione e ritorno all'accesso alla scadenza, storico accessi e cruscotto con nodi partner e utenti separati, registrazione degli utenti dal portale con conferma via email, portale installabile come app e vista smartphone compatta con filtri a tendina, banner sponsor Aurora, versione di DXSpider (1.57 build 688) e autore nel piè di pagina. Il server di prova gira sulla 0.4.0 (aggiornato con /root/update-native.sh, che tiene le ultime tre versioni precedenti). Aperto: prova della registrazione con una vera email, localizzazione di date e numeri del pannello (oggi it-IT), creazione guidata dei filtri, bordo tratteggiato del modo dedotto, verifica email nel partner automatico.
- Nodi partner: il pulsante di invio rimane visibile con email presente anche se Mojo richiede una verifica, con il motivo accanto. Build Vite riuscita.
- Nodi partner: Aggiorna stato mostra il controllo in corso e il risultato accanto al pulsante. Build Vite riuscita.
- Updater nativo: log completo della build web in caso di errore o con DXREGIA_VERBOSE=1.
- Log DXSpider: i nuovi timeout mostrano la direzione e il nominativo del partner in uscita. Verificati il parser Go e la routine Perl isolata. La prova su un nodo Linux resta da fare.
- Il modulo partner distingue il nominativo remoto completo (SSID incluso se presente) dal nome descrittivo. Build Vite riuscita.
- Nodi partner: la spunta salva subito la riconnessione tramite PUT o DELETE. La disattivazione ferma i nuovi tentativi automatici; una richiesta gia accodata puo concludersi. Test Go e build Vite riusciti.
- Nodi partner: Mojo distingue il blocco di un nodo Spider e il sysop puo sbloccarlo con conferma, conservando la password locale. Test Go, build Vite e sintassi Perl verificati.
- Accessi al nodo: geolocalizzazione indicativa con DB-IP Lite locale, scaricato nelle installazioni e negli aggiornamenti nativi. Gli IP non vengono inviati a DB-IP. TE5TA su loopback e indicato come test. Riquadro Listener Telnet: recapito pubblico dal profilo, con esito riferito al controllo locale. Test Go, build Vite e sintassi shell verificati.
- Log DXSpider: eventi di preparazione Mojo e fasi Telnet, IP sui timeout storici correlabili, eventi di sessione nodo e ricerca. Verificati test Go, build Vite e routine di log Perl isolata.
- Blocchi partner: il nuovo snapshot legge lockout in Mojo e lo mostra subito in dashboard e nella lista. Il componente storico Local.pm effettivamente trovato sul server e incluso tra le versioni aggiornabili. Test Go, build Vite e sintassi Perl verificati; aggiornamento live da verificare.
- Collegamenti > Storico email partner: API riservata al sysop con 2FA. Archivia istruzioni, reinvii, avvisi e conferme dal primo invio dopo upgrade in local_data/dxregia-partner-mail, incluso nei backup. Salva il testo oscurato e lo stato SMTP, senza HTML originale o errori SMTP grezzi. Test Go e build Vite riusciti; invio SMTP reale da verificare sul server.
- Email del portale: il renderer comune aggiunge GitHub, il DX Cluster pubblico di IU3WJQ-7 e il credito a Tiziano Cassone (IU3WJQ). Icona PNG incorporata in multipart/related via CID. Test Go passati, incluso il parsing MIME. Resa nei client di posta reali da verificare.

- Portale: Statistiche apre /statistiche con conteggi dalle 00:00 UTC, attività per banda e ora filtrabile per continente DX e classifiche con bandiere. Stato del nodo mostra salute e sessioni. Analytics admin mantiene le visite.
- Accesso sysop: occhio e indicatore Bloc Maiusc. La sessione verificata si riprende con la stessa scadenza, dopo un primo accesso effettuato con la versione nuova.
- Gestione > Server di posta: linguette Email e SMTP. Impostazioni e stato degli avvisi in local_data. Il worker controlla i partner ogni 30 secondi quando Mojo invia lo snapshot e segnala una volta per interruzione, con ripristino opzionale.
- Verificati test Go, incluso riavvio della sessione e notifiche simulate, e build Vite. Restano la prova SMTP reale e il collaudo visivo sul server.
- Tabella spot desktop: DXCC 5 px a sinistra e testo di paesi e commenti al 94%. Build Vite passata.
- Login sysop: campi della stessa larghezza e footer con versione da /api/engine, repository e autore. Build Vite passata.
- Banner Aurora mobile: il link copre il banner, con la chiusura separata. Build Vite passata.
- Dashboard Sala operativa con tabella partner, storico email con ultimo esito dai job esistenti e guida inline per hop e filtri. Test Go e build Vite passati.
- Installatore nativo: bind IPv4 pubblico per configurazioni nuove, URL LAN e utente sysop stampati con la password temporanea. Guide in sei lingue aggiornate. Sintassi shell verificata; prova pulita su Debian/Ubuntu da fare.
- 2FA opzionale in Gestione > Password; le configurazioni 2FA esistenti restano attive. Cambio iniziale della password sempre obbligatorio. Login senza fattore aggiuntivo e sessione persistente verificati nei test Go; build web passata. Commento desktop 10 px a sinistra.
- Dashboard: stato e ora della sessione partner sulla stessa riga, con separatore. Build verificata.
- Messaggio e pianificazione: selettore con bandiera, banner predefinito in issue e traduzioni in issue.<lingua>. Il cambio di lingua al collegamento resta da implementare. Test Go e build web passati.
- Il bollino Nuova app di Aurora oscilla brevemente con una pausa tra i cicli. Build Vite passata.

## Registrazione e profilo Telnet

Il modulo raccoglie nome, località (QTH) e locator (QRA). Alla conferma email Mojo salva il profilo e assegna il nominativo del nodo locale come Home Node. Le richieste in coda nel vecchio formato con il solo hash restano compatibili.

## Richieste partner e conferma in ingresso

I README e le guide di installazione in sei lingue documentano la configurazione automatica dal modulo, l’accesso CC Cluster senza password per IP o DNS e la seconda email, che parte solo quando il partner avvia una sessione attiva. Lo storico registra l’esito SMTP. Il test Go verifica la distinzione tra ingresso e uscita e la build web è passata. Restano da provare sul server SMTP reale e il collegamento in ingresso di un CC Cluster remoto.

## Console Telnet e navigazione

La console resta montata nella sessione di amministrazione. Passare a Stato e sessioni lascia aperto il collegamento Telnet e conserva cronologia e output. Il riquadro Radioamatori collegati deve mostrare la sessione come Telnet dopo il successivo aggiornamento della presenza. La verifica su un nodo installato resta aperta.

La sessione CC Cluster senza password di N4ZKF-1 è stata osservata come attiva in ingresso. La conferma SMTP resta da verificare nello storico. Le diciture del pannello usano DXSpider e i messaggi di salvataggio Telnet indicano direttamente l’esito.

## Inventario licenze e distribuzione

`docs/LICENSING.md` registra la base e le modifiche di DXSpider, distingue i due testi di licenza upstream, documenta i componenti presenti nelle build e le verifiche ancora aperte. Gli avvisi Go e web sono inclusi nel bundle distribuito. Restano da formalizzare la licenza DXRegia separata, i termini dei dati radio e l’inventario dell’immagine finale.


## 2026-11-19 — Diagnostica e smoke test Linux

- Le metriche del nodo sono disponibili nell’API autenticata /api/node-health e nel riquadro Diagnostica del nodo. La conservazione è disattivata in partenza e si abilita con DXREGIA_EVENT_RETENTION_DAYS. Impostare DXREGIA_PUBLIC_STATUS=on per pubblicare /stato. Vedi [Operazioni](OPERATIONS.md).
- Il collaudo HTTP usa file sintetici, letture concorrenti del feed e 250 schede partner. Le prove esercitano il backend e la generazione della risposta. Il carico di 250 connessioni effettive tra nodi e l’installazione su Raspberry Pi fisico richiedono un ambiente dedicato.
- Il workflow Linux salta le modifiche ai soli documenti, annulla i run superati sullo stesso riferimento e verifica esplicitamente 2FA opzionale e persistenza della sessione. Tutte e quattro le varianti sono passate nel [run di verifica](https://github.com/wifi75/dxregia/actions/runs/44531010488).


## 2026-11-19 — Portale spot e ricerca DXCC

- Le schede **Spot in diretta** e **Ricerca e DXCC** usano il feed pubblico. Il browser salva fino a 20 filtri per profilo locale. Vista normale, compatta, per banda e confronto esportano solo gli spot mostrati. Nel filtro commenti le parole si combinano, le virgolette delimitano una frase, il segno meno la esclude e OR separa alternative.
- La spunta indica riscontro nello stesso insieme di spot da due nominativi distinti entro 3 minuti e 1 kHz. La novità DXCC si misura sugli spot ricevuti dal browser nella giornata UTC. Il locator nei commenti può riferirsi a una delle stazioni citate. Il locator della scheda DXCC indica il riferimento del catalogo.
- Il portale continua a campionare durante la pausa e recupera gli arrivi dall’archivio usando giorno UTC e posizione del record. Il server recupera fino a 1.000 eventi da un massimo di 7 giorni, con lettura limitata a 256 KiB per file. Il portale segnala il recupero parziale e mostra gli arrivi in ordine cronologico.
- Catalogo derivato dai dati del motore, con CQ e ITU del record, continente del DX e dello spotter per entità ed eccezioni per nominativo esatto da cty.dat. La cronologia cerca nei 7 giorni disponibili, fino a 8.000 record o 1 MiB per giorno. Il riepilogo bande usa tutte le righe indicizzate dell’entità. La tabella ritorna al massimo 200 risultati.
- Test Go per metadati, esclusione degli IP dagli spot pubblici, cronologia e ripresa tra giorni UTC. Prova della build del portale su Edge headless con feed simulato a larghezza ridotta.


## 2026-11-19 — Allarmi radio del portale

- La scheda **Allarmi radio** salva nel browser fino a 20 regole per nominativo, banda, modo, frequenza, nuova entità DXCC o attività portatile e mobile. Più criteri della stessa regola devono corrispondere insieme. L’avviso DXCC scatta una volta per entità e regola. Un filtro salvato e una regola possono scadere alla fine del giorno locale indicato.
- Il portale deve restare aperto e ricevere gli aggiornamenti del feed. La prima fotografia inizializza il monitoraggio. Le notifiche di sistema si attivano su richiesta e dipendono dal browser. Il registro continua anche con notifiche spente, in pausa nella vista spot o durante le ore di silenzio.
- Segnalazioni con lo stesso nominativo, banda e modo si raggruppano per 15 minuti. Il browser riceve al massimo una notifica al minuto. Il riepilogo conserva gli ultimi sette giorni con attività e il registro fino a 200 gruppi.
- Verifica Edge con feed e notifiche simulate nella build web, regola CW portatile a frequenza esatta, consenso esplicito, gruppi, riepilogo, silenzio e deduplicazione per entità. La consegna del sistema operativo su dispositivi reali resta da provare.


## 2026-11-19 — Release 0.5.0

- I pacchetti Linux amd64 e arm64 vengono dal workflow del commit di release, con SHA-256 verificabile e riferimento al sorgente nel tag v0.5.0. I README e le guide nelle sei lingue descrivono l’installatore.
- Il listener Telnet seleziona il banner prima del login. Il Paese dell’IP pubblico si ricava dall’indice locale generato dal backend da DB-IP. Per la LAN il nodo usa l’entità del proprio indicativo. Inglese è il ripiego quando manca la traduzione.
- Stato e sessioni misura disco e RAM ogni 30 secondi e conserva in memoria fino a un’ora di campioni. La ricerca unica del portale accetta dx:, de:, comment:, virgolette, esclusioni e OR. Le schede sono nel menu superiore.

- La barra avanzata degli spot riunisce filtri salvati, vista, azioni e le opzioni di conferma e DXCC. I menu a tendina condividono lo stile dei pulsanti. I controlli vanno a capo quando lo spazio è ridotto. Verificata la build e provati i controlli in Edge con larghezze di 504 e 1280 pixel.

- Il feed /api/spots riutilizza il JSON e l’ETag finché i file degli spot restano uguali. Nei benchmark locali Windows amd64 da 100 righe, la mediana di tre prove passa da circa 137 a 59 µs per le risposte complete. Le richieste condizionali passano da circa 71 a 62 µs. I test Go verificano il JSON aggiornato quando arriva uno spot, il 304 e 32 lettori simultanei.

## Filtri spot, recupero, bollettini e QSL

- **Hop e filtri** salva una regola spot nativa di DXSpider (posizioni 0–9, una regola di ogni azione per posizione). Banda e modo usano i segmenti del piano DXSpider, il continente indica quello del DX e il nominativo è esatto. Il motore rilegge il file entro il prossimo ciclo del modulo locale. Le regole della console si possono consultare ed eliminare e le regole create dal pannello si possono anche modificare.
- Il portale prova a recuperare gli spot dell’interruzione quando il feed ritorna. Mantiene l’ultimo cursore valido anche dopo un errore e deduplica le righe recuperate rispetto al feed. Il recupero usa l’archivio esistente, fino a 7 giorni e 1.000 righe, e segnala se è parziale.
- **Annunci** pubblica gli annunci generali e locali dei log DXSpider. Le comunicazioni indirizzate ai sysop, i messaggi privati e gli annunci respinti restano esclusi. **Propagazione** legge gli archivi WWV e WCY e mostra i due bollettini separatamente.
- Gli indicatori LoTW ed eQSL AG cercano il nominativo DX esatto negli elenchi pubblici ufficiali. DXRegia li aggiorna ogni 72 ore, riutilizza la copia locale se il download fallisce e riprova dopo 6 ore. Per disattivare il download e gli indicatori, impostare `DXREGIA_QSL_DIRECTORIES=off`. I badge indicano presenza nell’elenco del servizio. [ARRL LoTW](https://lotw.arrl.org/lotw-help/developer-query-lotw-users/?lang=en) · [eQSL AG](https://www.eqsl.cc/qslcard/Programming.cfm)
- Verificati gli endpoint Go, l’esclusione di annunci privati, l’aggiornamento del feed QSL e la persistenza delle regole esistenti. Provati in Edge le due schede, i badge e il filtro, la simulazione di interruzione e il modulo sysop. Il filtro nativo è stato caricato, applicato e rimosso con il modulo Perl di DXSpider in ambiente isolato.

## Mappa di propagazione e accesso dai logger

- **Propagazione** conta gli spot degli ultimi 30 minuti per continente dello spotter, continente DX e banda. Legge fino a 8.000 righe o 1 MiB per giorno coinvolto e aggiorna il riepilogo ogni 30 secondi. La mappa mostra le otto direzioni con più segnalazioni per la selezione corrente. La posizione dei marcatori è rappresentativa del continente. Le frecce indicano il verso della segnalazione. I bollettini WWV/WCY restano disponibili sotto la mappa.
- La registrazione e la conferma riuscita mostrano una guida Telnet con i valori pubblici configurati dal sysop, sette logger documentati e i link ai manuali ufficiali. I testi sono in italiano, inglese, spagnolo, tedesco, francese e portoghese.
- L’API pubblica degli annunci e la scheda **Annunci** sono state rimosse. Nella tabella **Spot in diretta** compare la colonna QSL con i badge accanto alle altre colonne. Sono stati rimossi anche il menu Servizio QSL e l’opzione Evidenzia nuove entità.
- Collaudi Go per aggregazione a cavallo della mezzanotte UTC, disponibilità dei bollettini e dati pubblici della registrazione. In Edge con dati di prova, verificati mappa, filtri, colonna QSL a 375 e 1280 px, lingue della guida e conferma.

## Radioamatori connessi nella dashboard

- In **Stato e sessioni** il riquadro dei radioamatori occupa tutta la larghezza sotto i nodi. L’elenco è ordinato per nominativo e paginato con 6 schede sulle larghezze ridotte e 12 alle larghezze maggiori. Ogni scheda conserva modalità e ora di collegamento. La bandiera usa il DXCC del nominativo ricavato da DXSpider e compare quando esiste nel catalogo.
- La presenza continua a leggere le istantanee precedenti prive di DXCC. La migrazione aggiorna il modulo locale distribuito nella versione precedente. Verificati il mapping DXCC con il modulo Prefix di DXSpider, la build e i test Go. Provate in Edge griglia, bandiere, cambio pagina e layout a 390 e 1440 px con 15 sessioni dimostrative.

## Condizioni solari e mappa MUF

- **Propagazione** apre con gli indici solari, le condizioni HF giorno/notte e le condizioni VHF di HamQSL. DXRegia aggiorna il feed XML al massimo una volta all’ora, mantiene l’ultima copia valida e mostra l’ora UTC del fornitore. `DXREGIA_HAMQSL=off` disattiva il prelievo. [HamQSL, N0NBH](https://www.hamqsl.com/solar.html)
- La mappa MUF a tutta larghezza usa il rendering originale di KC2G, caricato dal browser direttamente dalla fonte con crediti visibili. La mappa degli spot osservati e i bollettini WWV/WCY si aprono nelle due sezioni sotto la mappa MUF. [Descrizione e attribuzioni KC2G](https://prop.kc2g.com/about/)
- Verificati parser e API con il feed reale, build Svelte e test Go. In Edge provati i due prospetti, la mappa KC2G reale, la vista a 390 e 1440 px e le due sezioni espandibili.

## 2026-11-20 — Release 0.6.0

- Il tag v0.6.0 comprende la nuova pagina Propagazione, la griglia delle sessioni radioamatori, l’editor dei filtri spot e la guida Telnet per logger. Il changelog elenca le modifiche e le guide nelle sei lingue puntano alla nuova versione.
- Nella dashboard ogni radioamatore occupa una riga con bandiera, nominativo, tipo di accesso e orario. I nodi partner sono paginati a gruppi di otto. L’aggiornamento riconosce l’hash del Local.pm distribuito con v0.5.0 e installa l’hook con i dati DXCC.
- La procedura di pubblicazione usa i pacchetti Linux amd64 e arm64 del workflow Linux del commit di release, con i relativi checksum SHA-256.

## 2026-11-20 — Release 0.6.1

- La griglia dei radioamatori mostra quattro righe per pagina e dieci colonne su desktop ampio. Su schermi più stretti riduce le colonne in base allo spazio disponibile.

## 2026-11-20 — Release 0.6.2

- Il riquadro dei radioamatori conserva la bandiera e il tipo di accesso. I nominativi sono più grandi e il tema scuro usa testo bianco. L’orario resta disponibile negli altri riquadri delle sessioni.
- La pagina Propagazione mostra le schede quadrate e colorate di HamQSL, le previsioni HF e VHF e la mappa MUF di KC2G. I riquadri Bollettini e Propagazione osservata sono stati rimossi dalla vista pubblica.

## 2026-11-20 — Release 0.6.3

- Disco e RAM condividono un riquadro con grafici ad anello per la percentuale di utilizzo corrente. La diagnostica conserva tutte le misure con meno spazio verticale.

## 2026-11-20 — Release 0.6.4

- Disco, RAM e Diagnostica sono affiancati su desktop. Le torte di disco e RAM distinguono lo spazio usato e libero. Le bandiere DXCC mostrano il nome del Paese o del territorio nella lingua del pannello al passaggio del mouse.

## 2026-11-20 — Release 0.7.0

- L’RBN si attiva da Collegamenti → Reverse Beacon Network ed è disattivato inizialmente. Il motore collega il feed CW/RTTY e il portale legge fino a 100 rapporti degli ultimi 30 minuti. L’utente Telnet sceglie le modalità con SET/SKIMMER CW RTTY.
- I test coprono autorizzazione, attivazione e disattivazione, riconnessione cadenzata e inoltro dei candidati raggruppati in un ambiente simulato. Il collegamento al relay pubblico va verificato sul nodo al primo avvio del feed.

## 2026-11-20 — Account locali dei nodi partner

- La dashboard spiega che DXSpider può creare già bloccati gli account locali per i nodi appresi dalla rete. Il sysop verifica il blocco dalla scheda del partner prima di preparare l’uscita.
- La scelta di accesso senza password tramite IP o DNS è disponibile anche per i partner DXSpider preparati. Un account bloccato o senza credenziale locale per DXSpider deve essere preparato prima. L’invio dei soli dati pubblici è disponibile a parte.
- Verificati il salvataggio via API, il rifiuto di account non idonei e di indirizzi diversi da quello autorizzato, e il percorso del pulsante di invio con SMTP simulato in Chromium.
- La tabella Nodi partner e la dashboard mostrano prima i collegamenti attivi e ordinano i gruppi per indicativo. L’ordinamento si aggiorna insieme allo stato delle sessioni.

## 2026-11-20 — Paesi dei partner e degli spot RBN

- La colonna Paese dei partner mostra bandiera e nome in base al DXCC del nominativo completo, anche senza una sessione attiva. La tabella RBN affianca bandiera e nome nella colonna DXCC come la tabella Spot. La pagina spiega il ruolo dello skimmer e la lettura di SNR e Q.
- L’entità speciale ITU HQ (4U1ITU) è identificata con il simbolo ONU già presente negli asset e conserva il proprio nome DXCC.

## 2026-11-21 — DXRegia 0.7.1

Verifica 2FA ricordata per 10 giorni con cookie dispositivo fidato (trusted_device.go, test in web_session_test.go). Server di prova 192.168.1.124 aggiornato a cc7f1cdf. Da verificare a mano: login con codice, chiusura del browser e nuovo login con sola password.

## 2026-11-22 — DXRegia 0.9.0 (aggiornamento senza riavvio)

- Stato: `update-native.sh` confronta il commit installato con quello nuovo; se cambiano solo `dxregia/web/`, `docs/` e i file `.md` sostituisce le pagine del pannello senza fermare il servizio (Telnet e partner restano collegati) e ripristina le pagine precedenti se il controllo fallisce. Ogni altra modifica fa l'aggiornamento completo con riavvio. `DXREGIA_FULL=1` forza quello completo.
- Verifica: sintassi dello script e logica di confronto (fetch del commit vecchio, diff, classificazione) provate in locale; l'aggiornamento completo che installa lo script nuovo è riuscito sul server di prova. In corso: prova dell'aggiornamento leggero dal pulsante con questo commit di sola documentazione, controllando che il processo di DXSpider non si riavvii.
- Limiti: un cambiamento al programma Go richiede ancora il riavvio completo (Mojo è un processo figlio del servizio); separarli è il passo successivo, più delicato.

## 2026-11-22 — DXRegia 0.8.0

- Stato: rilasciata la 0.8.0 con l'aggiornamento dal pannello (senza richiesta di password, conferma in finestra) e la scheda nodi partner su due colonne con paese, bandiera, separatore, solo LED per i nodi attivi e IN/OUT. Guide di installazione nelle sei lingue aggiornate.
- Verifica: test Go e build web verdi; aggiornamento dal pulsante provato a mano dall'utente solo nella fase iniziale, esito finale da confermare.
- Nota di rilascio: a ogni release portare `version` di `dxregia/web/package.json` e `package-lock.json` alla versione (la barra laterale del pannello la mostra).
- Prossimo passo: confermare un aggiornamento reale dal pulsante; riavvio automatico di Mojo se si ferma a servizio attivo.

## 2026-11-21 — Aggiornamento dal pannello

- Stato: il pannello (Gestione > Aggiornamenti) mostra versione installata e disponibile (ultimo commit di `main`, confrontato con `/opt/dxregia/revision`) e avvia l'aggiornamento. Il backend (`update.go`) scrive soltanto il file vuoto `/var/lib/dxregia/update-request`; `dxregia-update.path` avvia `dxregia-update.service`, che esegue `/usr/local/bin/dxregia-update` (cioè `update-native.sh`) come root e scrive il registro in `/var/lib/dxregia/update.log`. Serve il sysop con verifica completata e la conferma nella finestra, senza richiedere di nuovo la password (scelta dell'utente); nessun parametro dell'utente arriva allo script.
- Verifica: `go test ./...` verde (test con GitHub e systemd simulati); sul server di prova la catena path → servizio → registro funziona ("già aggiornato"). Non provato: un aggiornamento reale avviato dal pulsante.
- Limiti: solo installazioni native Linux; le installazioni precedenti vanno aggiornate una volta da terminale per installare le unità; l'aggiornamento richiede go, npm, git sul server come lo script manuale.
- Prossimo passo: prova reale con una versione più nuova, poi rilascio. La prova è in corso dal pulsante del pannello con un commit di sola documentazione.

## 2026-11-21 — Riavvio del motore e scheda nodi partner (DXRegia 0.7.4)

- Stato: dopo il riavvio del container (23:11) il DNS non era pronto, Mojo è uscito con "Temporary failure in name resolution" e DXRegia non ha riprovato. Correzione in `mojo.go` (`retryMojo`, chiamata da `setup.go`) e in `dxregia.service` (`Wants=network-online.target`, `After=network-online.target nss-lookup.target`). La scheda "Nodi partner" di `Dashboard.svelte` ora ha due colonne e 16 nodi per pagina.
- Verifica: `go vet` e `go test ./cmd/dxregia` verdi, build web ed `svelte-check` senza errori nel file. Server di prova aggiornato a `3435295b`. Non provati: un riavvio reale con DNS in ritardo e la scheda a schermo.
- Limiti: i tentativi coprono solo l'avvio; se Mojo si ferma più tardi non viene riavviato.
- Prossimo passo: riavvio di prova del container, poi valutare il riavvio automatico di Mojo se si ferma a servizio attivo.

## 2026-11-21 — DXRegia 0.7.3

- Stato: sblocco ed eliminazione dei nodi partner usano una finestra di conferma dell'applicazione (`Partners.svelte`, funzione `askConfirm`) al posto di `confirm()` del browser.
- Verifica: build web ed `svelte-check` senza errori nel file; non provata a mano nel browser.
- Limiti: l'indicativo di un nodo salvato è `readonly` di proposito (chiave di anagrafica e motore); per correggerlo si elimina e si ricrea. Altre schermate potrebbero usare ancora `confirm()`.
- Prossimo passo: valutare un pulsante "Rinomina nodo" e cercare altri `confirm()`.

## 2026-11-21 — DXRegia 0.7.2

Corretto «Feed RBN non disponibile» dopo un aggiornamento: il browser teneva in cache la pagina con i vecchi nomi dei file. Ora le pagine pubbliche hanno Cache-Control no-cache e /assets è immutabile; main.ts ricarica una volta su vite:preloadError. Footer del portale in fondo alla pagina. Server di prova aggiornato.

## 2026-11-22 — Allineamento al motore upstream 1.57 build 929 (DXRegia 0.10.0)

Il motore in `engine/` è stato portato da 9b55f03f (build 688) a 5e4139b5 (build 929), 242 commit del fork EA3CV/dxspider che integra le correzioni di Dirk (G1TLH). Procedura usata: `git diff --binary` fra le due basi nel clone upstream, senza `dxweb/` e `dxweb-admin/` (il pannello DXRegia li sostituisce), applicato con `git apply --directory=engine`; i file conflittuali sono stati uniti a mano. Le nostre modifiche al motore sono in sette file (`ExtMsg.pm`, `DXCommandmode.pm`, `Verify.pm`, `Version.pm`, `cmd/set/password.pl`, `cmd/unset/password.pl`, `cmd/sysop.pl`): ogni sincronizzazione deve ricontrollarli.

Differenze di comportamento upstream da ricordare:

- I file predefiniti (`cty.dat`, `wpxloc.raw`, `badip.*`, `badnode`, `badspotter`, `badword`, `baddx`) stanno ora in `initial_data/` e upstream li copia in `data/` all'avvio da un percorso fisso `/spider`, che da noi non esiste. Li teniamo copiati anche in `engine/data/` (aggiunti con `git add -f`, perché `data/.gitignore` li esclude). A ogni sincronizzazione: `cp -up engine/initial_data/* engine/data/`.
- `is_callsign` rifiuta i nominativi che iniziano per Q: i nominativi di laboratorio dei collaudi sono ora T3TTT, T7TTT, T8TTT, T9TTT.
- I PC61 con sorgente non pubblica sono scartati: il collaudo Docker porta lo spot su una sottorete interna 192.31.196.0/24.
- Un secondo accesso con lo stesso nominativo (da loopback o con password) chiede «(Y/N)» prima di sostituire la sessione precedente: la console del pannello (`loginConsole`) risponde Y, come faceva di fatto il comportamento precedente.
- Nuovi controlli `badnode`/`badspotter` al login, limite di 12 connessioni utente per nominativo base, comandi `load/badnode` e `load/badspotter`, routing PC92 con rollback.
- `cty.dat` non contiene più `DX0NE`: il test Go usa `3D2CCC` come eccezione esatta.
- Non integrati: DXWeb, DXWeb Admin e il loro installatore. DXReg (registrazione utenti) è presente nel codice ma disattivato (`$reg_enable = 0` nel `DXVars.pm.issue` upstream).

Verifica: test Go e smoke Linux su GitHub (amd64 e arm64, nativo e Docker) verdi; motore avviato sul server di prova, su una copia dei dati reali e porte diverse, con login, `sh/dx` e spot RBN corretti. Non verificato: PC92 con partner reali per un periodo prolungato e Version.pm: il saluto Telnet mostra «DXSpider V build» vuoti perché senza `.git` cluster.pl non legge Version.pm (comportamento già presente prima dell'allineamento).

Aggiornamento del server di prova alla 0.10.0 dal pannello (aggiornamento completo, 2026-11-22 09:38 CET): revisione installata 4b452b13, servizio attivo, nuovo Mojo avviato, listener Telnet pronto, scambio PC18 con altri nodi in corso. Il registro update.log non è leggibile dall'utente sysop (permessi): lo stato dell'aggiornamento si legge dal pannello. Da controllare nel tempo: stabilità del PC92 con i partner diretti.

## 2026-11-22 — DXRegia 0.11.0 (voci 1, 2 e 3 del piano del pannello)

`mojo_routing.go` riconosce frasi note del registro di debug del motore (Bad Spot, PC61 da indirizzo non pubblico, PC92_MALFORMED, PC92_ADMISSION, rollback di PC92 A/C/D, troppi errori, accessi troppo ravvicinati, badip/badnode/badspotter, troppe connessioni, secondo accesso respinto) e le mostra come eventi «Routing» e «Accessi» senza IP degli utenti né contenuto dei pacchetti. Le frasi arrivano da `cluster.pl` e `DXProtHandle.pm` del motore build 929: a ogni allineamento upstream vanno riviste (il test `TestMojoRoutingAndAccessEvents` ne fissa il formato). Alcune righe PC92 sono scritte solo con certe categorie di debug (routelow, chanerr): se non compaiono nel log, non sono attive. `mojo_watch.go` riavvia Mojo dopo un'uscita con codice diverso da zero, al massimo 5 volte in 10 minuti. Prova pratica con `kill -9` sul motore reale ancora da fare.

## 2026-11-23 — DXRegia 0.15.5 (pagina Aggiornamenti)

L'aggiornatore confronta la revisione installata con l'ultimo commit di `main` (`githubHead`), non con l'ultima release. Dopo la 0.15.4 un commit di sola documentazione compariva come «Aggiornamento disponibile» con l'hash grezzo. Ora `updatePreview.DocsOnly` (tutti i file in `docs/`, `*.md`, `dxregia/deploy/README.md`, meno di 300) e, se il commit non ha tag, `info()` imposta `UpToDate`. Un commit con tag si propone sempre; una modifica che non è solo documentazione anche. La riga «Disponibile» compare solo se `!UpToDate`. Test: `TestUntaggedDocsCommitIsNotAnUpdate`, `TestDocsOnlyPreview`. Conseguenza: per essere proposta, una versione va sempre taggata; i commit di documentazione dopo la release non fanno più comparire aggiornamenti.

## 2026-11-23 — DXRegia 0.15.4 (paese sotto il nominativo)

`Dashboard.svelte`: nella griglia dei radioamatori ogni casella ha bandiera e, accanto, nominativo con il nome del paese sotto (`.session-text`, `small` con puntini di sospensione); tolto il suggerimento al passaggio del mouse (`.flag-name`). Altezza minima della casella 40 px (42 px su schermi larghi). Verifiche: `npm run build`, layout provato in una pagina di prova nel browser (nomi lunghi accorciati). Attenzione: `{@const}` va subito dentro `{#each}`, altrimenti il build fallisce.

## 2026-11-23 — DXRegia 0.15.3 (aggiornamento dell'hook Local.pm)

Nella stessa versione: `Dashboard.svelte` ordina le sessioni per `Since` decrescente e dà la classe `fresh` (12 s) a chi si è appena collegato, con anello verde animato (`@property --spin`, conic-gradient e maschera) e lampeggio dello sfondo; con `prefers-reduced-motion` resta un contorno fisso. L'animazione è stata provata in una pagina di prova nel browser, non nella dashboard con accesso.

`installPresenceHook` (`presence.go`) sostituisce `local/Local.pm` solo se la sua SHA-256 è nell'elenco delle versioni gestite; l'elenco non era stato esteso dopo la versione del 20/11 (`bfc99c7e`, impronta `19e1e360`), quindi il server aveva ancora quella: mancavano il ricaricamento di `badnode.local`/`badspotter.local` (0.12.0) e `SpotAt` (0.15.0), senza alcun errore (scoperto controllando `SpotAt` vuoto per tutti i nodi dopo ore). Ora: aggiunte le impronte di `19e1e360`, `1488b1d8` (0.12.0) e `9fc731b1` (0.15.x) e un marcatore `# DXREGIA-MANAGED-HOOK` sulla seconda riga di `presence.pm`; un file che lo contiene si sostituisce sempre. Da qui in poi non serve più estendere l'elenco. Test: `presence_hook_test.go` (marcatore presente, installazione, aggiornamento, file del sysop intatto). Lezione: dopo ogni release che cambia `presence.pm`, controllare sul server che `Local.pm` sia cambiato.

## 2026-11-23 — DXRegia 0.15.2 (motore build 931 e nuovo formato delle liste)

Il repository non contiene il commit upstream `5e4139b5` (l'allineamento alla build 929 fu una copia dei file): i due commit `9756930d` (DXUser, recupero di `dxusers.db` da `user_json`) e `30e57466` (validazione delle liste bad* e IPv6 CIDR) sono stati scaricati con `gh api -H "Accept: application/vnd.github.v3.diff"` e applicati con `git apply --directory=engine`; `Version.pm` è a build 931, git id `30e57466[i]`. Cambio che pesa: `DXHash.pm` ora tratta ogni file con estensione (`badnode.local`, `badspotter.local`) come testo semplice e, all'avvio (`_badlist_run`), converte una volta l'hash vecchio in testo con copia in `local_data/badlist-backups/`. Il pannello (`node_lists.go`) leggeva e scriveva solo l'hash, quindi dopo l'aggiornamento avrebbe mostrato le liste vuote e le avrebbe sovrascritte: `parseBadDX` ora legge entrambi i formati e `renderBadText` salva testo (commenti `#` conservati, nessuna data). `baddx` e i file di base senza estensione restano in formato hash. Verifiche: go vet, go test (nuovi `TestBadListPlainTextAndHash`, `TestRenderBadTextKeepsCommentsAndDropsHash`), build del pannello, sintassi Perl. Da provare sul server: conversione reale delle liste, avvio del motore con la nuova `DXHash.pm`, comandi `set/badnode` e `unset/badnode` da Telnet.

## 2026-11-23 — DXRegia 0.15.1 (pagina in cima a ogni cambio)

`Auth.svelte`: `$effect` che a ogni cambio di `page` chiama `window.scrollTo` in cima (menu, tasto indietro, `location.hash`). Verifica: `npm run build`; non provato in un browser con accesso. Solo pannello, quindi l'aggiornamento non riavvia il motore.

## 2026-11-23 — DXRegia 0.15.0 (schede della dashboard)

`presence.pm`: in `init` si avvolge `DXProt::handle_11` (anche PC26 e PC61 vi passano con `goto`) e, se la chiamata non è ricorsiva e il canale è un nodo, si salva l'ora in `%node_spot_at`; lo snapshot di presenza scrive `SpotAt` per i nodi (`presenceChannel.SpotAt` in `presence.go`). `Dashboard.svelte`: `silentPartners` = nodi in stato `normal`, non bloccati, con `max(SpotAt, Since)` più vecchio di 30 minuti; «Ultimi eventi» legge `/api/mojo-log` e mostra gli ultimi 4 eventi `Avvio` e `Collegamenti` (accessi e routing restano fuori, per i login ripetuti). «Nodi partner»: tre colonne, 18 per pagina, paese accanto al nominativo. Verifiche: sintassi Perl, `go vet`, test Go, build del pannello. Da provare sul server: `SpotAt` in `/api/presence` dopo qualche minuto; il primo aggiornamento riavvia il motore perché `presence.pm` cambia.

## 2026-11-22 — DXRegia 0.14.2 (dashboard con una sola scheda dei nodi)

`Dashboard.svelte`: tolta la scheda «Nodi collegati in ingresso» con `inboundNodes` e gli stili `.inbound`; `.columns` resta a colonna unica. `partnerRows` è l'unione dei partner registrati e di tutti i nodi collegati, quindi gli ingressi non registrati restano visibili con il tag IN. Verifica: `npm run build`; pagina non aperta in un browser con accesso (la modifica è solo una rimozione). Da guardare dopo l'aggiornamento: la dashboard con almeno un nodo in ingresso.

## 2026-11-22 — DXRegia 0.14.1 (pulizia dell'avvio e rete di sicurezza)

`engine/perl/cluster.pl`: `cp` da `/spider/initial_data` solo se la cartella esiste; `git describe` e `git branch` solo se `$root/.git` esiste (righe datate 2026-11-22, in attesa della verifica di licenza). `dxregia.service`: `ExecStopPost` ferma perl se il servizio termina con `SERVICE_RESULT=success` e non c'è un `keep-engine` più recente di 10 minuti; dopo un crash non interviene, perché `Restart=on-failure` fa adottare il motore dal nuovo servizio. Verifiche: collaudo GitHub completo (amd64 e arm64, nativo e Docker), che prova il riavvio con e senza `keep-engine`. Da provare sul server: aggiornamento da 0.14.0 (cambia `engine/`, quindi riavvio completo e registro di avvio senza le due righe).

## 2026-11-22 — DXRegia 0.14.0 (voce 10: motore separato dal servizio web)

Linux: `mojo_detach_unix.go` avvia il motore con `sh -c` (script `engineScript`) in un gruppo di processi proprio; la shell scrive `local_data/dxregia-engine.pid` (pid di perl) e `dxregia-engine.exit` (codice di uscita). `startMojo` prima prova `adoptRunningEngine`: pid file valido, `/proc/<pid>/cmdline` contiene `cluster.pl` e Telnet pronto; se sì il motore si adotta e un poller (ogni 2 s) segnala l'uscita, anomala se il codice non è 0. All'arresto del servizio il motore si ferma solo se manca `/var/lib/dxregia/keep-engine` (file recente, entro 10 minuti), creato da `update-native.sh` quando `git diff` tra revisione installata e nuova non tocca `engine/`, `COPYING`, `LICENSE`, `dxregia/backend/cmd/dxregia/presence.pm` (l'hook è incorporato nel binario e caricato dal motore all'avvio). In quel caso lo script salta anche lo scambio delle cartelle del motore. `dxregia.service` ha `KillMode=process`. SIGTERM va solo a perl: se va alla shell, la shell muore subito e il servizio esce prima che Mojo tolga il file di blocco (fallito il backup a freddo del collaudo). Windows e Docker restano come prima (motore figlio, si ferma col contenitore). Verifiche: go vet su linux e windows, test Go sul server Linux, collaudo GitHub completo con la prova `engine_survives_web_restart` in `smoke_linux.py`. Limiti: il primo aggiornamento che installa questa versione riavvia ancora il motore; un motore orfano dopo un crash del servizio si adotta al riavvio, ma uno `systemctl stop` successivo a un crash non lo ferma.

## 2026-11-22 — DXRegia 0.13.0 (voci 7, 8 e 9 del piano del pannello)

Rinomina: `partner_rename.go`, `POST /api/partners/{call}/rename`; ammessa solo se `Provisioning` è vuoto, il nodo non è collegato e in `local_data/dxregia-partners/` esiste al più `<call>.reconnect.json` (che viene spostato). Il campo `RenameBlocked` della scheda dice perché non si può e nasconde il modulo. Anteprima: `update_preview.go`, confronto GitHub `compare/<installata>...<disponibile>` fatto in `refresh`, con la stessa regola di `update-native.sh` per «solo grafica e documentazione» (nel dubbio, o con 300 file, risulta riavvio). `ConfirmDialog.svelte` sostituisce `window.confirm` (restano `askConfirm` di Partners e il `prompt` del browser per l'installazione dell'app). Saluto Telnet: già presente in `Messages` (`l2`: versione e build in inglese, italiano, tedesco), nessuna modifica. Verifiche: go vet, go test, npm build. Da provare sul nodo: rinomina di un partner appena registrato, anteprima con una versione nuova. Prossimo passo: voce 10, che richiede un piano per le unità systemd e per `update-native.sh` prima di scrivere codice.

## 2026-11-22 — DXRegia 0.12.0 (voci 4, 5 e 6 del piano del pannello)

Riepilogo giornaliero: `daily_summary.go`, stato in `local_data/dxregia-summary.json`, scelta in `DailySummary` delle preferenze email; usa gli stessi conteggi del log (`parseMojoRouting`). Nodi e utenti bloccati: `badnode.local` e `badspotter.local` in `local_data` (formato DXHash, nome dell'elenco uguale al file), unite dal motore all'elenco di base; l'hook `presence.pm` li rilegge cambiando data o dimensione del file. Backup: `backups.go` esegue `backup_node.py backup --online` e salva in `/var/lib/dxregia/backups` (0700, accanto alla cartella del nodo), con pianificazione in `local/dxregia-backup.json`; la banca utenti si copia con l'API di backup di SQLite e si controlla con `PRAGMA integrity_check`; i file `-journal`, `-wal` e `-shm` non entrano nell'archivio. Lo smoke test Linux prova un backup a caldo, la verifica e il ripristino. Da verificare su un nodo reale: durata del backup a caldo con molto traffico e comportamento della banca utenti sotto scrittura.
