# Piano di lavoro DXRegia

## Stato

- [x] Analizzare le [lacune operative di Mojo](docs/PRODUCT-ANALYSIS.md) e confrontarle con progetti esistenti.
- [x] Aggiungere un controllo del prompt Telnet per il banco di prova.

- [x] Clonare il ramo mojo e aprire il branch locale dxregia/main.
- [x] Impostare struttura, architettura e passaggio di consegne.
- [x] Definire [pagina pubblica degli spot e area sysop](docs/INTERFACES.md).
- [ ] Implementare pagina spot pubblica con feed, filtri e layout mobile.
- [ ] Riprodurre l'avvio di Mojo su Linux e registrare dipendenze, configurazione minima e accesso locale alla console.
- [ ] Verificare la baseline con due nodi locali e una sessione Telnet. Annotare gli scambi attesi e i criteri per spot, riconnessione e persistenza.
- [x] Implementare la diagnostica locale Go/Svelte sul listener Telnet.
- [ ] Verificare la diagnostica web contro Mojo reale su Linux.
- [x] Collegare il modulo web di prima configurazione alla generazione dei file Mojo in una copia nuova.
- [x] Provare avvio gestito di Mojo dopo il salvataggio e verificare la porta Telnet scelta su Windows.
- [x] Verificare il riavvio gestito con la porta Telnet e la banca utenti persistenti su Windows.
- [ ] Realizzare il wizard di primo avvio e la gestione di un nodo partner interamente dal browser, dalla macchina appena installata al listener Telnet operativo, senza modifica manuale di file.
- [ ] Verificare accesso al terminale classico via Telnet da un client esterno e da un logger. Integrare console sysop web con cronologia e aiuto, utenti e spot in diretta.
- [ ] Distribuire OCI amd64/arm64 e realizzare lo [script di installazione Linux](dxregia/deploy/README.md) con URL del pannello, password temporanea e cambio obbligatorio al primo accesso. Verificare avvio, riavvio e riesecuzione su Linux.
- [ ] Chiarire l'applicabilità di COPYING e LICENSE e censire le dipendenze prima della prima distribuzione.

- [ ] Implementare monitoraggio dei nodi partner, registro eventi per categoria, revisioni della configurazione e backup con ripristino verificato. Vedi [Operazioni del sysop](docs/OPERATIONS.md).

## Criterio per la prima versione operativa

Da Linux, il sysop configura il nodo via browser, collega un partner Mojo, riceve uno spot sul logger via Telnet e ritrova configurazione e collegamento dopo il riavvio.
