# Guida rapida a DXRegia

## Per chi gestisce il nodo

1. Installa DXRegia seguendo la [guida di installazione](INSTALLATION.it.md). Apri l’indirizzo del pannello stampato dall’installatore, accedi come sysop e sostituisci la password temporanea.
2. Imposta l’indicativo del nodo con SSID, l’indicativo del sysop, il locator e la porta Telnet. In **Stato e sessioni** verifica che il listener sia pronto. Prova l’indirizzo pubblico da una rete esterna se intendi ricevere connessioni da Internet.
3. In **Server di posta → SMTP** salva la configurazione e usa **Invia prova**. Da **Nodi partner** registra l’indicativo remoto completo, l’host e la porta. Scegli la direzione e le credenziali concordate nei [modelli di peering](PEERING.md).
4. Cerca il nominativo nel **Log DXSpider** se la sessione non diventa attiva. **In ingresso** indica che l’ha avviata il partner. **In uscita** indica che l’ha avviata il tuo nodo. Una sessione attiva può scambiare spot in entrambe le direzioni secondo protocollo e filtri.
5. Crea un backup da **Gestione → Backup** e verificane l’integrità. Consulta le [operazioni del sysop](OPERATIONS.md) per diagnostica e conservazione.

## Per radioamatori

- Apri [la pagina degli spot](https://dxregia.iu3wjq.com/spots) oppure l’indirizzo del tuo nodo. Usa i filtri per banda, modo o nominativo.
- In **Spot in diretta** salva i filtri nel browser, confronta due bande o metti il feed in pausa. **Riprendi dalla pausa** mostra quanto è arrivato nel frattempo. **Esporta CSV** usa i risultati della vista attiva.
- In **Ricerca e DXCC** filtra gli spot recenti per continente, CQ, ITU o locator citato. Apri un nominativo per la cronologia e un’entità del catalogo per vedere le bande osservate nell’archivio disponibile.
- In **Allarmi radio** salva una regola per nominativo, banda, modo, frequenza, nuova entità DXCC o attività portatile e mobile. Tieni aperto il portale per rilevare i nuovi spot. Puoi attivare le notifiche del browser e configurare ore di silenzio. Il registro continua a raccogliere gli avvisi durante il silenzio.
- Per un logger o un client Telnet usa host e porta pubblicati dal sysop. Al prompt usa il tuo indicativo. Se il nodo richiede registrazione, segui il modulo **Registrati** e conferma l’indirizzo email quando il servizio è attivo.
- Prova i comandi HELP e SHOW/DX 10. Se il nodo ti autorizza a pubblicare spot, un esempio di sintassi è DX 14025 JA1ABC CW. Invia uno spot solo per un ascolto o un collegamento reale.
- Usa l’indicativo con SSID indicato dal sysop soltanto quando configuri un collegamento tra nodi. Per un accesso da operatore usa il tuo nominativo personale.

## Protocolli e interfacce

| Protocollo o interfaccia | Utilizzo | Collegamento |
| --- | --- | --- |
| HTTP / HTTPS | Spot, API e pannello sysop | HTTP dal servizio. HTTPS tramite reverse proxy |
| HTTP SSE | Aggiornamenti della dashboard autenticata | GET /api/live, con riconnessione automatica del browser |
| Telnet | Logger, terminali, comandi del cluster | Host e porta del listener. Regole di accesso del nodo |
| Peering DXSpider | Scambio di spot e messaggi tra cluster | TCP/Telnet con inizializzazione DXSpider. Indicativo completo e autenticazione concordata |
| Peering CC Cluster | Collegamento con un nodo CC Cluster | In ingresso può usare l’indicativo del partner e un IP pubblico o DNS autorizzato |
| SMTP | Email di prova, istruzioni e conferme | STARTTLS o TLS, secondo fornitore |

## Glossario

| Termine | Significato |
| --- | --- |
| Spot | Segnalazione di una stazione, con nominativo, frequenza e ora |
| DX | Stazione segnalata nella riga dello spot |
| Spotter | Stazione che ha trasmesso la segnalazione |
| Entità DXCC | Entità radioamatoriale associata al nominativo, anche quando non coincide con uno Stato |
| Locator | Riferimento geografico Maidenhead, per esempio JN55aa |
| SSID | Suffisso dell’indicativo, per esempio -7 in IU3WJQ-7 |
| Sysop | Operatore che amministra il nodo |
| Peering | Collegamento tra due nodi per scambiare informazioni |
| Hop | Budget di inoltro di un messaggio tra nodi |
| UTC | Ora di riferimento usata negli spot e nei log |

## Installazione su Raspberry Pi

Usa un’immagine a 64 bit con systemd. Raspberry Pi OS Lite a 64 bit, disponibile tramite Raspberry Pi Imager, è un’opzione per Pi 4 e Pi 5. Verifica sul dispositivo con uname -m, che deve restituire aarch64. La CI verifica Ubuntu ARM64. Il collaudo su un Raspberry Pi fisico è una verifica distinta. [Documentazione ufficiale di Raspberry Pi](https://www.raspberrypi.com/documentation/computers/os.html)

1. Collegati al Pi via SSH o terminale locale. Verifica di avere curl e sudo o una shell root.
2. Scarica e avvia l’installatore della release ARM64. Lo script sceglie automaticamente il pacchetto.

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.14.0/dxregia/deploy/install-release.sh
sh install-release.sh v0.14.0
```

3. Apri da un PC della LAN l’indirizzo stampato, completa il wizard e prova Telnet sulla porta configurata. Per aggiornamenti e backup usa la [guida dell’installatore Linux](../dxregia/deploy/README.md).

## Migrazione da DXSpider

1. Identifica prima il formato della banca utenti del nodo. Per una configurazione nuova, il wizard può importare i dati supportati di DXVars.pm e Listeners.pm. Per mantenere utenti e dati serve migrare l’intero nodo.
2. Il migratore gestito richiede un albero DXSpider Mojo con local_data/dxusers.db in SQLite, nominativi sysop e nodo presenti nella banca utenti e un solo listener Telnet compatibile. Per un’installazione con users.v2, users.v3 o users.v3j, prepara e verifica prima una conversione upstream a Mojo/SQLite in un ambiente separato.
3. Ferma il nodo sorgente e crea un backup. Usa un percorso di destinazione nuovo con directory padre già esistente. Il migratore verifica il database e prepara la copia gestita.

~~~sh
python3 -B dxregia/deploy/migrate_mojo.py /percorso/mojo /percorso/mojo-dxregia
~~~

4. Avvia la copia in un ambiente isolato con il flag -setup-root puntato alla destinazione, seguendo le [istruzioni per il backend](../dxregia/backend/README.md). Verifica utenti, filtri, Telnet e collegamenti prima di trasferire l’indirizzo di produzione. Tieni il nodo originale e il backup disponibili per un eventuale ripristino.
