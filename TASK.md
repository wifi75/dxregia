# Piano di lavoro DXRegia

## Stato

- [x] Clonare il ramo mojo e aprire il branch locale dxregia/main.
- [x] Impostare struttura, architettura e passaggio di consegne.
- [x] Definire [pagina pubblica degli spot e area sysop](docs/INTERFACES.md).
- [ ] Implementare pagina spot pubblica con feed, filtri e layout mobile.
- [ ] Riprodurre l'avvio di Mojo su Linux e registrare dipendenze, configurazione minima e accesso locale alla console.
- [ ] Verificare la baseline con due nodi locali e una sessione Telnet. Annotare gli scambi attesi e i criteri per spot, riconnessione e persistenza.
- [ ] Implementare backend Go e pannello Svelte/TypeScript con un primo flusso di sola lettura che mostri lo stato effettivo di DXSpider.
- [ ] Realizzare il wizard di primo avvio e la gestione di un nodo partner.
- [ ] Integrare console sysop, utenti e spot in diretta.
- [ ] Preparare distribuzione OCI amd64/arm64, con volumi persistenti, e provare installazione e riavvio su Linux.
- [ ] Chiarire l'applicabilità di COPYING e LICENSE e censire le dipendenze prima della prima distribuzione.

## Criterio per la prima versione operativa

Da Linux, il sysop configura il nodo via browser, collega un partner Mojo, riceve uno spot sul logger via Telnet e ritrova configurazione e collegamento dopo il riavvio.
