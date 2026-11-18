# Installazione e configurazione

## Installare DXRegia

Su Debian o Ubuntu con systemd puoi usare l'installatore nativo. Prepara le dipendenze Perl, il servizio e la directory dati, poi stampa l'URL del pannello e la password temporanea. Docker è disponibile come alternativa. Segui la [procedura Linux con porte personalizzabili](../dxregia/deploy/README.md). Installazione, primo avvio con TOTP, Telnet, backup offline e riavvio sono stati verificati su Ubuntu 24.04 amd64 e arm64. Il collaudo su Proxmox LXC è ancora aperto.

Per verificare la raggiungibilità Telnet dal computer del logger, esegui `python3 dxregia/deploy/check_telnet.py HOST PORTA` usando l'indirizzo e la porta effettivi del nodo. Il test controlla il prompt. Per verificare anche il flusso degli spot, effettua il login con il logger o client Telnet e usa `show/dx`. I comandi amministrativi via Telnet vanno usati su una rete fidata perché il protocollo trasmette in chiaro.

## Primo avvio del pannello DXRegia

L'installatore avvia il pannello anche prima che DXSpider sia configurato. Dopo il cambio password, il sysop completa tutte le impostazioni del nodo dal browser. DXRegia genera e aggiorna i file interni richiesti da Mojo.

1. **Installazione.** Esegui l'installatore dal checkout o dall'archivio nativo della tua architettura. Stampa URL del pannello, utente e password temporanea.
2. **Primo accesso.** Apri l'URL e accedi con le credenziali mostrate. Imposta una nuova password prima di proseguire.
3. **Configurazione del nodo.** Nel wizard inserisci nominativo del nodo, nominativo del sysop, locator e porta Telnet. Controlla il riepilogo e applica la configurazione.
4. **Nodo partner.** Apri Nodi partner → Aggiungi nodo. Compila nominativo, host e porta remota. Usa Verifica indirizzo per controllare DNS e TCP e Salva e connetti per avviare il collegamento. Lo stato segue l'esito comunicato dal motore.
5. **Verifica e backup.** Controlla che il logger riceva gli spot via Telnet e crea un backup completo con lo strumento offline descritto nella guida Linux. Dopo un riavvio verifica che configurazione e collegamenti siano recuperati.

[Specifiche dell'installatore](../dxregia/deploy/README.md) · [Gestione dei nodi partner e backup](OPERATIONS.md) · [Stato di implementazione](../TASK.md)

## Codice 2FA rifiutato

Attiva data e ora automatiche sul telefono. Sul server Linux controlla la sincronizzazione con `timedatectl status`. Se DXRegia gira in un container LXC su Proxmox, esegui il controllo sull'host Proxmox. Se gira in una VM, controlla la VM. Quando il sistema risulta sincronizzato, seleziona nell'app la voce DXRegia associata a questa installazione e attendi il codice successivo.

Un codice rifiutato come già usato va sostituito con quello della finestra successiva. Se i codici nuovi continuano a essere rifiutati con gli orologi sincronizzati, puoi usare uno dei codici di recupero salvati durante l'attivazione per associare di nuovo l'app.

## Importazione della configurazione iniziale

Da Prima configurazione, apri Importa configurazione DXSpider e scegli DXVars.pm. Puoi poi scegliere Listeners.pm per importare la porta di un unico listener. Il modulo riconosce anche oggetti JSON di impostazioni con campi supportati. I file Perl vengono letti come testo e le assegnazioni letterali riconosciute compilano i campi. Controlla i dati e applica dal wizard. Gli export di utenti, le regole Perl personalizzate, i partner e i filtri richiedono procedure di migrazione dedicate.

Il wizard consente di configurare nome, indicativo, email, località, locator e coordinate del sysop, indicativo del nodo, porta Telnet e host pubblico. L'host pubblico serve come riferimento di configurazione. Non modifica DNS, firewall o inoltro porte. La porta web si imposta ancora all'avvio del backend e, in Docker, la porta pubblicata deve corrispondere alla configurazione del container.

## Migrare un nodo Mojo esistente

Su un nodo fermo con banca utenti SQLite e un listener Telnet raggiungibile da loopback, crea una copia separata:

~~~sh
python3 -B dxregia/deploy/migrate_mojo.py /percorso/mojo /percorso/mojo-dxregia
~~~

Il comando conserva i file di configurazione, gli script e i dati, controlla la banca SQLite e verifica la presenza degli account del nodo e del sysop. I dati di gestione vengono aggiunti alla copia. Avvia il backend con la nuova directory tramite -setup-root e con Perl tramite -setup-perl. La prima credenziale del pannello è separata dagli account Mojo. Sostituisci il nodo in esercizio solo dopo aver verificato Telnet, filtri e collegamenti nella copia. Un Local.pm personalizzato viene conservato e può richiedere l'integrazione manuale con l'hook di presenza DXRegia. Le banche utenti precedenti a SQLite devono prima essere convertite con Mojo.
