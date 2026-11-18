<script lang="ts">
  let call = $state('')
  let command = $state('show/dx 10')
  let busy = $state(false)
  let entries = $state<{ command: string; text: string }[]>([])

  async function execute(event: SubmitEvent) {
    event.preventDefault()
    if (busy || !call.trim() || !command.trim()) return
    const submitted = command
    busy = true
    try {
      const response = await fetch('/api/console', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ call, command: submitted }),
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

<section aria-labelledby="console-title">
  <h1 id="console-title">Console Telnet</h1>
  <form onsubmit={execute}>
    <label>Nominativo <input bind:value={call} required maxlength="12" autocomplete="off" spellcheck="false" /></label>
    <label>Comando <input bind:value={command} required maxlength="64" autocomplete="off" spellcheck="false" /></label>
    <button disabled={busy}>{busy ? 'Attendo DXSpider…' : 'Invia comando'}</button>
    <button type="button" class="secondary" onclick={() => entries = []} disabled={busy || !entries.length}>Pulisci</button>
  </form>
  <details>
    <summary>Comandi disponibili</summary>
    <p><code>show/dx 10</code>, <code>show/cluster</code>, <code>show/users</code>, <code>show/links</code>, <code>show/uptime</code>, <code>show/version</code>, <code>show/prefix IU3WJQ</code>, <code>show/qrz IU3WJQ</code></p>
  </details>
  {#if entries.length}
    <div class="transcript" role="log" aria-label="Risposte di DXSpider">
      {#each entries as entry}
        <pre><strong>&gt; {entry.command}</strong>{'\n'}{entry.text}</pre>
      {/each}
    </div>
  {/if}
</section>

<style>
  section { padding: clamp(16px, 2.5vw, 26px); border: 1px solid #b7c9d5; border-radius: 9px; background: #ffffffcf; backdrop-filter: blur(10px); }
  h1 { margin: 0 0 16px; color: #173a51; }
  form { display: flex; flex-wrap: wrap; align-items: end; gap: .8rem; }
  label { display: grid; gap: .3rem; font-weight: 650; }
  input { min-height: 42px; width: min(100%, 260px); border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
  details { margin-top: 16px; }
  summary { cursor: pointer; }
  .transcript { max-height: 58vh; overflow: auto; margin-top: 16px; border: 1px solid #c5d4de; border-radius: 7px; background: #edf4f8; }
  pre { margin: 0; padding: .8rem 1rem; border-bottom: 1px solid #d3e1e9; white-space: pre-wrap; overflow-wrap: anywhere; }
  pre strong { color: #135477; }
</style>
