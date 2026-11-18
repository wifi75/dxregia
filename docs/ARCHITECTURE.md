# Architettura

- **Motore Perl** — DXSpider Mojo serve Telnet e parla con i nodi partner. Sta nella cartella `engine/` con i percorsi interni della base upstream; gli installatori ne copiano solo `cmd`, `connect`, `data`, `filter`, `msg` e `perl`.
- **Backend Go** — espone API di gestione al pannello e usa un adattatore DXSpider locale per leggere lo stato e applicare azioni. La modalità concreta di collegamento alla console va verificata sulla baseline.
- **Frontend Svelte e TypeScript** — pagina spot pubblica e area sysop autenticata con wizard, stato reale del nodo, partner e console. Vedi [Interfacce](INTERFACES.md). La build di produzione può essere servita da Go.
- **Configurazione** — il backend distingue modifiche immediate, persistenti e soggette a riavvio. Le configurazioni preesistenti e le modifiche manuali devono essere conservate durante l'importazione.
- **Distribuzione Linux** — prima immagine OCI per amd64 e arm64. I dati del nodo devono sopravvivere alla sostituzione del container. Aggiungere supporto nativo in base alle distro effettivamente provate.

Flusso esemplificativo: il pannello chiede lo stato dei partner al backend, l'adattatore interroga il motore e il pannello visualizza la risposta, inclusi eventuali errori.
