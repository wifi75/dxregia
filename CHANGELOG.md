# Changelog

Formato [Keep a Changelog](https://keepachangelog.com/it/1.1.0/), versioni [SemVer](https://semver.org/lang/it/). Si riferisce a DXRegia (cartella `dxregia/`); il motore upstream mantiene la propria cronologia in `Changes`.

## [Non rilasciato]

### Aggiunto
- Nuovo logo (icona DX e parola Regia) in alta definizione, favicon e icone per telefono e sito installabile.
- Centro di analytics nel pannello: pubblico del portale (visite, visitatori unici senza cookie, lingua, dispositivi, orari) e attività del cluster (spot per giorno e ora, bande, modi, entità, nominativi).
- Pagina degli spot rifatta: filtri a selezione multipla su una riga, ricerca, lente sui nominativi, nuovi spot evidenziati con suono, modo dedotto dalla frequenza, tema chiaro e scuro, impaginazione per telefono e tablet.
- Spedizioni DX annunciate (feed pubblico NG3K ADXO) con popup accanto al nominativo.
- Portale e email in sei lingue (italiano, inglese, spagnolo, tedesco, francese, portoghese); email HTML ai partner con conferma al primo collegamento.
- Configurazione automatica dei partner dal form Contatti, con interruttore e limite giornaliero.
- Eliminazione di richieste e nodi partner; dashboard live con LED; login con utente e password; pannello su `/admin`.

## [0.2.0] - 2026-11-18

### Aggiunto
- Dashboard dello stato con LED, contatori, spot per banda e ultimi spot.
- Pagina Nodi partner in due schede (uscita e ingresso), con «Abilita e invia l'email» e reinvio delle istruzioni.
- Pannello Gestione → Server di posta con prova di invio.
- Bandiere per tutte le entità DXCC nel cluster web.

### Corretto
- Salvare l'anagrafica non cancella più la password remota del partner.
- Un account Mojo senza password non blocca più il collegamento in uscita.
- Il passo di accesso prima del 2FA dura 20 minuti.

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
