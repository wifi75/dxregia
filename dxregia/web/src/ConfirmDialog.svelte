<script lang="ts">
  // Conferma nello stile del pannello, al posto di window.confirm. Uso: <ConfirmDialog bind:this={confirmer} />
  // e poi `if (!await confirmer.ask(titolo, testo, etichetta, true)) return`.
  let dialog: HTMLDialogElement
  let state = $state({ title: '', text: '', label: '', danger: false })
  let done: ((ok: boolean) => void) | undefined

  export function ask(title: string, text: string, label: string, danger = false) {
    if (dialog.open) return Promise.resolve(false)
    state = { title, text, label, danger }
    dialog.returnValue = ''
    return new Promise<boolean>(resolve => { done = resolve; dialog.showModal() })
  }

  function finish() {
    done?.(dialog.returnValue === 'confirm')
    done = undefined
  }
</script>

<dialog class="glass-panel confirm-dialog" bind:this={dialog} onclose={finish} aria-labelledby="confirm-title" aria-describedby="confirm-text">
  <h2 id="confirm-title">{state.title}</h2>
  <p id="confirm-text" class="confirm-text">{state.text}</p>
  <form method="dialog" class="actions">
    <button class="secondary" value="cancel" autofocus>Annulla</button>
    <button class:danger={state.danger} value="confirm">{state.label}</button>
  </form>
</dialog>

<style>
  .confirm-dialog { width: min(440px, calc(100% - 32px)); box-sizing: border-box; padding: 24px; color: var(--glass-text); }
  .confirm-dialog::backdrop { background: #0b2238a0; backdrop-filter: blur(3px); }
  .confirm-text { white-space: pre-line; }
  .actions { display: flex; flex-wrap: wrap; gap: 10px; justify-content: flex-end; margin-top: 14px; }
</style>
