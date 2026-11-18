# Interfacce DXRegia

## Direzione visiva

DXRegia usa una sola identità visiva per pagina pubblica e area sysop. Entrambe le aree usano esclusivamente un tema chiaro, con fondo bianco, separatori grigio chiaro e blu per i comandi. Il layout riprende la densità di uno strumento radio: barra operativa compatta, titoli sobri, tabelle e moduli ben allineati. Evitare gradienti, effetti vetro, grandi riquadri decorativi e titoli promozionali. Ambra e rosso segnalano stati operativi e sono accompagnati da testo. Frequenze e orari usano cifre tabellari per rimanere allineati mentre arrivano nuovi dati.

- **Spot pubblici.** La tabella è il centro della pagina. Ricerca, filtri e pausa del feed sono raggiungibili sopra le righe. Frequenza e modo si distinguono a colpo d'occhio, con aggiornamenti che preservano la posizione di lettura.
- **Area sysop.** Navigazione laterale su desktop e menu compatto su mobile. Nodi partner apre una vista con elenco e dettaglio. Selezionando un nodo si vedono stato, ultima attività ed eventi recenti, con le azioni principali vicino al nominativo.
- **Wizard.** Una decisione per schermata, errori accanto al campo, riepilogo prima dell'applicazione e verifica del risultato nella stessa pagina.
- **Dettagli e accessibilità.** Gerarchia tipografica coerente, focus visibile, contrasto leggibile e stati di caricamento, errore, assenza dati e riconnessione. Le animazioni rispondono alle azioni e rispettano la preferenza di movimento ridotto.

Prima di estendere il pannello, verificare due schermate con dati rappresentativi: Spot pubblici e Nodi partner con un errore di connessione. Esaminarle a 390 e 1440 pixel, nel tema chiaro e con navigazione da tastiera. Per valutare il layout usare nominativi lunghi, commenti estesi e almeno un aggiornamento mentre l'operatore legge una riga.

## Marchio e icone

Il marchio è l'icona **DX** (lettere D bianca e X ambra su fondo blu notte) seguita dalla parola **Regia**. I file sono in `dxregia/web/public/brand/` e sono serviti dal sito.

| File | Uso |
| --- | --- |
| `dxregia-logo.svg`, `dxregia-logo-2000.png` | Logo completo per sfondi chiari |
| `dxregia-logo-dark.svg`, `dxregia-logo-dark-2000.png` | Logo completo per sfondi scuri (testo bianco) |
| `dxregia-wordmark.svg`, `dxregia-wordmark-dark.svg` | Solo la parola «Regia», usata nella pagina di accesso sotto l'icona animata |
| `dxregia-icon.svg`, `dxregia-icon-1024.png`, `-512.png`, `-256.png` | Solo l'icona, in alta definizione |

La parola è convertita in tracciati vettoriali (carattere Inter ExtraBold, licenza OFL): gli SVG non dipendono da caratteri installati. Colori: blu notte `#143a60`, ambra `#f0a23a`, azzurro `#3a82c0`, bianco. Lascia attorno al logo uno spazio pari almeno a metà dell'altezza dell'icona e non deformarlo né ricolorarlo.

Nella radice di `dxregia/web/public/` ci sono le icone del sito: `favicon.svg`, `favicon.ico` (16, 32 e 48 px), `apple-touch-icon.png` (180 px), `icon-192.png`, `icon-512.png` e `site.webmanifest`, collegati da `index.html`. L'intestazione sceglie il logo chiaro o scuro in base al tema.

## Area pubblica

La pagina Spot è l'ingresso pubblico, consultabile senza login. Il riferimento visivo fornito mostra una tabella ad alta densità, filtri a portata di mano e un riferimento Telnet in testata. DXRegia adotta una propria identità grafica con fondo bianco, accenti blu e tipografia tabellare per frequenze e orari.

- Barra con nome del nodo, stato del feed, orologio UTC e pulsante per copiare indirizzo e porta Telnet.
- Filtri immediati per banda e modo, ricerca per nominativo o testo e azione per azzerare i filtri. I nuovi spot rispettano i filtri attivi.
- Colonne: spotter, frequenza, modo, nominativo DX, entità DXCC, commento e UTC. Il colore della banda affianca il dato testuale. L'entità DXCC può essere assente quando il motore non la determina.
- Aggiornamento in diretta con pausa e contatore degli spot in arrivo. Su mobile, i campi principali restano immediatamente visibili e i dettagli si aprono su richiesta.
- Stati espliciti per caricamento, feed disconnesso, nessuno spot e nessun risultato per i filtri correnti.

## Area sysop

L'area privata usa lo stesso sistema visivo e una navigazione orientata ai compiti: Panoramica, Nodi partner, Utenti, Filtri, Configurazione e Console.

- Al primo accesso con la password temporanea fornita dall'installatore, il sysop deve impostare una nuova password. Poi si apre il wizard.
- Tutta la prima configurazione avviene nel browser, anche quando il motore è ancora spento. Il sysop inserisce nominativi, locator e porta Telnet, controlla il riepilogo e vede l'esito della configurazione.
- La pagina dei partner mostra stato, ultima connessione ed eventuale errore. Aggiunta, test e riconnessione sono azioni visibili nel flusso.
- Quando esiste un comando equivalente, mostrarlo accanto all'esito dell'azione. La console classica rimane disponibile per operazioni avanzate.
- Per le modifiche persistenti mostrare valori correnti, anteprima delle modifiche e se è richiesto un riavvio. Mostrare chiaramente se una modifica è stata applicata oppure è fallita.
- La pagina pubblica legge solo gli spot e le informazioni rese pubbliche dal nodo. Le azioni di gestione e la console richiedono autenticazione sysop.

## Terminale Telnet

La console classica deve essere raggiungibile anche da un client Telnet esterno, sulla porta TCP configurata per DXSpider. La pagina pubblica mostra host e porta. La sessione supporta login per nominativo, prompt e comandi DXSpider. Per le operazioni amministrative valgono i privilegi del sysop.

Migliorare l'uso interattivo con aiuto consultabile, messaggi chiari e, nella console web, cronologia e completamento dei comandi. Per eventuali estensioni Telnet usare impostazioni esplicite e verificare che login, prompt, terminatori di riga, formato spot e output dei comandi continuino a funzionare con i logger. L'amministrazione via Telnet va prevista su LAN fidata o VPN, perché Telnet trasmette in chiaro.

Verificare con un client Telnet esterno e un programma di logging: login, show/dx, ricezione di spot in diretta e controllo dei privilegi sysop. Eseguire un comando dalla console web e dalla sessione classica per verificare che entrambe interagiscano con lo stesso motore.

## Verifiche della prima interfaccia

Provare la pagina spot su desktop e mobile, con tastiera e con feed interrotto. Provare il wizard e l'aggiunta di un partner senza dover aprire la console, compreso un tentativo di connessione fallito.

Le schermate Nodi partner, Registro eventi e Backup seguono i requisiti in [Operazioni del sysop](OPERATIONS.md).

## Riferimento approvato per gli spot pubblici

L'anteprima del 17 novembre 2026 definisce un portale in tema chiaro con testata compatta, filtri banda a pulsante, ricerca e filtro modo. La tabella alterna righe bianche e azzurrate, usa bandierine accanto al nome dell'entità, codifica la banda nel colore della frequenza e distingue CW, SSB e FT8 con etichette testuali colorate. Usare la skill design-taste-frontend per l'implementazione delle interfacce. L'approvazione riguarda la pagina pubblica.
