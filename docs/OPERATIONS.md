# Operazioni del sysop

## Nodi partner

Per ogni collegamento mostrare nominativo, host e porta, tipo di nodo, direzione, stato, durata della sessione, ultimo traffico, ultimo errore e prossimo tentativo. Distinguere connesso, in connessione, in attesa di riconnessione, disconnesso e disabilitato. Lo stato deve venire da DXSpider e indicare quando l'ultima osservazione non è più attuale.

Il dettaglio mostra la cronologia delle connessioni e riconduce ai relativi eventi. Quando il motore fornisce il dato, distinguere errori DNS, TCP, login, protocollo e timeout. Offrire test di connessione, connessione, disconnessione e gestione della riconnessione. Registrare le azioni con esito e operatore.

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
