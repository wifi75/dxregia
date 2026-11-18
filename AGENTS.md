# Istruzioni di progetto

1. Leggi TASK.md, quindi docs/HANDOFF.md e le decisioni pertinenti in docs/DECISIONS.md.
2. Controlla branch, stato Git e diff prima di editare. Conserva il lavoro già presente.
3. Completa una funzionalità verticale alla volta, dalla UI fino al motore, e registra come è stata verificata.
4. Scrivi il nuovo codice in dxregia/. Il motore upstream sta in engine/ con i suoi percorsi interni invariati, per facilitare aggiornamenti e confronto; COPYING e LICENSE restano nella radice. Mantieni Telnet, output dei logger e protocollo tra nodi compatibili.
5. Usa azioni backend esplicite e parametri validati per i pulsanti. Riserva la console DXSpider ai sysop autenticati. Evita di salvare credenziali nel repository.
6. Conserva avvisi e licenze upstream. Registra data e descrizione nel file upstream modificato, in attesa della verifica di licenza.
7. Al termine aggiorna TASK.md e docs/HANDOFF.md con stato, verifiche, limiti e prossimo passo. Registra in docs/DECISIONS.md solo le decisioni che cambiano il progetto. Git conserva la cronologia delle modifiche.
8. Quando puoi scegliere il modello, usa quello più economico adatto per bozze grafiche e lavori semplici. Riserva i modelli più capaci per protocollo, persistenza e cambiamenti architetturali.
9. Durante il lavoro attivo, circa ogni 20 minuti salva i cambiamenti in un commit significativo, verifica il diff ed esegui il push. Per lavori ancora in collaudo, usa il ramo remoto di verifica. Evita commit vuoti e mantieni fuori da Git credenziali e dati di laboratorio.
