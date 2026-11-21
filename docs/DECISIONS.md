# Registro delle decisioni

## 2026-11-17 — Motore compatibile e interfaccia separata

Riutilizzare Mojo come baseline. Sviluppare Go e Svelte/TypeScript in dxregia/. Conservare la cronologia Git e la disposizione dei file upstream. Valutare sostituzioni del motore per componente, con prove di compatibilità prima di rimuovere i percorsi originali.

## 2026-11-17 — Identità e licenza

DXRegia è il nome di lavoro. Il clone contiene COPYING (Artistic License) e LICENSE (Artistic License 2.0). Conservare entrambi e verificare i termini applicabili prima di distribuire. MIT per il codice originale separato è una proposta da confermare prima di introdurre licenze nei componenti.

## 2026-11-18 — Numerazione delle release

La prima release completa e collaudata sarà 1.0.0, con tag Git v1.0.0. Aggiornare il changelog a ogni pubblicazione. Gli aggiornamenti ordinari successivi incrementano l'ultima cifra: 1.0.1, 1.0.2 e così via. Il tag 1.0.0 si crea quando sono soddisfatti i criteri di accettazione in TASK.md.

## 2026-11-18 — Registrazione degli utenti dal portale

Chi vuole inviare spot si registra dal portale con nominativo, email e password e conferma un link ricevuto via email; solo allora Mojo crea l'utente registrato, con la password salvata come hash bcrypt (`{DXREGIA-BCRYPT}`), e Telnet la chiede. Scelta tra modulo con email, approvazione del sysop e registrazione a mano: modulo con email, perché non richiede interventi e verifica almeno l'indirizzo. L'indirizzo pubblico del sito nei link lo imposta il sysop e non si ricava mai dalla richiesta (altrimenti un'intestazione Host falsa farebbe spedire link verso siti altrui). La funzione richiede l'opzione Mojo «solo utenti registrati dal sysop inviano spot» (reqreg=1); con reqreg=2 il flag registered significa il contrario. Limite noto: la conferma prova il possesso dell'email, non del nominativo. Dati conservati: nominativo, email e data in `.dxregia-registrations.json` (hash della password, mai la password; del link solo l'hash SHA-256), da dichiarare nell'informativa di chi pubblica il nodo.

## 2026-11-19 — Sessioni sysop e statistiche pubbliche

DXRegia conserva la sessione già verificata nel database di stato locale del nodo come file privato con hash del cookie, impronta delle credenziali e scadenza. Il file resta fuori dai backup. Le statistiche degli spot sono pubbliche da mezzanotte UTC. Analytics del portale resta riservato al sysop.

## 2026-11-19 — Conservazione degli eventi e stato pubblico facoltativi

La conservazione automatica è esplicita e disattivata alla prima installazione. Agisce sui soli archivi riconosciuti dei log, del debug e dello storico email. La pagina /stato è disattivata in partenza ed espone solo lo stato osservato del motore e il timestamp dell’ultimo spot. La diagnostica di processi e disco rimane riservata al sysop.


## 2026-11-19 — Allarmi radio locali al browser

Le regole e gli esiti radio vengono salvati nel browser. Il portale usa il feed degli spot già disponibile e il consenso esplicito per le notifiche di sistema. Il riepilogo e le ore di silenzio seguono l’ora locale del dispositivo. Il monitoraggio si svolge mentre il portale è aperto.

## 2026-11-20 — Feed RBN

DXRegia usa il collegamento RBN nativo di DXSpider e pubblica un’istantanea privata dei rapporti CW e RTTY già raggruppati. La pagina web legge quella istantanea e mette il feed in una vista dedicata. Il gestore riutilizza un collegamento RBN in uscita sulla porta 7000 quando è già presente. La gestione Telnet resta affidata alle preferenze di ogni utente DXSpider.

## 2026-11-21 — 2FA ricordata per 10 giorni

La sessione completa persisteva già al riavvio, ma il login con password la cancellava e il cookie di sessione moriva con il browser: il codice veniva richiesto a ogni accesso. Un cookie persistente «dispositivo fidato» (10 giorni, solo hash su disco in .dxregia-trusted.json, legato a password e chiave 2FA) salta il codice al login con password giusta. Scartato: allungare solo la sessione, perché non copre il login dopo la chiusura del browser. Il cambio password o 2FA invalida ogni dispositivo.

## 2026-11-21 — Aggiornamento dal pannello tramite unità systemd

Il servizio web gira senza privilegi e viene riavviato dall'aggiornamento, quindi non può eseguirlo. Il pannello scrive solo un file di richiesta vuoto; una unità `.path` avvia un servizio `oneshot` che esegue lo script ufficiale come root. Scartati: sudo o polkit per l'utente del servizio (superficie d'attacco più ampia) e un aggiornamento eseguito dentro il processo web (sarebbe interrotto dal riavvio). Nessun URL, versione o argomento arriva dal browser.
