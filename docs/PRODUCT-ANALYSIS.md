# Analisi prodotto DXRegia

Analisi del sorgente DXSpider Mojo al commit 9b55f03f, della documentazione sysop e dei progetti esistenti. Le prove di interoperabilità restano da condurre su Linux.

| Priorità | Evidenza in DXSpider | Miglioria DXRegia | Prova |
| --- | --- | --- | --- |
| P0 — Primo avvio | cluster.pl richiede local/DXVars.pm. Il listener è definito in perl/Listeners.pm. Lo script EA3CV automatizza già l'installazione. | Un installatore con verifica del servizio, URL, credenziale temporanea e wizard. Il pannello distingue porta HTTP e porta Telnet. | Da Linux pulito, aprire il pannello dalla LAN e completare il cambio password. Da un secondo dispositivo verificare il prompt Telnet. |
| P0 — Nodi partner | cmd/connect.pl richiede uno script e restituisce l'avvio del tentativo. ExtMsg.pm gestisce connessione e login. Il sysop consulta links e debug per seguire l'esito. | Stato osservato, cronologia per nodo, errore correlato e prossimo tentativo. Mostrare lo stato sconosciuto quando non si può verificare l'esito. | In un laboratorio a due nodi osservare avvio, connessione, scambio spot, caduta e ripresa. Provare porta chiusa e timeout. |
| P0 — Migrazione e recupero | La guida al backup usa copie dei file. Mojo prevede un export della banca utenti. I contributi EA3CV includono backup e recupero degli aggiornamenti. | Importazione con anteprima, inventario dei dati e delle personalizzazioni, backup con manifest e ripristino verificato. Conservare gli script personalizzati e segnalarli per revisione. | Ripristinare su una seconda installazione e confrontare utenti, filtri e partner. |
| P1 — Filtri | Mojo offre regole di ingresso e uscita, per utente, nodo e profili predefiniti. Filter.pm e DXProt.pm applicano regole in momenti diversi. | Editor con precedenza leggibile e simulazione sullo storico. Mostrare la regola applicata dove osservabile. | Ripetere gli stessi spot nel laboratorio e confrontare la simulazione con le decisioni del motore. |
| P1 — Diagnostica | I livelli di debug e gli strumenti di ricerca esistono già. La diagnosi per nodo richiede di ricomporre log e comandi. | Registro ricercabile per categoria, con correlazione tra azione, collegamento ed esito. Debug su richiesta, con scadenza. | Partendo da un singolo evento, trovare tentativo, causa osservata e ripresa del collegamento senza consultare file a mano. |

[Installatore EA3CV](https://github.com/EA3CV/dxspider_update) · [Debug DXSpider](https://wiki.dxcluster.org/wiki/DXSpider_Tracing_Debug_Entries) · [Filtri e routing](https://wiki.dxcluster.org/wiki/Routing_and_Filtering) · [Backup](https://wiki.dxcluster.org/wiki/Backup) · [Ripristino banca utenti Mojo](https://wiki.dxcluster.org/wiki/Restoring_the_user_DB_%28Mojo_branch%29)

## Cosa può distinguere DXRegia

[pyCluster](https://github.com/AI3I/pyCluster) offre già Telnet, interfaccia pubblica e console sysop. [AR-Cluster](https://arcluster.org/index.html) offre strumenti di installazione, recupero e diagnostica. La proposta da verificare con sysop reali è un percorso completo su Linux: installare o migrare un nodo Mojo, capire una caduta di collegamento e recuperare la configurazione dal pannello.

Per i primi collaudi coinvolgere un sysop che parte da zero, uno con un nodo personalizzato e un utente con un logger già configurato. Misurare tempo al primo spot, passaggi manuali, successo della diagnosi e del ripristino. La scelta di sostituire componenti del motore dipenderà da prove di compatibilità.
