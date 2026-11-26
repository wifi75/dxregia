<script lang="ts">
  import { onMount } from 'svelte'

  type Event = { At: number; Kind: string; IP: string; Path: string }
  type Report = { Https: boolean; TwoFactor: boolean; SessionHours: number; Started: number; RateLimited: number; CrossSite: number; LoginFails: number; Events: Event[] | null }
  let report = $state<Report>()
  let failed = $state(false)
  const time = new Intl.DateTimeFormat('it-IT', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
  const since = new Intl.DateTimeFormat('it-IT', { dateStyle: 'medium', timeStyle: 'short' })

  async function load() {
    try {
      const response = await fetch('/api/security', { cache: 'no-store', signal: AbortSignal.timeout(6000) })
      if (!response.ok) throw new Error()
      report = await response.json()
      failed = false
    } catch { failed = true }
  }
  onMount(() => {
    void load()
    const timer = setInterval(load, 10000)
    return () => clearInterval(timer)
  })

  const checks = $derived(report ? [
    { ok: report.Https, label: 'Connessione cifrata (HTTPS)', detail: report.Https ? 'Stai usando HTTPS: cookie protetti e HSTS attivo.' : 'Questa pagina è aperta senza HTTPS. Metti il sito dietro un proxy con certificato (Cloudflare, nginx) prima di esporlo.' },
    { ok: report.TwoFactor, label: 'Verifica in due passaggi (2FA)', detail: report.TwoFactor ? 'Attiva per l’accesso sysop.' : 'Non attiva: attivala da Gestione > Password.' },
    { ok: report.SessionHours <= 24, label: 'Durata della sessione', detail: report.SessionHours + (report.SessionHours === 1 ? ' ora' : ' ore') + (report.SessionHours <= 24 ? '.' : ': troppo lunga, meglio al massimo 24 ore.') },
    { ok: true, label: 'Intestazioni di sicurezza', detail: 'Politica dei contenuti (CSP), blocco dell’incorniciamento, nosniff e riferimenti ridotti sono sempre attivi.' },
    { ok: true, label: 'Limite di richieste per indirizzo', detail: 'Accessi 10 al minuto, moduli pubblici 5 al minuto, resto con soglie ampie.' },
    { ok: true, label: 'Richieste da altri siti', detail: 'Le modifiche che arrivano da un sito esterno vengono respinte.' },
    { ok: true, label: 'Attesa dopo accessi errati', detail: 'Da 1 a 30 secondi, crescente a ogni errore di fila.' },
  ] : [])
  const okCount = $derived(checks.filter(item => item.ok).length)
</script>

<div class="head">
  <h1>Sicurezza</h1>
  <p class="stamp"><span class="led" data-tone={failed ? 'red' : report ? 'green' : 'neutral'}></span>{failed ? 'Dati non disponibili' : report ? 'Registro attivo dal ' + since.format(new Date(report.Started * 1000)) : 'Caricamento…'}</p>
</div>

{#if report}
  <section class="kpis">
    <div class="kpi glass-panel"><span>Controlli superati</span><strong>{okCount}/{checks.length}</strong></div>
    <div class="kpi glass-panel"><span>Accessi errati</span><strong>{report.LoginFails}</strong></div>
    <div class="kpi glass-panel"><span>Richieste limitate</span><strong>{report.RateLimited}</strong></div>
    <div class="kpi glass-panel"><span>Da altri siti</span><strong>{report.CrossSite}</strong></div>
  </section>

  <section class="card glass-panel">
    <h2>Controlli</h2>
    <ul class="checks">
      {#each checks as item (item.label)}
        <li><span class="led" data-tone={item.ok ? 'green' : 'amber'}></span><div><b>{item.label}</b><small>{item.detail}</small></div></li>
      {/each}
    </ul>
  </section>

  <section class="card glass-panel">
    <h2>Ultimi eventi respinti</h2>
    {#if report.Events?.length}
      <div class="tablewrap"><table>
        <thead><tr><th>Ora</th><th>Evento</th><th>Indirizzo</th><th>Percorso</th></tr></thead>
        <tbody>
          {#each report.Events as event (event.At + event.IP + event.Path + event.Kind)}
            <tr><td>{time.format(new Date(event.At * 1000))}</td><td>{event.Kind}</td><td><code>{event.IP}</code></td><td><code>{event.Path}</code></td></tr>
          {/each}
        </tbody>
      </table></div>
      <p class="note">Il registro è in memoria: si azzera al riavvio del servizio e non conserva indirizzi su disco.</p>
    {:else}
      <p class="note">Nessun evento da quando il servizio è partito.</p>
    {/if}
  </section>
{:else if failed}
  <p class="glass-panel notice" role="alert">Impossibile leggere i dati di sicurezza.</p>
{/if}

<style>
  .head { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px 16px; margin-bottom: 8px; }
  h1 { margin: 0; font-size: 1.65rem; }
  .stamp { display: inline-flex; align-items: center; gap: 8px; margin: 0; font-size: .85rem; color: #2f4b5f; font-weight: 600; }
  .led { width: 10px; height: 10px; flex: none; border-radius: 50%; background: #708391; display: inline-block; margin-top: 5px; }
  .stamp .led { margin: 0; }
  .led[data-tone=green] { background: #1fa86b; box-shadow: 0 0 0 3px #1fa86b33; }
  .led[data-tone=amber] { background: #e0a21a; box-shadow: 0 0 0 3px #e0a21a33; }
  .led[data-tone=red] { background: #d0343f; box-shadow: 0 0 0 3px #d0343f33; }
  .kpis { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 170px), 1fr)); gap: 12px; margin-bottom: 12px; }
  .kpi { display: grid; gap: 2px; padding: 12px 16px; }
  .kpi span { color: #2f4b5f; font-size: .78rem; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; }
  .kpi strong { font-size: 2rem; color: #143a60; font-variant-numeric: tabular-nums; }
  .card { padding: 14px 16px; margin-bottom: 12px; }
  h2 { margin: 0 0 12px; font-size: 1.05rem; color: #143a60; }
  .checks { margin: 0; padding: 0; list-style: none; display: grid; gap: 12px; }
  .checks li { display: grid; grid-template-columns: auto 1fr; gap: 12px; }
  .checks b { display: block; }
  .checks small { display: block; color: #4d6578; font-size: .88rem; }
  .tablewrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; font-size: .9rem; }
  th, td { text-align: left; padding: 6px 10px; border-bottom: 1px solid #dbe7f0; white-space: nowrap; }
  th { color: #2f4b5f; font-size: .78rem; text-transform: uppercase; letter-spacing: .05em; }
  .note { margin: 10px 0 0; color: #4d6578; font-size: .88rem; }
  .notice { padding: 16px; }
</style>
