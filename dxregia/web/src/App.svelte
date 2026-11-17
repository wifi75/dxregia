<script lang="ts">
  import { onMount } from 'svelte'
  import Setup from './Setup.svelte'

  type Status = { address: string; state: 'ready' | 'unexpected' | 'unreachable'; checkedAt: string }
  const labels = {
    ready: ['Listener pronto', 'DXSpider ha risposto con il prompt di accesso.'],
    unexpected: ['Prompt assente', 'La porta risponde, ma il prompt DXSpider non è arrivato.'],
    unreachable: ['Listener irraggiungibile', 'Controlla che DXSpider sia avviato e che l’indirizzo Telnet sia corretto.'],
  }
  let status = $state<Status>()
  let busy = $state(false)
  let failed = $state(false)
  let message = $derived(failed
    ? ['Verifica non disponibile', 'Il pannello non riesce a interrogare il backend. Riprova.']
    : status ? labels[status.state] : ['Verifica in corso', 'Interrogazione del listener Telnet.'])

  async function refresh() {
    if (busy) return
    busy = true
    try {
      const response = await fetch('/api/listener', { signal: AbortSignal.timeout(6000) })
      if (!response.ok) throw new Error('API unavailable')
      status = await response.json() as Status
      failed = false
    } catch {
      failed = true
    } finally {
      busy = false
    }
  }

  function utc(value: string) {
    return new Intl.DateTimeFormat('it-IT', {
      timeZone: 'UTC', hour: '2-digit', minute: '2-digit', second: '2-digit',
    }).format(new Date(value))
  }

  onMount(() => { void refresh() })
</script>

<header>
  <div class="brand">DXRegia</div>
  <span class="section">Diagnostica locale</span>
</header>

<main>
  <div class="heading">
    <h1>Stato del nodo</h1>
    <button onclick={refresh} disabled={busy}>{busy ? 'Verifica in corso…' : 'Verifica ora'}</button>
  </div>
  <article>
    <div class="summary" class:ready={!failed && status?.state === 'ready'} role="status" aria-busy={busy}>
      <span class="signal" aria-hidden="true"></span>
      <div>
        <h2>{message[0]}</h2>
        <p>{message[1]}</p>
      </div>
    </div>
    <dl>
      <div><dt>Listener Telnet</dt><dd>{status?.address ?? '—'}</dd></div>
      <div><dt>Ultima risposta</dt><dd>{status ? utc(status.checkedAt) + ' UTC' : '—'}</dd></div>
    </dl>
  </article>
  <Setup />
</main>

<style>
  :global(*) { box-sizing: border-box; }
  :global(html) { color-scheme: light; }
  :global(body) {
    margin: 0;
    min-width: 280px;
    min-height: 100vh;
    background: #f7f8fa;
    color: #182636;
    font: 15px/1.5 system-ui, sans-serif;
  }
  header {
    display: flex;
    align-items: baseline;
    gap: 18px;
    background: #fff;
    border-bottom: 1px solid #d7dfe7;
    padding: 12px clamp(16px, 4vw, 48px);
  }
  .brand { font-size: 1.15rem; font-weight: 750; color: #143a60; }
  .section, p, dt { color: #485b6d; }
  .section { font-size: 0.85rem; }
  main { max-width: 1000px; margin: auto; padding: 30px clamp(16px, 4vw, 40px); }
  .heading {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }
  h1 { margin: 0; font-size: 1.65rem; line-height: 1.2; }
  button {
    padding: 7px 13px;
    border: 1px solid #214e80;
    border-radius: 4px;
    background: #214e80;
    color: #fff;
    font: inherit;
    font-weight: 650;
    cursor: pointer;
  }
  button:disabled { opacity: 0.72; cursor: wait; }
  button:focus-visible { outline: 2px solid #214e80; outline-offset: 3px; }
  article { background: #fff; border: 1px solid #d7dfe7; }
  .summary { display: flex; align-items: flex-start; gap: 10px; padding: 14px 16px; }
  .signal { margin-top: 8px; width: 8px; height: 8px; flex: 0 0 8px; background: #9a5600; border-radius: 50%; }
  .ready .signal { background: #217346; }
  h2 { margin: 0; font-size: 1rem; }
  p { margin: 3px 0 0; }
  dl {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 185px), 1fr));
    margin: 0;
    border-top: 1px solid #d7dfe7;
    gap: 12px 20px;
    padding: 12px 16px;
  }
  dt { font-size: 0.82rem; }
  dd { margin: 3px 0 0; font: 600 0.9rem ui-monospace, monospace; overflow-wrap: anywhere; }
</style>
