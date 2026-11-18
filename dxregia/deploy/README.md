# Installazione Linux

## Installazione di prova da sorgente

Da Linux con Docker e curl già disponibili, clonare il repository e lanciare `sh dxregia/deploy/install-local.sh` dalla sua radice. Lo script costruisce l'immagine, avvia un container con volume persistente e stampa URL e password temporanea. La riesecuzione conserva il container e il volume. La password temporanea è custodita nel volume fino al cambio obbligatorio.

L'anteprima pubblica sulla macchina locale le porte HTTP 8080 e Telnet 7300. Nel wizard usare la porta proposta 7300, che corrisponde alla mappatura del container. La build e il primo avvio con Docker restano da verificare su Linux. La procedura da release firmata e l'accesso dalla LAN sono attività successive.

## Esperienza di installazione

1. Il sysop scarica e avvia un solo script sul PC Linux. Lo script verifica architettura, sistema e privilegi, installa le dipendenze mancanti e scarica una release DXRegia verificandone l'integrità.
2. Lo script prepara dati e configurazione persistenti, avvia il servizio e ne verifica la risposta HTTP. Se qualcosa fallisce, indica il passaggio e dove consultare i log.
3. Alla prima installazione completata, lo script mostra nel terminale l'indirizzo del pannello, con porta e un indirizzo utilizzabile da un altro dispositivo della LAN quando disponibile, il nome utente amministrativo e una password temporanea casuale.
4. Al primo login il sysop deve cambiare la password prima di poter accedere alle funzioni di gestione. Il backend applica il vincolo anche alle API. Dopo la modifica invalida la password temporanea e apre il wizard di configurazione DXSpider.

## Installazioni e riesecuzioni

- La prima distribuzione prevista è OCI per amd64 e arm64. Supporto e procedura automatica saranno documentati per le distro effettivamente provate.
- Lo script conserva configurazione, dati e account alla riesecuzione. La password temporanea viene generata per la prima inizializzazione e non finisce nei log. Il backend conserva soltanto il suo hash.
- La configurazione iniziale è pensata per accesso locale o dalla LAN. L'installazione non pubblica il pannello su Internet e non apre automaticamente porte sul router.
- Il controllo finale verifica che l'indirizzo stampato corrisponda al servizio avviato. Un messaggio distingue l'indirizzo locale da quello LAN.

## Accettazione

Su una macchina Linux pulita, eseguire lo script e aprire il pannello all'indirizzo stampato da un browser nella LAN. Verificare che la password temporanea richieda il cambio, che quella precedente venga rifiutata dopo il cambio, e che riavvio e riesecuzione conservino dati e credenziali. Provare anche un download fallito e un servizio che non risponde.

## Controllo base di Telnet

Eseguire python3 dxregia/deploy/check_telnet.py HOST PORTA dal sistema che deve raggiungere il nodo. Verifica il prompt di login senza inviare credenziali o comandi. La compatibilità con login, logger e collegamenti tra nodi va provata separatamente.

## Avvio del pannello prima di Mojo

L'installatore avvia e verifica il pannello anche con una directory dati nuova, senza configurazione Mojo. Il wizard raccoglie i valori dal browser. Il backend valida i dati, genera la configurazione interna e inizializza il nodo, poi avvia Mojo e verifica il listener Telnet. Se un passaggio fallisce, mostra l'errore e consente di correggere e riprovare dal browser. Il wizard deve poter riprendere dopo un riavvio e non deve reinizializzare una banca utenti esistente.

Verifica di accettazione: dopo lo script, completare cambio password e wizard dal browser, verificare Telnet e riavviare. Nessun passaggio richiede di creare o modificare file a mano.

## Backup offline del nodo

Richiede Python 3.10 o successivo. Arresta DXRegia e Mojo prima di copiare il nodo. Per un container, esegui lo strumento sull'archivio persistente con il servizio arrestato.

~~~sh
python3 dxregia/deploy/backup_node.py backup /percorso/mojo /percorso/salvataggi/nodo.zip
python3 dxregia/deploy/backup_node.py verify /percorso/salvataggi/nodo.zip
python3 dxregia/deploy/backup_node.py restore /percorso/salvataggi/nodo.zip /percorso/mojo-ripristinato
~~~

Il ripristino richiede una directory nuova. Lo strumento controlla percorsi, dimensioni e checksum prima di pubblicare i dati ripristinati. Il backup include motore e dati del nodo, banca utenti, filtri, anagrafica, impostazioni e dati di accesso. Tratta l'archivio come riservato. Il backend DXRegia e le dipendenze del sistema si installano separatamente. Il pannello offre anche il download delle sole impostazioni del nodo come JSON riutilizzabile nel wizard.
