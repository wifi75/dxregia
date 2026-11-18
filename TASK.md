# Piano di lavoro DXRegia

## Prossimo passo

Completare la migrazione di un nodo Mojo esistente e la gestione delle impostazioni dopo il primo avvio. Estendere i collaudi Linux a due nodi, feed DX e aggiornamento da release. La procedura è in dxregia/deploy/README.md.

## Stato

- [x] Analizzare le [lacune operative di Mojo](docs/PRODUCT-ANALYSIS.md) e confrontarle con progetti esistenti.
- [x] Aggiungere un controllo del prompt Telnet per il banco di prova.

- [x] Clonare il ramo mojo e aprire il branch locale dxregia/main.
- [x] Impostare struttura, architettura e passaggio di consegne.
- [x] Definire [pagina pubblica degli spot e area sysop](docs/INTERFACES.md).
- [x] Autenticazione sysop con cambio obbligatorio della password temporanea e protezione del wizard lato API.
- [x] Implementare la prima pagina spot pubblica con dati Mojo, filtri e layout mobile.
- [ ] Completare la mappa delle entità DXCC.
- [x] Verificare lo stesso spot al client Telnet e ai feed dei due nodi Linux, su amd64 e arm64. Evidenza nei job Docker del run 44277402955, che si fermano dopo, alla riconnessione.
- [x] Riprodurre l'avvio di Mojo su Linux e registrare dipendenze, configurazione minima e accesso locale alla console.
- [ ] Verificare la baseline con due nodi locali e una sessione Telnet. Annotare gli scambi attesi e i criteri per spot, riconnessione e persistenza.
- [x] Implementare la diagnostica locale Go/Svelte sul listener Telnet.
- [x] Verificare la diagnostica web contro Mojo reale su Linux.
- [x] Collegare il modulo web di prima configurazione alla generazione dei file Mojo in una copia nuova.
- [x] Provare avvio gestito di Mojo dopo il salvataggio e verificare la porta Telnet scelta su Windows.
- [x] Verificare il riavvio gestito con la porta Telnet e la banca utenti persistenti su Windows.
- [x] Proporre la porta Telnet configurata e rifiutare una porta occupata prima di salvare.
- [ ] Realizzare il wizard di primo avvio e la gestione di un nodo partner interamente dal browser, dalla macchina appena installata al listener Telnet operativo, senza modifica manuale di file.
- [ ] Verificare accesso al terminale classico via Telnet da un client esterno e da un logger. Integrare console sysop web con cronologia e aiuto, utenti e spot in diretta.
- [x] Collaudare installazione da sorgente su Linux, backup offline, login Telnet e riavvio con volume persistente.
- [ ] Distribuire OCI amd64/arm64 e realizzare lo [script di installazione Linux](dxregia/deploy/README.md) con URL del pannello, password temporanea e cambio obbligatorio al primo accesso. Verificare avvio, riavvio e riesecuzione su Linux.
- [ ] Chiarire l'applicabilità di COPYING e LICENSE e censire le dipendenze prima della prima distribuzione.

- [ ] Implementare monitoraggio dei nodi partner, registro eventi per categoria, revisioni della configurazione e backup con ripristino verificato. Vedi [Operazioni del sysop](docs/OPERATIONS.md).

## Criterio per la prima versione operativa

Da Linux, il sysop configura il nodo via browser, collega un partner Mojo, riceve uno spot sul logger via Telnet e ritrova configurazione e collegamento dopo il riavvio.

- [x] Anagrafica nodi partner con referente, email e storico note datate.
- [x] Registrare automaticamente la prima connessione confermata per i partner presenti nell'anagrafica.

- [x] Proteggere configurazione, anagrafica e presenza con TOTP e blocco del riutilizzo.
- [x] Mostrare sulla dashboard sysop nodi e sessioni da una fotografia recente di Mojo. Test reale Telnet su Windows.
- [x] Collaudare nella dashboard i collegamenti tra due nodi su Linux, inclusa la riconnessione automatica.
- [ ] Collaudare nella dashboard le sessioni Web Mojo.
- [x] Contatore pubblico e storico privato delle visualizzazioni giornaliere e dei domini di provenienza.
- [x] Mostrare nell'analytics sysop una stima delle schede pubbliche attive e le visualizzazioni per pagina.
- [x] Centro di analytics del sysop: pubblico del portale (visitatori unici, lingue, dispositivi, orari) e attività del cluster (bande, modi, entità, nominativi). Resta da integrare il paese di provenienza dei visitatori.
- [x] Integrare recupero accesso sysop e QR per l'iscrizione TOTP.

## Prima release

- [ ] Dopo il completamento e il collaudo del pacchetto, aggiornare il changelog e pubblicare la release 1.0.0 con tag v1.0.0. Gli aggiornamenti ordinari proseguono con 1.0.1, 1.0.2 e così via.

- [x] Ampliare il wizard con dati sysop, rinvio e importazione iniziale dei campi supportati Mojo.
- [ ] Consentire modifica dei dati sysop dopo l'avvio, cambio porta HTTP dal pannello e migrazione completa dei dati Mojo.

- [x] Modulo pubblico contatto e peering, coda privata e scheda partner precompilata.
- [x] Pannello Server di posta con prova di invio (verificato solo il backend con test; consegna reale da provare).
- [x] Dashboard con LED, Nodi partner in due schede con reinvio email e bandiere DXCC complete (compilate, da verificare nel browser).
- [x] Form Contatti con configurazione automatica del partner (interruttore, limite di 10 al giorno), email HTML in sei lingue e conferma alla prima connessione. Provata a mano la notifica al sysop; la configurazione automatica completa con un nodo reale è da collaudare.
- [ ] Verificare l'indirizzo email del richiedente prima di creare l'accesso automatico (oggi chi invia il form può indicare un indicativo non suo).
- [x] Pagina spot pubblica rifatta (filtri multi-selezione, lente, modo dedotto, suono e giallo sui nuovi spot, tema scuro, layout telefono e tablet), spedizioni DX annunciate (NG3K ADXO), logo «DXREGIA a blocchi pesanti», favicon e documentazione in sei lingue. Verificati nel browser: layout, filtri, popup; non verificati dal vivo: suono, giallo sui nuovi spot e le pagine nelle lingue diverse da italiano.
- [x] Pannello su `/admin`, accesso con utente e password, dashboard live (SSE), eliminazione di richieste e nodi partner.
- [ ] Decidere se il modo dedotto dalla frequenza resta con il bordo tratteggiato, diventa continuo o si nasconde.
- [ ] Tradurre il pannello admin nelle sei lingue (oggi solo italiano) e i documenti tecnici oltre README e guida di installazione.
- [ ] Collaudare SMTP con un provider reale e completare approvazione, istruzioni al partner e verifica con esito via email.

- [x] Abilitare dal pannello un partner in ingresso con password e conferma di Mojo.
- [x] Accodare su scelta sysop istruzioni dopo la conferma dell'account e un esito dopo cinque minuti.
- [ ] Collaudare le due notifiche con SMTP reale.
- [x] Aggiungere riconnessione automatica per partner con limite di frequenza e vincolo a host e porta.
- [x] Collaudare la riconnessione automatica dopo disconnessione fra due nodi Mojo reali su Windows.

- [x] Collegare un partner in uscita dal pannello. Verifica Windows fra due istanze Mojo con password e porta remota dedicata.

- [x] Backup offline verificabile e ripristino in nuova directory. Test con avvio Mojo e login Telnet.
- [x] Download dal pannello delle impostazioni sysop e del nodo come JSON reimportabile.
- [x] Integrare la verifica di un archivio di backup dall'area sysop protetta da TOTP. Collaudo completato su Linux nativo amd64 e arm64 (run 44279401543).
- [ ] Integrare backup completo e pianificazione dal pannello con arresto e ripresa coordinati.
- [x] Migrazione offline di Mojo SQLite in una nuova directory gestita e login Telnet su copia.

- [x] Modificare i dati del profilo sysop a nodo avviato con applicazione nell'hook Mojo gestito.
- [ ] Completare la modifica di porte e indicativi con migrazione e verifica dell'endpoint.
- [x] Installazione nativa da sorgente con systemd, verificata su Ubuntu 24.04. Prova LXC Proxmox ancora aperta.

- [ ] Portare nel pannello le opzioni Mojo con inventario di variabili e comandi, includendo filtri, protocolli e funzioni a configurazione avanzata. Il primo menu Opzioni Telnet gestisce accesso, registrazione, sessioni e riconnessioni. Messaggio e pianificazione gestisce motd e local_cmd/crontab (solo run_cmd validato). Pagine Script di avvio, Liste di blocco e Hop e filtri (hop predefiniti e per PC, elenco e rimozione dei filtri; la creazione resta ai comandi accept/reject). Restano la creazione guidata dei filtri e la localizzazione di date e numeri del pannello (oggi sempre it-IT).
- [x] Pannello admin tradotto in sei lingue con selettore nell’intestazione; dizionario in dxregia/web/src/adminTranslations.ts.
- [x] Sessione sysop: ritorno alla pagina di accesso alla scadenza e durata configurabile (Password).

- [x] Salvare il banner Telnet in `local_data/issue` per mostrarlo prima di `login:`. Il pannello recupera il vecchio testo, che va risalvato una volta.
- [x] Registrazione utenti dal portale con conferma via email (modulo, link, hook Mojo, pagina sysop Utenti registrati). Verificato: hook Mojo su nodo reale (creazione, login con password, rifiuto della password errata, revoca); NON ancora provato l'invio dell'email e il clic sul link.
