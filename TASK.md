# Piano di lavoro DXRegia

[Proposte numerate per le prossime versioni](docs/PROPOSTE.md)

## Prossimo passo

Completare la migrazione di un nodo Mojo esistente e la gestione delle impostazioni dopo il primo avvio. Estendere i collaudi Linux a due nodi, feed DX e aggiornamento da release. La procedura è in dxregia/deploy/README.md.

## Stato

- [x] Riunire i comandi avanzati degli spot e le opzioni di conferma e DXCC in un’unica barra. Verificato su Edge a larghezza desktop e ridotta.

- [x] Preparare DXRegia 0.5.0 con guide in sei lingue, pacchetti nativi, grafici disco/RAM, ricerca unica, navigazione stabile e localizzazione del banner Telnet.

- [x] Integrare le proposte 21–30. Regole per gli spot e registro locale, riepilogo, ore di silenzio, scadenze e consenso per le notifiche. Flussi verificati su Edge con feed e consegna di sistema simulati.

- [x] Integrare nel portale le proposte 1–20 dopo le anteprime. Verificati catalogo e metadati DXCC, storico, recupero da archivio, CSV e interazioni del portale su Edge con dati simulati.

- [x] Completare le proposte 171–190 con misure sul feed, diagnostica del nodo, conservazione opzionale, guida rapida e modelli per le issue. Le prove HTTP usano 32 lettori del feed e una rubrica con 250 partner sintetici. Il collaudo Raspberry Pi fisico resta al punto 131.
- [x] Correggere lo smoke test Linux per il 2FA facoltativo e per le sessioni persistenti. La matrice nativa e Docker è passata sul ramo di verifica, su amd64 e arm64.

- [x] Analizzare le [lacune operative di Mojo](docs/PRODUCT-ANALYSIS.md) e confrontarle con progetti esistenti.
- [x] Aggiungere un controllo del prompt Telnet per il banco di prova.

- [x] Clonare il ramo mojo e aprire il branch locale dxregia/main.
- [x] Impostare struttura, architettura e passaggio di consegne.
- [x] Definire [pagina pubblica degli spot e area sysop](docs/INTERFACES.md).
- [x] Autenticazione sysop con cambio obbligatorio della password temporanea e protezione del wizard lato API.
- [x] Implementare la prima pagina spot pubblica con dati Mojo, filtri e layout mobile.
- [x] Usare il catalogo completo delle entità DXCC presente nei dati distribuiti con DXSpider.
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
- [ ] Collaudare il terminale Telnet classico da un client esterno e da un logger. Sul nodo installato, verificare anche la console sysop web con una sessione DXSpider reale.
- [x] Proposta 31 — Console Telnet web con sessione persistente, cronologia e scorrimento.
- [x] Proposta 32 — Riquadri separati per risposte ai comandi e spot in diretta.
- [x] Proposta 33 — Completamento dei comandi supportati con Tab, spiegazione degli argomenti ed esempi.
- [x] Collaudare installazione da sorgente su Linux, backup offline, login Telnet e riavvio con volume persistente.
- [ ] Distribuire OCI amd64/arm64 e realizzare lo [script di installazione Linux](dxregia/deploy/README.md) con URL del pannello, password temporanea e cambio obbligatorio al primo accesso. Verificare avvio, riavvio e riesecuzione su Linux.
- [x] Verificare la provenienza di COPYING e LICENSE, censire le dipendenze Go e web distribuite e includere gli avvisi nel pacchetto. Vedi [Licenze](docs/LICENSING.md).
- [ ] Formalizzare la licenza del codice DXRegia originale, verificare i termini dei dati AD1C e ADXO e controllare i pacchetti Debian e gli avvisi dell’immagine OCI finale.

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
- [ ] Completare le voci mancanti del pannello admin nelle sei lingue e tradurre i documenti tecnici oltre README e guida di installazione.
- [x] Configurare automaticamente i partner idonei dal modulo pubblico e inviare le istruzioni dopo l’abilitazione locale.
- [x] Configurare l’accesso CC Cluster per IP o nome DNS e confermare via email la prima sessione in ingresso rilevata.
- [x] Verificare un collegamento in ingresso senza password da un CC Cluster remoto (N4ZKF-1 attivo in ingresso).
- [ ] Collaudare sul server l’intero flusso dal modulo pubblico alle due email, incluso l’esito SMTP della conferma per N4ZKF-1.

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
- [x] Registrazione utenti dal portale con conferma via email (modulo, link, hook Mojo, pagina sysop Utenti registrati). Verificato: test automatico dell’intero percorso con email e Mojo simulati; hook Mojo su nodo reale (creazione, login con password, rifiuto della password errata, revoca); NON ancora provato l'invio dell'email e il clic sul link.
- [x] Mostrare il pulsante di invio delle istruzioni ai partner anche quando lo stato Mojo lo disabilita.
- [x] Dare riscontro visibile al controllo manuale dello stato del nodo partner.
- [x] Mostrare il log della build web su errore o in modalita dettagliata.
- [x] Distinguere nel log DXSpider i timeout Telnet in ingresso dai timeout dei partner in uscita.
- [x] Indicare nel modulo partner che il nominativo remoto include il suffisso SSID se presente.
- [x] Salvare la riconnessione dei partner al cambio della spunta, anche con il collegamento gia attivo.
- [x] Sbloccare dal pannello un singolo nodo partner confermato come bloccato da Mojo.
- [x] Mostrare il paese degli IP pubblici negli accessi e distinguere il controllo Telnet locale dal recapito pubblico.
- [x] Mostrare fasi, direzione, sessione e IP quando deducibili nel log DXSpider, con ricerca e filtro degli accessi anonimi.
- [x] Avvisare in dashboard se un partner e bloccato in Mojo e offrire lo sblocco con un pulsante rosso.
- [x] Aggiungere lo storico sysop delle email dei nodi partner, con contenuto oscurato e stato di invio.
- [x] Firmare le email del portale con icona DXRegia incorporata, link al repository, link al cluster pubblico e credito autore.

- [x] Aggiungere la visibilità della password e l’indicatore Bloc Maiusc nella schermata sysop.
- [x] Riprendere la sessione sysop già verificata dopo il riavvio del servizio.
- [x] Separare Email e SMTP e notificare disconnessione e ripristino dei nodi partner.
- [x] Aggiungere le bandiere al log DXSpider e riallineare DXCC e Commento nella tabella desktop.
- [x] Pubblicare le statistiche radio nella home del cluster e semplificare Stato del nodo.
- [x] Rifinire su desktop la posizione di DXCC e la dimensione del testo di paesi e commenti.
- [x] Allineare i campi di accesso e aggiungere i riferimenti nel footer del login.
- [x] Rendere cliccabile tutto il banner Aurora su smartphone.
- [x] Applicare Sala operativa, mostrare gli esiti email preesistenti e spiegare Hop e filtri nel pannello.
- [x] Installazione nativa nuova accessibile dalla LAN, con URL e credenziali iniziali documentati.
- [x] Rendere il 2FA facoltativo e disattivato per nuove installazioni; spostare Commento di altri 10 px su desktop.
- [x] Affiancare orario e stato dei partner nella dashboard.
- [x] Modificare e salvare i banner Telnet per lingua in una sola pagina.
- [x] Animare il bollino Aurora con supporto a movimento ridotto.

- [x] Semplificare la registrazione pubblica e impostare automaticamente il nodo di riferimento per il primo accesso Telnet.

- [x] Mantenere aperta la Console Telnet web passando da una pagina di amministrazione all’altra.
- [ ] Verificare sul server che la sessione web compaia in Stato e sessioni mentre la console è aperta.

- [x] Uniformare le diciture visibili del pannello a DXSpider e semplificare i messaggi delle Opzioni Telnet.
