# Installazione e configurazione

## Cosa puoi provare oggi

Il repository contiene il codice DXSpider Mojo e un controllo del prompt Telnet. Puoi [avviare la diagnostica web locale](../dxregia/backend/README.md). Lo script di installazione è in sviluppo.

### 1. Avvia un nodo Mojo di prova

Su una macchina Linux, segui il [manuale di installazione DXSpider Mojo](https://wiki.dxcluster.org/wiki/DXSpider_Installation_Manual_%28New%29) per installare le dipendenze, impostare i nominativi del nodo e del sysop e attivare il servizio. Per le impostazioni di rete e i collegamenti tra nodi consulta la [guida di configurazione DXSpider](https://wiki.dxcluster.org/wiki/Configuration).

Annota l'host raggiungibile dal tuo client e la porta del listener Telnet attivo. Per collegare un nodo partner usa la sua porta effettiva, che può essere diversa dalla porta locale.

### 2. Controlla il listener

Con Git e Python 3 disponibili, puoi eseguire il controllo da un altro dispositivo che deve raggiungere il nodo. Sostituisci HOST e PORTA con i valori del nodo Mojo in prova.

```sh
git clone https://github.com/wifi75/dxregia.git
cd dxregia
python3 dxregia/deploy/check_telnet.py HOST PORTA
```

Se lo script riceve il prompt di accesso, stampa `OK: DXSpider login prompt received`. In caso di errore, controlla indirizzo, porta, stato del servizio e accesso di rete. Il controllo termina al prompt. Per provare login e spot, accedi con un client Telnet o un logger, esegui `show/dx` e verifica separatamente la ricezione di spot in diretta. Usa una rete fidata per le operazioni amministrative via Telnet, che trasmette in chiaro.

## Come funzionerà il pannello DXRegia

L'installatore avvia il pannello anche prima che DXSpider sia configurato. Dopo il cambio password, il sysop completa tutte le impostazioni del nodo dal browser. DXRegia genera e aggiorna i file interni richiesti da Mojo.

1. **Installazione.** Lo script Linux scarica la release, installa le dipendenze, prepara i dati persistenti e verifica l'avvio del servizio. Stampa URL del pannello, utente e password temporanea.
2. **Primo accesso.** Apri l'URL e accedi con le credenziali mostrate. Imposta una nuova password prima di proseguire.
3. **Configurazione del nodo.** Nel wizard inserisci nominativo del nodo, nominativo del sysop, locator e porta Telnet. Controlla il riepilogo e applica la configurazione.
4. **Nodo partner.** Apri Nodi partner → Aggiungi nodo. Compila nominativo, host e porta remota. Usa Verifica indirizzo per controllare DNS e TCP e Salva e connetti per avviare il collegamento. Lo stato segue l'esito comunicato dal motore.
5. **Verifica e backup.** Controlla che il logger riceva gli spot via Telnet e crea un backup dalla sezione di gestione. Dopo un riavvio verifica che configurazione e collegamenti siano recuperati.

[Specifiche dell'installatore](../dxregia/deploy/README.md) · [Gestione dei nodi partner e backup](OPERATIONS.md) · [Stato di implementazione](../TASK.md)
