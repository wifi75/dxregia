# Piano di lavoro DXRegia

## Prossimo passo

Collaudare l'immagine e lo script locale su un host Linux con Docker, poi completare la distribuzione da release e la gestione dei nodi partner. La procedura è in dxregia/deploy/README.md.

## Stato

- [x] Analizzare le [lacune operative di Mojo](docs/PRODUCT-ANALYSIS.md) e confrontarle con progetti esistenti.
- [x] Aggiungere un controllo del prompt Telnet per il banco di prova.

- [x] Clonare il ramo mojo e aprire il branch locale dxregia/main.
- [x] Impostare struttura, architettura e passaggio di consegne.
- [x] Definire [pagina pubblica degli spot e area sysop](docs/INTERFACES.md).
- [x] Autenticazione sysop con cambio obbligatorio della password temporanea e protezione del wizard lato API.
- [x] Implementare la prima pagina spot pubblica con dati Mojo, filtri e layout mobile.
- [ ] Completare la mappa delle entità DXCC e collaudare il feed con due nodi collegati su Linux.
- [ ] Riprodurre l'avvio di Mojo su Linux e registrare dipendenze, configurazione minima e accesso locale alla console.
- [ ] Verificare la baseline con due nodi locali e una sessione Telnet. Annotare gli scambi attesi e i criteri per spot, riconnessione e persistenza.
- [x] Implementare la diagnostica locale Go/Svelte sul listener Telnet.
- [ ] Verificare la diagnostica web contro Mojo reale su Linux.
- [x] Collegare il modulo web di prima configurazione alla generazione dei file Mojo in una copia nuova.
- [x] Provare avvio gestito di Mojo dopo il salvataggio e verificare la porta Telnet scelta su Windows.
- [x] Verificare il riavvio gestito con la porta Telnet e la banca utenti persistenti su Windows.
- [x] Proporre la porta Telnet configurata e rifiutare una porta occupata prima di salvare.
- [ ] Realizzare il wizard di primo avvio e la gestione di un nodo partner interamente dal browser, dalla macchina appena installata al listener Telnet operativo, senza modifica manuale di file.
- [ ] Verificare accesso al terminale classico via Telnet da un client esterno e da un logger. Integrare console sysop web con cronologia e aiuto, utenti e spot in diretta.
- [ ] Distribuire OCI amd64/arm64 e realizzare lo [script di installazione Linux](dxregia/deploy/README.md) con URL del pannello, password temporanea e cambio obbligatorio al primo accesso. Verificare avvio, riavvio e riesecuzione su Linux.
- [ ] Chiarire l'applicabilità di COPYING e LICENSE e censire le dipendenze prima della prima distribuzione.

- [ ] Implementare monitoraggio dei nodi partner, registro eventi per categoria, revisioni della configurazione e backup con ripristino verificato. Vedi [Operazioni del sysop](docs/OPERATIONS.md).

## Criterio per la prima versione operativa

Da Linux, il sysop configura il nodo via browser, collega un partner Mojo, riceve uno spot sul logger via Telnet e ritrova configurazione e collegamento dopo il riavvio.

- [x] Anagrafica nodi partner con referente, email e storico note datate.
- [x] Registrare automaticamente la prima connessione confermata per i partner presenti nell'anagrafica.

- [x] Proteggere configurazione, anagrafica e presenza con TOTP e blocco del riutilizzo.
- [x] Mostrare sulla dashboard sysop nodi e sessioni da una fotografia recente di Mojo. Test reale Telnet su Windows.
- [ ] Collaudare nella dashboard i collegamenti tra due nodi e le sessioni Web Mojo.
- [x] Contatore pubblico e storico privato delle visualizzazioni giornaliere e dei domini di provenienza.
- [ ] Integrare visitatori attivi stimati e paesi di provenienza nell'analytics sysop.
- [ ] Integrare recupero accesso sysop e QR per l'iscrizione TOTP.

## Prima release

- [ ] Dopo il completamento e il collaudo del pacchetto, aggiornare il changelog e pubblicare la release 1.0.0 con tag v1.0.0. Gli aggiornamenti ordinari proseguono con 1.0.1, 1.0.2 e così via.

- [x] Ampliare il wizard con dati sysop, rinvio e importazione iniziale dei campi supportati Mojo.
- [ ] Consentire modifica dei dati sysop dopo l'avvio, cambio porta HTTP dal pannello e migrazione completa dei dati Mojo.
