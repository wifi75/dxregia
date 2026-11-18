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

Per la notifica al sysop, impostare DXREGIA_SMTP_ADDRESS (host:porta SMTP con TLS implicito), DXREGIA_SMTP_FROM, e se richiesti DXREGIA_SMTP_USER e DXREGIA_SMTP_PASSWORD nell'ambiente del backend. Il destinatario è l'email sysop inserita nel wizard. Un worker invia le notifiche in coda e aggiorna l'esito. Impostare queste variabili anche nell'ambiente del container se si usa Docker. In assenza di SMTP le richieste restano consultabili dal pannello.

Flusso da completare: approvazione o abilitazione automatica opzionale dopo verifica dell'indirizzo email e controllo della richiesta, invio dei parametri di collegamento ufficiali e successivo controllo del link reale in Mojo. Comunicare successo solo dopo lo stato operativo. In caso di fallimento riportare l'ultimo errore rilevato o chiedere i dati mancanti.

### Abilitazione in ingresso

Dalla scheda salvata, il sysop può scegliere Abilita accesso in ingresso. La richiesta viene eseguita dall'hook Mojo gestito. La scheda mostra la password generata al momento dell'accodamento e l'esito di Mojo quando si aggiorna lo stato. La password serve per il collegamento sul nodo partner. L'account del sysop e gli account preesistenti incompatibili non vengono modificati. Un account abilitato può risultare ancora disconnesso.
