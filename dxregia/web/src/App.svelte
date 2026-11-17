<script lang="ts">
  import { onMount } from 'svelte'

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
  <div class="brand"><span class="mark" aria-hidden="true">DX</span> DXRegia</div>
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
</main>

<style>
  :global(*) { box-sizing: border-box; }
  :global(body) { margin: 0; min-width: 280px; min-height: 100vh; background: #101923; color: #edf4fb; font-family: system-ui, sans-serif; }
  header { min-height: 70px; display: flex; justify-content: space-between; align-items: center; gap: 14px; background: #141f2b; border-bottom: 1px solid #293a4b; padding: 12px clamp(18px, 5vw, 72px); }
  .brand { display: flex; align-items: center; gap: 12px; font-size: 1.18rem; font-weight: 750; }
  .mark { display: grid; place-items: center; width: 38px; height: 38px; background: #83e4cc; color: #112b25; border-radius: 9px; font-size: 0.85rem; letter-spacing: -0.03em; }
  .section, p, dt { color: #a8b9ca; }
  .section { font-size: 0.85rem; }
  main { max-width: 1040px; margin: 0 auto; padding: clamp(30px, 8vw, 84px) clamp(18px, 5vw, 46px); }
  .heading { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 18px; margin-bottom: 26px; }
  h1 { font-size: clamp(2rem, 6vw, 3.4rem); letter-spacing: -0.05em; line-height: 1.04; margin: 0; }
  button { padding: 12px 18px; border: 0; border-radius: 8px; background: #83e4cc; color: #112b25; font: inherit; font-weight: 700; cursor: pointer; }
  button:disabled { opacity: 0.72; cursor: wait; }
  button:focus-visible { outline: 2px solid #fff; outline-offset: 4px; }
  article { background: #172532; border: 1px solid #304358; border-radius: 16px; overflow: hidden; }
  .summary { display: flex; align-items: flex-start; gap: 16px; padding: clamp(20px, 4vw, 32px); }
  .signal { margin-top: 8px; width: 12px; height: 12px; flex: 0 0 12px; background: #ffcb82; border-radius: 50%; box-shadow: 0 0 0 5px #ffcb8220; }
  .ready .signal { background: #83e4cc; box-shadow: 0 0 0 5px #83e4cc20; }
  h2 { margin: 0; font-size: 1.25rem; }
  p { margin: 8px 0 0; line-height: 1.55; }
  dl { display: grid; grid-template-columns: repeat(auto-fit, minmax(185px, 1fr)); margin: 0; border-top: 1px solid #304358; gap: 22px; padding: 22px clamp(20px, 4vw, 32px); }
  dt { font-size: 0.8rem; margin-bottom: 8px; }
  dd { margin: 0; font: 600 0.95rem ui-monospace, monospace; overflow-wrap: anywhere; }
</style>
