# Piano di lavoro DXRegia

[Proposte numerate per le prossime versioni](docs/PROPOSTE.md)

## Prossimo passo

Completare la migrazione di un nodo Mojo esistente e la gestione delle impostazioni dopo il primo avvio. Estendere i collaudi Linux a due nodi, feed DX e aggiornamento da release. La procedura è in dxregia/deploy/README.md.

## Piano miglioramenti del pannello (2026-11-22)

Regola: nessuna voce di menu nuova se la funzione sta bene dentro una pagina esistente (schede nella pagina). Unica novità di menu: il gruppo **Diagnostica** (Salute del nodo, Log DXSpider), che toglie il log dal gruppo Nodo. Un lavoro per volta, ciascuno verificato e rilasciato.

| # | Lavoro | Dove nel menu |
|---|--------|---------------|
| 1 | Registro connessioni e PC92 (rotte scartate, rollback, accessi respinti, troppe connessioni), filtro per partner | Diagnostica > Log DXSpider, con schede |
| 2 | Salute del nodo: attivo da, memoria, disco, riavvii di Mojo, DNS | Diagnostica > Salute del nodo (nuova) |
| 3 | Riavvio automatico di Mojo se si ferma, con avviso e voce nel registro | Impostazione dentro Salute del nodo |
| 4 | Avvisi sui partner spenti oltre una soglia e riepilogo giornaliero | Collegamenti > Storico email partner (rinominata Avvisi ed email partner) |
| 5 | Liste di blocco modificabili con ricarica (`load/badnode`, `load/badspotter`) | Nodo > Liste di blocco (esistente) |
| 6 | Backup pianificato e ripristino dal pannello | Gestione > Backup (esistente) |
| 7 | Rinomina di un nodo partner | Scheda del partner, nessuna voce |
| 8 | Anteprima delle novità prima di aggiornare | Gestione > Aggiornamenti (esistente) |
| 9 | Sostituire gli ultimi avvisi nativi del browser; versione del motore nel saluto Telnet | Nessuna voce |
| 10 | Mojo separato dal servizio web (aggiornamenti senza riavvio del motore) | Nessuna voce, lavoro delicato, per ultimo |

## Stato

- [x] Rilasciare DXRegia 0.12.0 (voci 4, 5 e 6 del piano). Voce 4: gli avvisi sui partner spenti con soglia e ripristino esistevano già nelle preferenze email; si è aggiunto il riepilogo giornaliero. Voce 5: elenchi nodi e utenti bloccati al login. Voce 6: backup a caldo, pianificato, con scarico e conservazione. Restano le voci 7-10 del piano.
- [x] Rilasciare DXRegia 0.11.0: log Routing e Accessi, riavvio automatico del motore, durata di servizio e motore nella diagnostica. Cambio di piano: la «Salute del nodo» non è una pagina nuova, perché disco, memoria, processi e ultimo spot erano già nella scheda Diagnostica del nodo; ora vi si aggiungono durata e riavvii.
- [x] Rilasciare DXRegia 0.10.0 (motore allineato a DXSpider 1.57 build 929, risposta alla conferma di sostituzione della sessione nella console del pannello). Installata sul server di prova dal pannello il 2026-11-22, motore avviato e Telnet pronto.
- [x] Rilasciare DXRegia 0.9.0 (aggiornamento senza riavvio per le sole modifiche alla grafica, piè di pagina del pannello, nomi delle versioni nella pagina Aggiornamenti).
- [x] Rilasciare DXRegia 0.8.0 (aggiornamento dal pannello, scheda nodi partner con paese e bandiera).
- [x] Aggiornamento dal pannello (Gestione > Aggiornamenti): `update.go`, unità `dxregia-update.service` e `dxregia-update.path`, pagina `Updates.svelte`. Verificato: test Go, catena path unit → servizio → registro sul server di prova (risultato "già aggiornato"). Da provare a mano: pulsante "Aggiorna ora" con una versione realmente più nuova.
- [x] Rilasciare DXRegia 0.7.4 (riavvio del motore dopo un boot con DNS non pronto e scheda nodi partner su due colonne).
- [x] Riavvio del server: dopo un boot con DNS non pronto Mojo non partiva e restava spento. Ora l'unità systemd attende la rete e DXRegia riprova l'avvio con attese crescenti (10 s, 20 s, 40 s... fino a 2 min). Solo avvio: un arresto successivo di Mojo non viene ancora riavviato. Da provare con un vero riavvio del container.
- [x] Scheda Nodi partner dello stato del nodo su due colonne compatte, 16 nodi per pagina.
- [x] Preparare DXRegia 0.7.3: conferma di sblocco ed eliminazione dei nodi partner in una finestra dell'applicazione invece dell'avviso del browser. L'indicativo di un nodo salvato resta non modificabile (per correggerlo: elimina e ricrea).
- [x] Preparare DXRegia 0.7.2: pagine pubbliche sempre ricontrollate dal browser dopo un aggiornamento, ricarica automatica una volta se manca un file, footer del portale sempre in fondo alla pagina.

- [x] Ricordare la verifica in due passaggi per 10 giorni su ogni browser con un cookie di dispositivo fidato, anche dopo riavvio del browser e del servizio. Aggiungere 10 giorni alle durate di sessione.

- [x] Mostrare bandiera e nome nella colonna Paese dei nodi partner e nella colonna DXCC degli spot RBN, inclusa la sede ITU. Spiegare il ruolo dello skimmer nella pagina RBN.

- [x] Ordinare automaticamente la tabella Nodi partner e il riquadro della dashboard con i nodi attivi in cima e ordine alfabetico per indicativo dentro ciascun gruppo.

- [x] Offrire ai partner DXSpider già preparati e sbloccati la scelta di accesso in ingresso senza password tramite IP o DNS autorizzato e l’invio delle relative istruzioni.

- [x] Chiarire nella dashboard che gli account locali dei partner possono essere già bloccati da DXSpider quando il nodo viene appreso dalla rete.

- [x] Integrare Reverse Beacon Network per CW e RTTY. Il sysop può attivare il collegamento in uscita dal pannello e il portale mostra gli spot RBN in una scheda dedicata.

- [x] Preparare DXRegia 0.6.4 con Disco, RAM e Diagnostica sulla stessa riga su desktop, torte colorate e nome del Paese al passaggio del mouse sulle bandiere DXCC.

- [x] Preparare DXRegia 0.6.3 con i grafici ad anello di disco e memoria su una riga e la diagnostica del nodo più compatta.

- [x] Preparare DXRegia 0.6.2 con nominativi più grandi, testo bianco nel tema scuro e griglia dei radioamatori senza orari. Rendere quadrate e colorate le schede delle condizioni solari e rimuovere i riquadri Bollettini e Propagazione osservata dalla vista pubblica.

- [x] Preparare DXRegia 0.6.1 con dieci colonne e quattro righe per i radioamatori collegati su desktop.

- [x] Preparare DXRegia 0.6.0 con il changelog e l’installazione aggiornata nelle sei lingue. Compattare la dashboard, mostrare fino a 30 sessioni per pagina su desktop, paginare i nodi partner e ripristinare le bandiere DXCC dopo l’aggiornamento dalla 0.5.0.

- [x] Organizzare Propagazione con indici solari, condizioni HF giorno/notte e VHF di HamQSL e mappa MUF di KC2G a tutta larghezza.

- [x] Disporre le sessioni dei radioamatori in una griglia a tutta larghezza, con paginazione e bandiera DXCC del nominativo.

- [x] Mostrare la propagazione osservata per area e banda. Aggiungere alla registrazione le istruzioni Telnet per sette logger in sei lingue. Rendere più compatta la tabella con colonna QSL e rimuovere Annunci, il filtro QSL e l’evidenziazione delle nuove entità.

- [x] Integrare le quattro proposte dal confronto con CC User. Editor regole spot DXSpider, recupero automatico su riconnessione, annunci e propagazione WWV/WCY, indicatori e filtro LoTW/eQSL AG.

- [x] Mostrare il portale pubblico e la dashboard sysop nei README in sei lingue con schermate dimostrative della build web.

- [x] Riutilizzare il JSON del feed spot per ogni versione dell’archivio. Il benchmark locale da 100 righe misura circa 137 → 59 µs per risposta completa e 34 → 9 allocazioni per richiesta.

- [x] Riunire i comandi avanzati degli spot e le opzioni di conferma e DXCC in un’unica barra, con i menu a tendina nello stile dei pulsanti. Verificato su Edge a larghezza desktop e ridotta.

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
