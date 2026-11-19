<script lang="ts">
  import { onDestroy, tick } from 'svelte'

  const commands = [
    { name: 'show/dx', args: '[numero]', help: 'Mostra gli ultimi spot. Numero facoltativo da 1 a 99.', example: 'show/dx 10' },
    { name: 'show/cluster', args: '', help: 'Mostra il riepilogo del cluster.', example: 'show/cluster' },
    { name: 'show/users', args: '', help: 'Elenca gli utenti collegati.', example: 'show/users' },
    { name: 'show/links', args: '', help: 'Elenca i collegamenti del nodo.', example: 'show/links' },
    { name: 'show/uptime', args: '', help: 'Mostra da quanto tempo il nodo è in funzione.', example: 'show/uptime' },
    { name: 'show/version', args: '', help: 'Mostra la versione del motore DXSpider.', example: 'show/version' },
    { name: 'show/prefix', args: 'NOMINATIVO', help: 'Ricava le informazioni DXCC dal prefisso o nominativo.', example: 'show/prefix IU3WJQ' },
    { name: 'show/qrz', args: 'NOMINATIVO', help: 'Cerca un nominativo tramite QRZ, se configurato sul nodo.', example: 'show/qrz IU3WJQ' },
  ]

  const endpoint = '/api/console/session'
  let call = $state('')
  let password = $state('')
  let command = $state('')
  let session = $state('')
  let busy = $state(false)
  let error = $state('')
  let output = $state<string[]>([])
  let spots = $state<string[]>([])
  let history = $state<string[]>([])
  let historyIndex = 0
  let draft = ''
  let disposed = false
  let pollTimer: ReturnType<typeof setTimeout>
  let transcript: HTMLDivElement
  let spotPanel: HTMLDivElement
  let commandInput: HTMLInputElement
  let commandName = $derived(command.trimStart().split(/\s/)[0].toLowerCase())
  let suggestions = $derived(commands.filter(item => item.name.startsWith(commandName)))
  let help = $derived(commands.find(item => item.name === commandName))

  function request(method: string, id: string, data?: object) {
    return fetch(endpoint, {
      method,
      headers: { 'Content-Type': 'application/json', ...(id ? { 'X-Console-Session': id } : {}) },
      body: data ? JSON.stringify(data) : undefined,
      keepalive: method === 'DELETE',
    })
  }

  async function append(responses: string[], incoming: string[] = []) {
    const follows = (panel: HTMLDivElement) => panel.scrollHeight - panel.scrollTop - panel.clientHeight < 30
    const followOutput = follows(transcript)
    const followSpots = follows(spotPanel)
    if (responses.length) output = [...output, ...responses].slice(-600)
    if (incoming.length) spots = [...spots, ...incoming].slice(-300)
    await tick()
    if (followOutput && transcript) transcript.scrollTop = transcript.scrollHeight
    if (followSpots && spotPanel) spotPanel.scrollTop = spotPanel.scrollHeight
  }

  async function poll(id: string) {
    try {
      const response = await request('GET', id)
      if (session !== id || disposed) return
      if (!response.ok) {
        session = ''
        error = (await response.text()).trim()
        return
      }
      const { Lines, Closed, Dropped } = await response.json() as {
        Lines: { Text: string; Spot: boolean }[] | null; Closed: boolean; Dropped: boolean
      }
      if (session !== id || disposed) return
      const lines = Lines ?? []
      const responses = lines.filter(line => !line.Spot).map(line => line.Text)
      if (Dropped) responses.unshift('Parte dell’output precedente è stata scartata.')
      await append(responses, lines.filter(line => line.Spot).map(line => line.Text))
      if (session !== id || disposed) return
      if (Closed) {
        session = ''
        error = 'DXSpider ha chiuso la sessione Telnet.'
        return
      }
      error = ''
    } catch {
      if (session === id && !disposed) error = 'Console non raggiungibile. Nuovo controllo in corso…'
    }
    if (session === id && !disposed) pollTimer = setTimeout(() => void poll(id), 800)
  }

  async function connect(event: SubmitEvent) {
    event.preventDefault()
    if (busy || session || !call.trim()) return
    busy = true
    error = ''
    try {
      const response = await request('POST', '', { call, password })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const { ID } = await response.json()
      if (disposed) {
        await request('DELETE', ID)
        return
      }
      session = ID
      void poll(ID)
      await tick()
      commandInput?.focus()
    } catch {
      if (!disposed) error = 'Console non raggiungibile.'
    } finally {
      password = ''
      busy = false
    }
  }

  function disconnect() {
    clearTimeout(pollTimer)
    const id = session
    session = ''
    if (id) void request('DELETE', id).catch(() => {})
  }

  onDestroy(() => {
    disposed = true
    disconnect()
  })

  async function execute(event: SubmitEvent) {
    event.preventDefault()
    if (!session || busy || !command.trim()) return
    const submitted = command.trim()
    const id = session
    busy = true
    error = ''
    void append(['> ' + submitted])
    try {
      const response = await request('POST', id, { command: submitted })
      if (!response.ok) {
        await append([(await response.text()).trim()])
        return
      }
      if (history.at(-1) !== submitted) history = [...history, submitted].slice(-100)
      historyIndex = history.length
      command = draft = ''
    } catch {
      await append(['Invio non confermato.'])
    } finally {
      busy = false
    }
  }

  function complete(item: (typeof commands)[number]) {
    command = item.args ? item.name + ' ' : item.name
    commandInput?.focus()
  }

  function commandKeys(event: KeyboardEvent) {
    if (event.key === 'Tab' && commandName && suggestions.length === 1 && !command.includes(' ') && command !== suggestions[0].name) {
      event.preventDefault()
      complete(suggestions[0])
      return
    }
    if ((event.key !== 'ArrowUp' && event.key !== 'ArrowDown') || !history.length) return
    event.preventDefault()
    if (historyIndex === history.length) draft = command
    historyIndex = Math.max(0, Math.min(history.length, historyIndex + (event.key === 'ArrowUp' ? -1 : 1)))
    command = history[historyIndex] ?? draft
  }
</script>

<section class="glass-panel" aria-labelledby="console-title">
  <h1 id="console-title">Console Telnet</h1>
  <form class="credentials" onsubmit={connect}>
    <label>Nominativo <input bind:value={call} required maxlength="12" autocomplete="off" spellcheck="false" disabled={!!session || busy} /></label>
    <label>Password Telnet (se richiesta) <input type="password" bind:value={password} maxlength="72" autocomplete="off" disabled={!!session || busy} /></label>
    {#if session}
      <button type="button" class="secondary" onclick={disconnect} disabled={busy}>Disconnetti</button>
    {:else}
      <button disabled={busy}>Connetti</button>
    {/if}
    <span class="connection" data-online={!!session} role="status">{session ? 'Connesso' : busy ? 'Connessione in corso…' : 'Disconnesso'}</span>
  </form>
  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="panels">
    <div class="terminal">
      <div class="terminal-bar">
        <strong>Risposte ai comandi</strong>
        <button type="button" class="secondary" onclick={() => output = []} disabled={!output.length}>Pulisci</button>
      </div>
      <div class="transcript" bind:this={transcript} role="log" tabindex="0" aria-label="Risposte di DXSpider">
        {#if !output.length}<p>Collegati e invia un comando.</p>{/if}
        <pre>{output.join('\n')}</pre>
      </div>
      <form class="command-line" onsubmit={execute}>
        <label for="telnet-command" aria-label="Comando Telnet">&gt;</label>
        <input id="telnet-command" bind:this={commandInput} bind:value={command} onkeydown={commandKeys} maxlength="64" autocomplete="off" spellcheck="false" placeholder="show/dx 10" disabled={!session} aria-describedby="console-command-help" />
        <button disabled={!session || busy || !command.trim()}>Invia comando</button>
      </form>
    </div>
    <div class="terminal spot-terminal">
      <div class="terminal-bar">
        <strong>Spot in arrivo</strong>
        <button type="button" class="secondary" onclick={() => spots = []} disabled={!spots.length}>Pulisci</button>
      </div>
      <div class="transcript" bind:this={spotPanel} role="region" tabindex="0" aria-label="Spot in arrivo">
        {#if !spots.length}<p>Gli spot ricevuti durante la sessione appariranno qui.</p>{/if}
        <pre>{spots.join('\n')}</pre>
      </div>
    </div>
  </div>
  <p id="console-command-help">
    {#if help}<code>{help.name} {help.args}</code> · {help.help}
    {:else}Tab completa un comando se la corrispondenza è unica. ↑ e ↓ scorrono la cronologia.{/if}
  </p>
  {#if session && commandName && !help && suggestions.length}
    <div class="suggestions" aria-label="Completamento comandi">
      {#each suggestions as item (item.name)}
        <button type="button" class="secondary" onclick={() => complete(item)} title={item.help}>{item.name} {item.args}</button>
      {/each}
    </div>
  {/if}
  <details>
    <summary>Comandi disponibili</summary>
    <dl>
      {#each commands as item (item.name)}
        <dt><code>{item.name} {item.args}</code></dt>
        <dd>{item.help} Esempio <code>{item.example}</code></dd>
      {/each}
    </dl>
  </details>
</section>

<style>
  section { padding: clamp(16px, 2.5vw, 26px); }
  h1 { margin: 0 0 16px; }
  .credentials { display: flex; flex-wrap: wrap; align-items: end; gap: .8rem; margin-bottom: 16px; }
  .credentials label { display: grid; gap: .3rem; font-weight: 650; }
  .connection { align-self: end; padding: .6rem .2rem; font-weight: 650; }
  .connection[data-online='true'] { color: #14875d; }
  .error { color: #bb2639; }
  input { min-height: 42px; width: min(100%, 260px); border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
  .panels { display: grid; grid-template-columns: minmax(0, 3fr) minmax(0, 2fr); gap: 14px; }
  .terminal { display: flex; flex-direction: column; min-width: 0; overflow: hidden; border: 1px solid #38516a; border-radius: 12px; background: #101d2a; color: #e4eff8; }
  .terminal-bar, .command-line { display: flex; align-items: center; gap: .7rem; padding: .65rem .9rem; background: #1c3043; }
  .terminal-bar { justify-content: space-between; }
  .terminal-bar button { min-height: 34px; padding: .2rem .7rem; }
  .transcript { height: clamp(240px, 46vh, 530px); overflow: auto; padding: .7rem .9rem; }
  .spot-terminal .transcript { flex: 1; min-height: 0; }
  .transcript p { margin: 0; color: #a8bed1; }
  .command-line label, .command-line input, pre { font-family: ui-monospace, monospace; }
  .command-line input { flex: 1; min-width: 0; width: auto; }
  .command-line button { flex-shrink: 0; }
  pre { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.5; }
  .spot-terminal pre { color: #8de1b8; }
  .suggestions { display: flex; flex-wrap: wrap; gap: .45rem; margin: .6rem 0; }
  summary { cursor: pointer; }
  dt { font-weight: 650; }
  dd { margin: 0 0 .6rem; }
  @media (max-width: 900px) {
    .panels { grid-template-columns: 1fr; }
    .spot-terminal .transcript { flex: initial; height: 220px; }
  }
  @media (max-width: 480px) {
    .command-line { flex-wrap: wrap; }
    .command-line input { flex-basis: 75%; }
  }
</style>
