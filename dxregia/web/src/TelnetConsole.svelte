<script lang="ts">
  let call = $state('')
  let password = $state('')
  let command = $state('show/dx 10')
  let busy = $state(false)
  let transcript: HTMLDivElement
  let entries = $state<{ command: string; text: string }[]>([])

  $effect(() => {
    if (entries.length && transcript) transcript.scrollTop = transcript.scrollHeight
  })

  async function execute(event: SubmitEvent) {
    event.preventDefault()
    if (busy || !call.trim() || !command.trim()) return
    const submitted = command
    busy = true
    try {
      const response = await fetch('/api/console', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ call, command: submitted, password }),
      })
      const text = response.ok ? (await response.json()).output : (await response.text()).trim()
      entries = [...entries.slice(-19), { command: submitted, text }]
    } catch {
      entries = [...entries.slice(-19), { command: submitted, text: 'Console non raggiungibile.' }]
    } finally {
      busy = false
    }
  }
</script>

<section class="glass-panel" aria-labelledby="console-title">
  <h1 id="console-title">Console Telnet</h1>
  <form onsubmit={execute}>
    <div class="credentials">
      <label>Nominativo <input bind:value={call} required maxlength="12" autocomplete="off" spellcheck="false" /></label>
      <label>Password Telnet (se richiesta) <input type="password" bind:value={password} maxlength="72" autocomplete="off" /></label>
    </div>
    <div class="terminal">
      <div class="terminal-bar">
        <strong>Sessione DXSpider</strong>
        <button type="button" class="secondary" onclick={() => entries = []} disabled={busy || !entries.length}>Pulisci</button>
      </div>
      <div class="transcript" bind:this={transcript} role="log" aria-label="Risposte di DXSpider" aria-busy={busy}>
        {#if !entries.length}<p>Inserisci il nominativo e invia un comando per vedere la risposta di DXSpider.</p>{/if}
        {#each entries as entry}
          <pre><strong>&gt; {entry.command}</strong>{'\n'}{entry.text}</pre>
        {/each}
      </div>
      <div class="command-line">
        <label for="telnet-command" aria-label="Comando Telnet">&gt;</label>
        <input id="telnet-command" bind:value={command} required maxlength="64" autocomplete="off" spellcheck="false" placeholder="show/dx 10" />
        <button disabled={busy}>{busy ? 'Attendo DXSpider…' : 'Invia comando'}</button>
      </div>
    </div>
  </form>
  <details>
    <summary>Comandi disponibili</summary>
    <p><code>show/dx 10</code>, <code>show/cluster</code>, <code>show/users</code>, <code>show/links</code>, <code>show/uptime</code>, <code>show/version</code>, <code>show/prefix IU3WJQ</code>, <code>show/qrz IU3WJQ</code></p>
  </details>
</section>

<style>
  section { padding: clamp(16px, 2.5vw, 26px); }
  h1 { margin: 0 0 16px; color: #173a51; }
  form { display: grid; gap: 16px; }
  .credentials { display: flex; flex-wrap: wrap; gap: .8rem; }
  .credentials label { display: grid; gap: .3rem; font-weight: 650; }
  input { min-height: 42px; width: min(100%, 260px); border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
  details { margin-top: 16px; }
  summary { cursor: pointer; }
  .terminal { min-width: 0; overflow: hidden; border: 1px solid #a3bfce; border-radius: 12px; background: #f7fbfdc9; }
  .terminal-bar, .command-line { display: flex; align-items: center; gap: .8rem; padding: .65rem .9rem; background: #d5eaf1a6; }
  .terminal-bar { justify-content: space-between; border-bottom: 1px solid #c4d9e4; }
  .terminal-bar button { min-height: 34px; padding: .2rem .7rem; }
  .transcript { min-height: 270px; max-height: 55vh; overflow: auto; padding: .55rem 0; }
  .transcript p { margin: .6rem 1rem; color: #476778; }
  .command-line { border-top: 1px solid #c4d9e4; }
  .command-line label, .command-line input, pre { font-family: ui-monospace, monospace; }
  .command-line input { flex: 1; min-width: 0; width: auto; }
  .command-line button { flex-shrink: 0; }
  pre { margin: 0; padding: .65rem 1rem; border-bottom: 1px solid #d3e1e9; white-space: pre-wrap; overflow-wrap: anywhere; }
  pre strong { color: #135477; }
  @media (max-width: 480px) { .command-line { flex-wrap: wrap; } .command-line input { flex-basis: 75%; } }
</style>
