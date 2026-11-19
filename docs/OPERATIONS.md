# Operazioni del sysop

## Nodi partner

Per ogni collegamento mostrare nominativo, host e porta, tipo di nodo, direzione, stato, durata della sessione, ultimo traffico, ultimo errore e prossimo tentativo. Distinguere connesso, in connessione, in attesa di riconnessione, disconnesso e disabilitato. Lo stato deve venire da DXSpider e indicare quando l'ultima osservazione non è più attuale.

Il dettaglio mostra la cronologia delle connessioni e riconduce ai relativi eventi. Quando il motore fornisce il dato, distinguere errori DNS, TCP, login, protocollo e timeout. Offrire test di connessione, connessione, disconnessione e gestione della riconnessione. Registrare le azioni con esito e operatore.

### Aggiunta di un nodo partner

Da **Nodi partner → Aggiungi nodo**, chiedere nominativo remoto, host e porta. Preselezionare il profilo DXSpider. Lasciare vuoto il campo obbligatorio Porta remota e accettare qualsiasi numero intero da 1 a 65535. Il valore 7300 puo comparire come esempio. Usare la porta inserita per la verifica TCP e la connessione, indipendentemente dalla porta del listener locale. Mostrare tipo di cluster, credenziali del collegamento e riconnessione nelle opzioni avanzate.

**Verifica indirizzo** controlla DNS e raggiungibilità TCP senza inviare credenziali. **Salva e connetti** salva la configurazione, avvia il collegamento e mostra nella stessa schermata la fase osservata. Indicare come riuscito il collegamento solo quando il motore lo conferma. Se il tentativo fallisce, conservare la definizione per correggerla o riprovare. Se il nodo remoto richiede una configurazione da parte del suo sysop, spiegarlo nel dettaglio del tentativo.

Verifica di usabilità: un sysop deve poter aggiungere un partner DXSpider conoscendo nominativo, host e porta, completare il flusso senza console e correggere una porta errata dalla stessa schermata. Provare con un nodo di laboratorio in ascolto su una porta diversa da 7300.

## Registro eventi

Vista cronologica con ricerca, filtri per periodo UTC, gravità, categoria e nominativo, dettaglio evento ed esportazione. Categorie iniziali: collegamenti, protocollo, spot e filtri, utenti e accessi, modifiche amministrative, servizio, installazione e aggiornamenti, backup e ripristino.

Mostrare origine, ora UTC, gravità, categoria, esito e identificativo di correlazione quando disponibile. Rendere configurabili conservazione e rotazione in base allo spazio disponibile. Separare il debug dettagliato dal funzionamento ordinario, con attivazione temporanea. Oscurare password, token e altri segreti anche nelle esportazioni. Mostrare quali eventi il motore consente di osservare e indicare eventuali interruzioni della raccolta.

## Configurazioni, backup e ripristino

- Conservare revisioni delle configurazioni con data, autore e confronto, includendo le impostazioni di DXSpider e DXRegia. Offrire importazione, esportazione e ripristino di una revisione.
- Offrire backup manuali e pianificati, rotazione, download autorizzato e una destinazione esterna configurabile. Mostrare ultimo esito, prossimo backup e log dei tentativi.
- Distinguere esportazione della sola configurazione e backup completo. Il backup completo comprende configurazione, utenti, filtri, definizioni dei partner e dati persistenti del nodo. Identificare versione, contenuto e integrità con un manifest. Proteggere eventuali segreti inclusi nel backup ed escluderli dalle esportazioni ordinarie.
- Prima del ripristino, verificare integrità e compatibilità, mostrare cosa verrà ripristinato e creare un backup dello stato corrente. Applicare il ripristino in modo da poter recuperare lo stato precedente se il servizio non riparte.

## Verifiche di accettazione

Simulare la caduta di un collegamento e verificare che stato, cronologia, prossimo tentativo ed evento del log concordino. Verificare filtri, esportazione e oscuramento dei dati sensibili. Ripristinare su un'installazione di prova e controllare che utenti, filtri e partner siano recuperati. Provare un archivio corrotto e un ripristino interrotto.

### Anagrafica dei nodi

La scheda sysop conserva indicativo, nome del nodo, referente, email, host e porta remota. La data di registrazione viene assegnata alla creazione. Le note vengono aggiunte allo storico con data e ora. La prima connessione verrà registrata dopo la conferma del link da Mojo.

## Dashboard sysop e traffico web

La dashboard legge da Mojo i nodi presenti e le relative sessioni. Mostra i partner registrati assenti come disconnessi soltanto se la fotografia di Mojo è recente. Le sessioni radioamatoriali riportano indicativo, canale Telnet o Web Mojo e orario di ingresso. Il registro amministrativo richiede password e codice TOTP. La prima volta si inserisce nell'app autenticatore la chiave mostrata dal pannello e si conferma un codice.

Il portale pubblico mostrerà soltanto il totale delle visite. Nella dashboard privata: visitatori attivi stimati, visite e pagine viste per intervallo, paese quando determinabile, dominio di provenienza, pagina d'ingresso e pagine consultate. I visitatori pubblici senza login sono anonimi. Un referrer assente o una localizzazione indisponibile devono apparire come dati mancanti. L'analytics deve tenere distinti visitatori web del portale e sessioni operatore aperte nel motore.

## Richieste dal portale

Il modulo pubblico /contatti permette di inviare un messaggio o chiedere un collegamento. La richiesta si conserva nella coda privata del sysop. Prepara scheda partner copia i campi e il messaggio nella scheda da verificare e salvare. Una richiesta ricevuta e una scheda salvata non equivalgono a un collegamento attivo.

Per la notifica al sysop, impostare DXREGIA_SMTP_ADDRESS (host:porta del provider SMTP), DXREGIA_SMTP_FROM, e se richiesti DXREGIA_SMTP_USER e DXREGIA_SMTP_PASSWORD nell'ambiente del backend. Per TLS implicito usa `DXREGIA_SMTP_TLS=implicit` (predefinito), normalmente sulla porta 465. Per i provider che richiedono STARTTLS usa `DXREGIA_SMTP_TLS=starttls`, normalmente sulla porta 587. DXRegia verifica il certificato prima di inviare credenziali o messaggi. Il destinatario è l'email sysop inserita nel wizard. Un worker invia le notifiche in coda e aggiorna l'esito. Impostare queste variabili anche nell'ambiente del container se si usa Docker. In assenza di SMTP le richieste restano consultabili dal pannello.

In alternativa alle variabili d'ambiente, il sysop può configurare il server di posta da Gestione → Server di posta: server e porta, cifratura, mittente, utente e password, con preset e invio di una email di prova. Le impostazioni sono in `local_data/dxregia-smtp.json` (file riservato, la password non viene mai restituita dall'API) e valgono subito senza riavvio. Se il file esiste ha priorità sulle variabili `DXREGIA_SMTP_*`; un file non valido disattiva l'invio invece di ripiegare in silenzio sull'ambiente.

In **Collegamenti → Richieste ricevute**, l’opzione **Configura da solo i nodi che chiedono il collegamento** elabora le richieste dal modulo pubblico per nuovi partner idonei. Servono SMTP e l’host e la porta Telnet pubblici. Dopo l’abilitazione in Mojo parte l’email con i dati e la password dedicata. Il limite predefinito è 10 configurazioni nelle ultime 24 ore. Se l’automazione è disattivata, puoi usare **Prepara scheda partner**. La verifica dell’indirizzo email del richiedente resta da implementare.

### Abilitazione in ingresso e conferme email

Nella scheda del partner scegli **Autenticazione in ingresso**. Con la password, l’email di istruzioni parte dopo la conferma locale di Mojo. Se un partner CC Cluster non può inviare la password all’apertura della sessione, puoi scegliere l’accesso senza password con un IP pubblico o nome host autorizzato. Il pannello propone l’host salvato nella scheda. Il DNS viene aggiornato circa ogni minuto e la risoluzione scade dopo cinque minuti senza aggiornamenti validi. Si confronta l’indirizzo effettivo del socket Telnet. Se il partner esce da un indirizzo diverso da quello a cui punta il suo host, usa il suo indirizzo di uscita effettivo.

Dopo l’invio delle istruzioni, DXRegia attende per un massimo di sette giorni una sessione attiva avviata dal partner. La conferma viene inviata al primo ingresso rilevato. Per le istruzioni con password, dopo cinque minuti può partire anche un avviso se l’ingresso non è ancora attivo. Una sessione in uscita non attiva la conferma. **Storico email partner** mostra il contenuto con le credenziali oscurate e l’esito SMTP. L’accettazione SMTP attesta la presa in carico del messaggio.

### Collegamento in uscita e prova in ingresso

Per collegarti al partner, usa **Connetti ora** con la password remota se richiesta. Con **Riprova da solo se il collegamento cade**, il backend conserva le impostazioni e distanzia i tentativi di almeno due minuti. Se host o porta cambiano, la riconnessione resta sospesa finché non salvi le nuove impostazioni.

DXSpider usa una sola sessione per lo stesso indicativo. **Prova ingresso ora** sospende l’uscita per cinque minuti per permettere al partner di aprire la sessione. La modalità automatica ripete la finestra ogni 30 minuti finché l’ingresso è disconnesso, anche a portale chiuso o dopo un riavvio. In **Stato del nodo** e **Nodi partner**, l’elenco in ingresso mostra soltanto le sessioni attive avviate dai partner.

## Aggiornamenti: versioni precedenti

`update-native.sh` conserva le ultime tre versioni precedenti (runtime in `/opt/.dxregia-update-*` e parti del motore in `.dxregia-engine-previous-*` nella directory di stato) e cancella le più vecchie dopo un aggiornamento riuscito. Una copia dello script rimasta da prima dello spostamento del motore in `engine/` non funziona più: va riscaricata da `main`.


## Diagnostica, spazio e conservazione

In **Stato e sessioni → Diagnostica del nodo** trovi il ritardo di aggiornamento delle sessioni all’ultima lettura, l’età del più recente spot nel feed, lo spazio libero sul volume dati e i PID dei processi gestiti. L’assenza del PID di DXSpider può indicare un motore esterno al servizio. La presenza recente indica che il motore pubblica dati. Il riquadro del listener mostra a parte l’esito del controllo Telnet.

Il riquadro misura le dimensioni di log DXSpider, debug, log di avvio e storico email partner ogni 30 secondi. La differenza riporta la variazione rispetto al campione precedente, che può essere negativa dopo una rotazione. Un dato mancante è indicato con un trattino. Lo spazio disco è disponibile su Linux e Windows.

Il portale degli spot misura il tempo tra l’ultima scrittura dell’archivio aggiornato da DXSpider e la visualizzazione dei nuovi spot nel browser. La stima include il polling web e dipende dagli orologi del browser e del server. L’età dell’ultimo spot usa invece l’ora registrata nello spot. I dati coprono il feed, limitato a 100 spot da oggi e ieri in UTC.

Per impostare la conservazione sulla versione nativa, aggiungi le variabili desiderate a /etc/dxregia.env e riavvia con sudo systemctl restart dxregia. Su Docker passale nell’ambiente del container mantenendo il volume dati. Le impostazioni sono facoltative.

| Variabile | Valore | Effetto |
| --- | --- | --- |
| DXREGIA_EVENT_RETENTION_DAYS | 1–3650, vuota per impostazione iniziale | All’avvio e circa una volta all’ora elimina gli archivi DXSpider giornalieri di debug, i log mensili di periodi interamente scaduti e i messaggi archiviati nello storico email più vecchi della soglia. L’età viene controllata anche sulla data di ultima modifica del file. |
| DXREGIA_PUBLIC_STATUS | on, vuota per impostazione iniziale | Pubblica /stato con disponibilità dei dati del motore e ora UTC dell’ultimo spot. Risponde 503 quando non arriva una fotografia recente. |

La conservazione lascia fuori l’archivio spot, il database utenti, le configurazioni, le email ancora in coda e il file di avvio attivo. Prima di abilitare la cancellazione, salva le copie dei periodi che vuoi conservare.

Il flusso live amministrativo usa ID con epoca e sequenza. Alla riconnessione con Last-Event-ID, invia subito un aggiornamento se il client è indietro e la dashboard rilegge la fotografia corrente. Il feed pubblico usa un ETag e risponde 304 quando la fotografia è identica. Le sequenze ripartono a ogni processo e la ripresa restituisce lo stato disponibile al momento della richiesta.
