# Installazione Linux

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
