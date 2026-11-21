# Guida all'installazione e configurazione

[English](INSTALLATION.md) · **Italiano** · [Español](INSTALLATION.es.md) · [Deutsch](INSTALLATION.de.md) · [Français](INSTALLATION.fr.md) · [Português](INSTALLATION.pt.md)

[← Torna alla pagina del progetto](../README.it.md)

## 1. Cosa serve

- Un server o PC Linux con systemd: Debian o Ubuntu (provato su Ubuntu 24.04, amd64 e arm64). Va bene anche una macchina virtuale o un container LXC non privilegiato.
- Per l'installazione da sorgente: Git, Go 1.22+, Node.js 22.12+ con npm, curl e sudo (o root). L'alternativa è Docker, che richiede solo Docker e curl.
- Due porte TCP: una per il pannello web (predefinita 8080) e una per Telnet (predefinita 7300). Se altri nodi o utenti devono raggiungerti da Internet, inoltra la porta Telnet sul router.

## 2. Installazione

**Pacchetto precompilato.**

```sh
curl -fL -o install-release.sh https://raw.githubusercontent.com/wifi75/dxregia/v0.9.0/dxregia/deploy/install-release.sh
sh install-release.sh v0.9.0
```

Lo script verifica il checksum SHA-256, installa le dipendenze e avvia il servizio. Stampa l’indirizzo del pannello e una password temporanea per il primo accesso.

**Installazione nativa da sorgente.** Esegui questi comandi sul server:

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
sh dxregia/deploy/install-native.sh
```

Lo script compila pannello e backend, installa le dipendenze Perl, crea un utente dedicato e avvia il servizio systemd `dxregia`. Il programma sta in `/opt/dxregia` e i dati in `/var/lib/dxregia`. Alla fine stampa l'indirizzo del pannello e, al primo avvio, una password temporanea. Rieseguirlo avvia il servizio già installato e conserva i dati. Installa anche il comando `dxregia-update`.

Per scegliere le porte, impostale prima:

```sh
DXREGIA_WEB_PORT=8787 DXREGIA_TELNET_PORT=7373 sh dxregia/deploy/install-native.sh
```

**Docker.** Con Docker e curl già installati:

```sh
sh dxregia/deploy/install-local.sh
```

Costruisce l'immagine e avvia un container con un volume persistente, poi stampa indirizzo e password temporanea. Rieseguirlo conserva container e volume. Valgono le stesse due variabili per le porte.

## 3. Aprire il pannello

Apri uno degli indirizzi LAN stampati dallo script nativo. Sul server puoi usare anche `http://127.0.0.1:8080/admin`. Una nuova installazione nativa ascolta su tutte le interfacce IPv4. Con Docker usa un tunnel SSH, perché la porta del pannello è pubblicata sul loopback dell’host.

```sh
ssh -L 8080:127.0.0.1:8080 USER@SERVER
```

## 4. Primo accesso

1. Apri `/admin`. Accedi con l'utente `sysop` e la password temporanea stampata dall'installatore.
2. DXRegia chiede una nuova password (almeno 12 caratteri). Quella temporanea smette di funzionare.
3. Il 2FA è facoltativo e disattivato di default. Per attivarlo vai in **Gestione → Password → Verifica in due passaggi**, conferma la password web, scansiona il QR e salva i codici di recupero.

Se hai perso la password temporanea, puoi leggerla sul server prima del primo cambio:

```sh
sudo cat /var/lib/dxregia/sysop-initial
```

Da quel momento ogni accesso chiede utente, password e codice a 6 cifre. L'utente è il tuo indicativo da sysop appena hai completato la configurazione (prima è `sysop`); maiuscole e minuscole non contano.

## 5. Prima configurazione (wizard)

Dopo l'accesso apri Nodo → Configurazione e compila: indicativo del nodo con SSID (per esempio `IU3WJQ-7`), indicativo del sysop, locator, porta Telnet e, se vuoi, nome, email, QTH, coordinate e un nome host pubblico. Il nome host pubblico è solo un riferimento: non modifica DNS né router. Se usi già DXSpider puoi importare `DXVars.pm` e `Listeners.pm` dalla stessa pagina. Applica: DXRegia scrive i file che servono a Mojo e avvia il nodo, e il pannello conferma quando Telnet risponde.

## 6. Configurare il resto

- **Telnet e firewall.** Verifica che la porta Telnet sia aperta nel firewall. Per essere raggiungibile da Internet, inoltrala sul router verso il server. Le opzioni sono in Nodo → Opzioni Telnet.
- **Server di posta.** Gestione → Server di posta. Scegli un preset (Gmail, Aruba, Office 365) oppure scrivi server e porta, cifratura, indirizzo mittente, utente e password, poi Salva e premi Invia prova. La password resta sul server in un file riservato. Con Gmail serve una password per le app. La posta serve per gli avvisi a te e per le istruzioni inviate ai partner.
- **Nodi partner.** Collegamenti → Nodi partner → Aggiungi nodo: nominativo, host, porta e, se l'altro nodo la richiede, la password remota. Premi Salva e connetti. Spunta la casella di riprova per riconnettersi da solo dopo una caduta.
- **Backup.** Vedi la sezione 9.
- **Lingua, tema e suono.** Le icone in alto a destra nelle pagine pubbliche.

### Richieste partner e collegamenti in ingresso

- **Modulo pubblico.** In **Collegamenti → Richieste ricevute**, attiva **Configura da solo i nodi che chiedono il collegamento**. Con SMTP e l’host e la porta Telnet pubblici del nodo configurati, DXRegia crea le schede dei nuovi partner idonei. Appena DXSpider abilita l’accesso in ingresso, DXRegia invia i dati di collegamento e una password dedicata. Il limite predefinito è di 10 configurazioni in 24 ore. Se disattivi l’automazione, apri la richiesta e usa **Prepara scheda partner**.
- **Accesso CC Cluster senza password.** In **Collegamenti → Nodi partner**, scegli **Autenticazione in ingresso → Senza password**, conferma l’IP pubblico o il nome host da cui il partner si collega e premi **Salva e invia senza password**. DXRegia propone l’host salvato nella scheda e aggiorna gli indirizzi DNS. Il partner deve collegarsi da uno di questi indirizzi. Puoi anche scegliere **Salva senza inviare**.
- **Conferma del collegamento.** Dopo l’invio delle istruzioni DXRegia attende per un massimo di sette giorni una sessione attiva avviata dal partner, poi invia una seconda email quando la rileva. La sessione in uscita non fa scattare questa conferma. Controlla l’accettazione da parte di SMTP e il contenuto in **Collegamenti → Storico email partner**. Se la sessione in uscita occupa il collegamento, **Prova ingresso ora** apre una finestra di cinque minuti. La prova automatica è facoltativa e ripete la finestra ogni 30 minuti finché l’ingresso è disconnesso.

## 7. Verificare che funzioni

- Apri `http://SERVER:8080/spots`: la pagina degli spot in diretta.
- Nodo → Stato e sessioni: le quattro tessere devono essere verdi.
- Da un PC prova Telnet, poi accedi con un client Telnet o un logger e digita `show/dx`:

```sh
python3 dxregia/deploy/check_telnet.py HOST 7300
```

Telnet trasmette in chiaro: usa i comandi amministrativi solo su una rete fidata.

## 8. Aggiornare

Su un'installazione nativa esegui, da qualsiasi directory:

```sh
dxregia-update
```

Scarica l'ultima versione, la compila, conserva la copia precedente e torna indietro da solo se pannello o Telnet smettono di rispondere. Dati, credenziali, porte e personalizzazioni restano. `dxregia-update --check` mostra soltanto la versione installata e quella disponibile. Il servizio si riavvia: i collegamenti cadono per circa un minuto e si ristabiliscono da soli, e devi accedere di nuovo al pannello. Le installazioni fatte prima che esistesse questo comando possono prendere `dxregia/deploy/update-native.sh` ed eseguirlo con `sh update-native.sh`. Con Docker usa `sh update-docker.sh`.

Puoi aggiornare anche dal pannello di amministrazione: **Gestione > Aggiornamenti**. La pagina mostra la versione installata e quella disponibile, **Controlla ora** interroga GitHub e **Aggiorna ora** avvia lo stesso aggiornamento dopo una conferma. L'avanzamento compare in diretta e la pagina si ricollega dopo il riavvio. Telnet e i collegamenti con i partner si interrompono per circa un minuto. Se hai installato prima di questa funzione, esegui una volta `dxregia-update` da terminale per attivarla. Quando cambiano solo le pagine del pannello o la documentazione, i file vengono sostituiti senza riavvio e Telnet resta collegato.

## 9. Backup e ripristino

Lo strumento offline richiede Python 3.10+. Ferma prima il servizio, così la copia è coerente:

```sh
sudo systemctl stop dxregia
python3 dxregia/deploy/backup_node.py backup /var/lib/dxregia/mojo /path/to/node.zip
python3 dxregia/deploy/backup_node.py verify /path/to/node.zip
sudo systemctl start dxregia
```

Salva motore e dati del nodo, utenti, filtri, partner, impostazioni e dati di accesso: tratta l'archivio come riservato. Per verificare un archivio dal pannello usa Gestione → Backup. Per ripristinare, usa una directory nuova e vuota:

```sh
python3 dxregia/deploy/backup_node.py restore /path/to/node.zip /path/to/restored-node
```

## 10. Spostare un nodo Mojo esistente

Su un nodo fermo con banca utenti SQLite, crea una copia separata e avvia DXRegia su quella. Sostituisci il nodo in esercizio solo dopo aver verificato Telnet, filtri e collegamenti sulla copia:

```sh
python3 -B dxregia/deploy/migrate_mojo.py /path/to/mojo /path/to/mojo-dxregia
```

## 11. Risoluzione dei problemi

- **Non riesco ad accedere.** L'utente è il tuo indicativo da sysop (o `sysop` prima della configurazione), non l'indicativo del nodo. Dopo un tentativo sbagliato aspetta un secondo.
- **Il codice a due fattori viene rifiutato.** Gli orologi devono coincidere: attiva data e ora automatiche sul telefono e controlla il server con `timedatectl status` (per un container LXC controlla l'host). Se un codice risulta già usato, passa al successivo. In ultima istanza usa un codice di recupero.
- **Il pannello non si apre da un altro PC.** Segui la sezione 3 e controlla il firewall per la porta web.
- **Telnet non è raggiungibile.** Controlla che la porta sia in ascolto (`ss -ltn | grep 7300`), aprila nel firewall e, da Internet, inoltrala sul router.
- **Un partner non si collega.** Controlla host, porta e password, e quale lato avvia il collegamento: nella scheda del partner, «si collega al tuo nodo» richiede che la tua porta Telnet sia raggiungibile dall'esterno.
- **Log.** Sul server:

```sh
sudo journalctl -u dxregia -f
```

Il log del motore è anche in Nodo → Log DXSpider.

## Altro

- [Riferimento dell'installatore Linux](../dxregia/deploy/README.md)
- [Operazioni del sysop](OPERATIONS.md)
- [Registro delle modifiche](../CHANGELOG.md) (in inglese)

## DXRegia 0.7.0

La pagina Propagazione mostra gli indici solari HamQSL in schede quadrate e colorate, le condizioni HF diurne e notturne, l’attività VHF e la mappa MUF di KC2G. La registrazione pubblica spiega come configurare Telnet in sette logger. La dashboard sysop mostra fino a 40 radioamatori collegati per pagina su schermi ampi, con le bandiere DXCC.

RBN è disattivato di default. Apri **Collegamenti → Reverse Beacon Network** nel pannello sysop e premi **Attiva RBN** per avviare il collegamento in uscita CW/RTTY. La scheda **RBN** del portale pubblico mostra gli spot ricevuti. In Telnet o nel logger, digita `SET/SKIMMER CW RTTY` per riceverli e `UNSET/SKIMMER` per interromperli.

[DXRegia 0.7.0](https://github.com/wifi75/dxregia/releases/tag/v0.7.0)
