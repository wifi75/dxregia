# Interfacce DXRegia

## Area pubblica

La pagina Spot è l'ingresso pubblico, consultabile senza login. Il riferimento visivo fornito mostra una tabella ad alta densità, filtri a portata di mano e un riferimento Telnet in testata. DXRegia adotta una propria identità grafica con fondo neutro, accenti teal e tipografia tabellare per frequenze e orari.

- Barra con nome del nodo, stato del feed, orologio UTC e pulsante per copiare indirizzo e porta Telnet.
- Filtri immediati per banda e modo, ricerca per nominativo o testo e azione per azzerare i filtri. I nuovi spot rispettano i filtri attivi.
- Colonne: spotter, frequenza, modo, nominativo DX, entità DXCC, commento e UTC. Il colore della banda affianca il dato testuale. L'entità DXCC può essere assente quando il motore non la determina.
- Aggiornamento in diretta con pausa e contatore degli spot in arrivo. Su mobile, i campi principali restano immediatamente visibili e i dettagli si aprono su richiesta.
- Stati espliciti per caricamento, feed disconnesso, nessuno spot e nessun risultato per i filtri correnti.

## Area sysop

L'area privata usa lo stesso sistema visivo e una navigazione orientata ai compiti: Panoramica, Nodi partner, Utenti, Filtri, Configurazione e Console.

- Al primo accesso con la password temporanea fornita dall'installatore, il sysop deve impostare una nuova password. Poi si apre il wizard.
- Il primo avvio è guidato. Il sysop inserisce nominativi, locator e porta Telnet, controlla il riepilogo e vede l'esito della configurazione.
- La pagina dei partner mostra stato, ultima connessione ed eventuale errore. Aggiunta, test e riconnessione sono azioni visibili nel flusso.
- Quando esiste un comando equivalente, mostrarlo accanto all'esito dell'azione. La console classica rimane disponibile per operazioni avanzate.
- Per le modifiche persistenti mostrare valori correnti, anteprima delle modifiche e se è richiesto un riavvio. Mostrare chiaramente se una modifica è stata applicata oppure è fallita.
- La pagina pubblica legge solo gli spot e le informazioni rese pubbliche dal nodo. Le azioni di gestione e la console richiedono autenticazione sysop.

## Verifiche della prima interfaccia

Provare la pagina spot su desktop e mobile, con tastiera e con feed interrotto. Provare il wizard e l'aggiunta di un partner senza dover aprire la console, compreso un tentativo di connessione fallito.

Le schermate Nodi partner, Registro eventi e Backup seguono i requisiti in [Operazioni del sysop](OPERATIONS.md).
