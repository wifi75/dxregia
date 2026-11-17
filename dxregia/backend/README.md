# Backend Go

La diagnostica controlla il listener DXSpider configurato all'avvio e mostra l'esito nel browser. Usa Go 1.22 o successivo e Node.js 22.12 o successivo con npm.

Dalla radice del repository:

~~~sh
cd dxregia/web
npm ci
npm run build
cd ../backend
go run ./cmd/dxregia -telnet 127.0.0.1:7300
~~~

Apri http://127.0.0.1:8080. Usa -port per cambiare la porta HTTP e -telnet HOST:PORTA per impostare il listener effettivo. L'anteprima HTTP ascolta solo sul loopback. Avvia DXSpider separatamente.

## Configurazione iniziale locale

Per mostrare il modulo, avvia il backend con -setup-root seguito dal percorso di una copia nuova di Mojo. Usa il codice temporaneo stampato dal processo. Il modulo salva DXVars.pm e Listeners.pm nella copia scelta e si disattiva dopo il salvataggio. Il listener generato è limitato al loopback. Per inizializzare la banca utenti e avviare Mojo, aggiungi -setup-perl seguito dall'eseguibile Perl con le dipendenze Mojo disponibili. Se fai girare più istanze sullo stesso PC, assegna una porta interna libera con -setup-internal-port. Il pannello conferma l'avvio dopo la risposta Telnet. Il registro di avvio è in local_data/dxregia-startup.log. Il backend non assume ancora la gestione del riavvio di installazioni già configurate.
