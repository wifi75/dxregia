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
