# Changelog

Formato [Keep a Changelog](https://keepachangelog.com/it/1.1.0/), versioni [SemVer](https://semver.org/lang/it/). Si riferisce a DXRegia (cartella `dxregia/`); il motore upstream mantiene la propria cronologia in `Changes`.

## [0.1.0] - 2026-11-18

Prima versione pubblicata, in fase di collaudo.

### Aggiunto
- Pagina pubblica degli spot con filtri, nomi DXCC e layout mobile.
- Area sysop con accesso protetto, cambio obbligatorio della password temporanea e conferma TOTP.
- Wizard di primo avvio: configura, inizializza e avvia Mojo da browser.
- Gestione dei nodi partner: stato in tempo reale, password remota, anteprima dello script di peering, provisioning degli account Mojo e riconnessione automatica.
- Console Telnet locale autenticata e opzioni Telnet gestite dal pannello.
- Installazione nativa Linux e Docker, con aggiornamento tramite `dxregia-update`.

### Modificato
- La password remota del partner è facoltativa al salvataggio; disattivare la riconnessione conserva la password.

### Limiti noti
- Build e test Go dell'ultima modifica non eseguiti su questo host (Go non installato).
- Collaudo con due nodi, feed DX e aggiornamento da release ancora in corso; vedi TASK.md.
