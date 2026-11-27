# Installazione Linux

## Installazione da release GitHub

Lo script scarica il pacchetto nativo amd64 o arm64, ne verifica lo SHA-256 e avvia l’installatore. Senza argomenti usa l’ultima release pubblicata. Per installare la versione 0.17.1, scarica lo script dal tag e avvialo così.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.17.1/dxregia/deploy/install-release.sh
sh install-release.sh v0.17.1
```

[Pacchetti e checksum di DXRegia 0.17.1](https://github.com/wifi75/dxregia/releases/tag/v0.17.1)

## Pacchetto nativo precompilato

Il pacchetto tar.gz contiene il backend, il frontend, Mojo e gli strumenti di installazione e backup. Per produrlo da un host Linux con Go e Node.js:

~~~sh
mkdir -p pacchetti
sh dxregia/deploy/build-native-bundle.sh pacchetti
~~~

Sul server di destinazione, verifica il checksum e installa il pacchetto della sua architettura. Per esempio, su amd64:

~~~sh
sha256sum -c dxregia-linux-amd64.tar.gz.sha256
tar -xzf dxregia-linux-amd64.tar.gz
sh dxregia-linux-amd64/dxregia/deploy/install-native.sh
~~~

L'installatore usa direttamente il binario incluso e installa le dipendenze Perl. Il server non richiede Go, Node.js o Docker.

## Installazione nativa da sorgente

Per Debian o Ubuntu con systemd, usa il checkout DXRegia e Go 1.22+, Node.js 22.12+ con npm, curl e sudo (o root).

~~~sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
~~~

Lo script compila frontend e backend, installa le dipendenze Perl e avvia il servizio systemd dxregia con il suo utente dedicato. Stampa l'indirizzo del pannello e, al primo avvio, la password temporanea. Le porte predefinite sono 8080 e 7300. I dati sono in /var/lib/dxregia e il runtime è in /opt/dxregia. Una nuova esecuzione avvia il servizio già installato e conserva i dati. Una nuova installazione nativa ascolta su tutte le interfacce IPv4 e stampa gli indirizzi IPv4 disponibili. Usa l’utente `sysop`, la password temporanea stampata e cambia la password al primo accesso. Il 2FA è facoltativo e si attiva in Gestione → Password. La password temporanea si può leggere con `sudo cat /var/lib/dxregia/sysop-initial` prima del primo cambio.

Il listener Telnet di DXSpider ascolta su `0.0.0.0` sulla porta scelta. Il firewall regola da quali reti accettare connessioni. Il pannello web usa un indirizzo di ascolto separato. All'avvio di un nodo gestito, DXRegia applica lo stesso standard anche alle installazioni precedenti e conserva una copia dell'eventuale configurazione sostituita.

Il collaudo automatico nativo è passato su Ubuntu 24.04 con systemd, su amd64 e arm64. Una LXC non privilegiata con systemd può usare lo stesso percorso; il collaudo Proxmox resta separato. Le immagini Docker rimangono disponibili.

## Aggiornare un'installazione nativa

Esegui `dxregia-update` da qualsiasi directory. Per le installazioni precedenti, scarica `update-native.sh` da `main` e avvialo con `sh update-native.sh`, oppure installalo con `sudo install -m 0755 update-native.sh /usr/local/bin/dxregia-update`. Servono Git, Go, Node.js con npm, curl, Python 3, flock e privilegi sudo o root. Lo script scarica `main`, compila backend e frontend, aggiorna il motore Mojo e lo strumento di backup e riavvia il servizio. Conserva dati, credenziali, porte e personalizzazioni in `local`, `local_cmd` e `local_data`. Le directory upstream distribuite con il programma vengono sostituite. Lo script conserva le copie precedenti e le ripristina se il pannello o il listener Telnet non risponde.

Usa `sh update-native.sh --check` per vedere il commit installato e quello disponibile senza modificare il servizio. L'avvio normale salta l'aggiornamento se coincidono. Durante l'installazione lo script indica le cinque fasi e usa segni verdi per gli esiti positivi e rossi per gli errori su un terminale a colori.

## Installazione di prova da sorgente

Da Linux con Docker e curl già disponibili, clonare il repository e lanciare `sh dxregia/deploy/install-local.sh` dalla sua radice. Lo script costruisce l'immagine, avvia un container con volume persistente e stampa URL e password temporanea. La riesecuzione conserva il container e il volume. La password temporanea è custodita nel volume fino al cambio obbligatorio.

Per scegliere le porte prima dell'installazione, esegui ad esempio:

~~~sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-local.sh
~~~

I valori predefiniti restano 8080 e 7300. La porta Telnet è pubblicata su tutte le interfacce IPv4. L'aggiornamento Docker applica lo stesso indirizzo alle installazioni precedenti. Il pannello resta sul loopback e il wizard usa la porta Telnet effettivamente mappata. Nelle riesecuzioni usa gli stessi valori: Docker assegna le porte al momento della creazione del container. La build e il primo avvio con Docker sono stati verificati in CI su Ubuntu 24.04 amd64 e arm64 con l'immagine Debian. La procedura da release firmata e l'accesso dalla LAN sono attività successive.

## Aggiornare il container Docker

Scarica `update-docker.sh` e lancialo sull'host Docker con `sh /percorso/update-docker.sh`, da qualunque directory. Serve un utente autorizzato a usare Docker, con Git, curl e Python 3 disponibili. Lo script e destinato al container `dxregia` creato con `install-local.sh`. Costruisce la nuova immagine prima di fermare il servizio. Conserva volume, numeri di porta, variabili DXRegia, configurazione e dati. Dopo il cambio verifica il pannello e, per un nodo configurato, Telnet. Se il controllo fallisce ripristina il vecchio container e le parti del motore sostituite.

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

## Verificare un archivio dal pannello

Accedi come sysop e apri **Verifica un backup**. Seleziona un archivio DXRegia ZIP fino a 512 MiB e premi **Verifica archivio**. Il controllo confronta il manifest con il contenuto e verifica gli SHA-256. Un archivio verificato può essere ripristinato con il comando offline descritto sotto. Il controllo carica temporaneamente l'archivio sul server e lo elimina al termine.

## Backup offline del nodo

Richiede Python 3.10 o successivo. L'installazione nativa include il comando `/opt/dxregia/bin/dxregia-backup` e l'immagine Docker include `/opt/dxregia/backup_node.py`. Arresta DXRegia e Mojo prima di copiare il nodo. Per un container, esegui lo strumento sull'archivio persistente con il servizio arrestato.

~~~sh
python3 dxregia/deploy/backup_node.py backup /percorso/mojo /percorso/salvataggi/nodo.zip
python3 dxregia/deploy/backup_node.py verify /percorso/salvataggi/nodo.zip
python3 dxregia/deploy/backup_node.py restore /percorso/salvataggi/nodo.zip /percorso/mojo-ripristinato
~~~

Il ripristino richiede una directory nuova. Lo strumento controlla percorsi, dimensioni e checksum prima di pubblicare i dati ripristinati. Il backup include motore e dati del nodo, banca utenti, filtri, anagrafica, impostazioni e dati di accesso. Tratta l'archivio come riservato. Il backend DXRegia e le dipendenze del sistema si installano separatamente. Il pannello offre anche il download delle sole impostazioni del nodo come JSON riutilizzabile nel wizard.
