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

Per ricreare una copia di laboratorio su un altro computer, da una directory esterna al repository DXRegia:

~~~sh
git clone --branch mojo https://github.com/EA3CV/dxspider.git mojo-lab
git -C mojo-lab checkout --detach 9b55f03ff870d50a9d3e57090dc7408035a4a2b1
~~~

Compila il frontend con i comandi precedenti. Poi, da `dxregia/backend` della copia DXRegia (sostituisci i percorsi di esempio):

~~~sh
go run ./cmd/dxregia -port 8082 -telnet 127.0.0.1:7474 \
  -setup-root /percorso/mojo-lab -setup-perl /percorso/perl \
  -setup-internal-port 27756
~~~

Apri http://127.0.0.1:8082, accedi con la password temporanea del terminale e cambiala e applica la configurazione con una porta libera. Su Windows esegui il comando su una sola riga con i percorsi Windows. Le dipendenze richieste da Mojo vanno predisposte prima di lanciare questa anteprima. L'avvio gestito è stato provato localmente su Windows. Il percorso Linux e l'installatore restano da collaudare.

Per mostrare il modulo, avvia il backend con -setup-root seguito dal percorso di una copia nuova di Mojo. Accedi con la password temporanea stampata dal processo e cambiala prima di aprire il wizard. Il modulo salva DXVars.pm e Listeners.pm nella copia scelta e si disattiva dopo il salvataggio. Il listener generato è limitato al loopback. Per inizializzare la banca utenti e avviare Mojo, aggiungi -setup-perl seguito dall'eseguibile Perl con le dipendenze Mojo disponibili. Se fai girare più istanze sullo stesso PC, assegna una porta interna libera con -setup-internal-port. Il pannello conferma l'avvio dopo la risposta Telnet. Il registro di avvio è in local_data/dxregia-startup.log. Al riavvio con gli stessi argomenti, il backend recupera la configurazione che ha generato, riavvia Mojo sulla porta salvata e riusa la banca utenti. Se l'avvio fallisce, il pannello permette di riprovare con i dati salvati. Le installazioni preesistenti prive dello stato DXRegia continuano a essere gestite separatamente.
