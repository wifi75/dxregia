# Registro delle decisioni

## 2026-11-17 — Motore compatibile e interfaccia separata

Riutilizzare Mojo come baseline. Sviluppare Go e Svelte/TypeScript in dxregia/. Conservare la cronologia Git e la disposizione dei file upstream. Valutare sostituzioni del motore per componente, con prove di compatibilità prima di rimuovere i percorsi originali.

## 2026-11-17 — Identità e licenza

DXRegia è il nome di lavoro. Il clone contiene COPYING (Artistic License) e LICENSE (Artistic License 2.0). Conservare entrambi e verificare i termini applicabili prima di distribuire. MIT per il codice originale separato è una proposta da confermare prima di introdurre licenze nei componenti.

## 2026-11-18 — Numerazione delle release

La prima release completa e collaudata sarà 1.0.0, con tag Git v1.0.0. Aggiornare il changelog a ogni pubblicazione. Gli aggiornamenti ordinari successivi incrementano l'ultima cifra: 1.0.1, 1.0.2 e così via. Il tag 1.0.0 si crea quando sono soddisfatti i criteri di accettazione in TASK.md.

## 2026-11-18 — Registrazione degli utenti dal portale

Chi vuole inviare spot si registra dal portale con nominativo, email e password e conferma un link ricevuto via email; solo allora Mojo crea l'utente registrato, con la password salvata come hash bcrypt (`{DXREGIA-BCRYPT}`), e Telnet la chiede. Scelta tra modulo con email, approvazione del sysop e registrazione a mano: modulo con email, perché non richiede interventi e verifica almeno l'indirizzo. L'indirizzo pubblico del sito nei link lo imposta il sysop e non si ricava mai dalla richiesta (altrimenti un'intestazione Host falsa farebbe spedire link verso siti altrui). La funzione richiede l'opzione Mojo «solo utenti registrati dal sysop inviano spot» (reqreg=1); con reqreg=2 il flag registered significa il contrario. Limite noto: la conferma prova il possesso dell'email, non del nominativo. Dati conservati: nominativo, email e data in `.dxregia-registrations.json` (hash della password, mai la password; del link solo l'hash SHA-256), da dichiarare nell'informativa di chi pubblica il nodo.
