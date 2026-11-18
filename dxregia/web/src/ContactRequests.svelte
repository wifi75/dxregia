<script lang="ts">
  import { onMount } from 'svelte'
  export type ContactRequest = {
    ID: string; Kind: string; Call: string; Name: string; Contact: string; Email: string;
    Host: string; Port: number; Message: string; MailStatus: string; ReceivedAt: string; Country?: string; Lang?: string; Auto?: boolean
  }

  let { onDraft }: { onDraft: (request: ContactRequest) => void } = $props()
  let requests = $state<ContactRequest[]>()
  let failed = $state(false)
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })
  const languageNames: Record<string, string> = { it: 'Italiano', en: 'English', es: 'Español', de: 'Deutsch', fr: 'Français', pt: 'Português' }
  function country(code?: string) {
    if (!code) return 'non indicato'
    try { return new Intl.DisplayNames(['it'], { type: 'region' }).of(code) ?? code } catch { return code }
  }

  let auto = $state<{ Enabled: boolean; DailyLimit: number }>()
  let autoError = $state('')

  async function loadAuto() {
    try {
      const response = await fetch('/api/auto-partner', { cache: 'no-store' })
      if (response.ok) auto = await response.json()
    } catch { /* l'interruttore resta nascosto */ }
  }

  async function saveAuto(enabled: boolean) {
    if (!auto) return
    autoError = ''
    try {
      const response = await fetch('/api/auto-partner', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Enabled: enabled, DailyLimit: auto.DailyLimit }),
      })
      if (!response.ok) { autoError = (await response.text()).trim(); return }
      auto = await response.json()
    } catch { autoError = 'Impossibile salvare l’impostazione.' }
  }

  async function refresh() {
    try {
      const response = await fetch('/api/contact-requests')
      if (!response.ok) throw new Error()
      const rows: ContactRequest[] | null = await response.json()
      requests = (rows ?? []).reverse()
      failed = false
    } catch {
      failed = true
    }
  }
  onMount(() => { void refresh(); void loadAuto() })
</script>

<section class="glass-panel" aria-labelledby="requests-title">
  <div class="heading"><h2 id="requests-title">Richieste dal portale</h2><button type="button" onclick={refresh}>Aggiorna</button></div>
  {#if auto}
    <div class="auto">
      <label class="switch"><input type="checkbox" checked={auto.Enabled} onchange={event => saveAuto(event.currentTarget.checked)} />
        <span>Configura da solo i nodi che chiedono il collegamento</span></label>
      <p class="hint">{auto.Enabled
        ? 'La richiesta crea la scheda del partner, abilita l’accesso in ingresso e gli manda le istruzioni via email, nella sua lingua. Non tocca mai un nodo già registrato né un account già presente, e si ferma a ' + auto.DailyLimit + ' nodi al giorno: le altre richieste restano da approvare qui.'
        : 'Le richieste restano da approvare: usa «Prepara scheda partner» e poi abilita l’accesso dalla scheda.'}</p>
      {#if autoError}<p role="alert">{autoError}</p>{/if}
    </div>
  {/if}
  {#if failed}<p role="status">Richieste non disponibili.</p>
  {:else if requests?.length}
    {#each requests as request (request.ID)}
      <details>
        <summary>{request.Call} · {request.Kind === 'partner' ? 'Peering' : 'Messaggio'} · {date.format(new Date(request.ReceivedAt))}</summary>
        <p>{request.Contact || 'Referente non indicato'} · <a href={'mailto:' + request.Email}>{request.Email}</a></p>
        {#if request.Auto}<p class="auto-tag">Configurato automaticamente: il partner ha ricevuto le istruzioni.</p>{/if}
        <p>Paese: {country(request.Country)} · Rispondi in: <strong>{languageNames[request.Lang ?? ''] ?? 'English'}</strong></p>
        {#if request.Kind === 'partner'}<p>{request.Name} · host {request.Host} · porta {request.Port}</p>{/if}
        {#if request.Message}<p class="message">{request.Message}</p>{/if}
        <p>Notifica email: {request.MailStatus}</p>
        {#if request.Kind === 'partner'}
          <button type="button" onclick={() => onDraft(request)}>Prepara scheda partner</button>
        {/if}
      </details>
    {/each}
  {:else if requests}<p>Nessuna richiesta ricevuta.</p>
  {:else}<p>Caricamento delle richieste…</p>{/if}
</section>

<style>
  .auto { margin: 10px 0 14px; padding: 10px 14px; border: 1px solid #cdd9e4; border-radius: 8px; background: #ffffffb8; }
  .switch { display: flex; align-items: center; gap: 10px; font-weight: 650; }
  .switch input { width: 18px; height: 18px; min-height: auto; accent-color: #167b50; }
  .hint { margin: 6px 0 0; color: #485b6d; font-size: .9rem; }
  .auto-tag { color: #167b50; font-weight: 650; }
  section { margin-top: 20px; padding: 16px; }
  .heading { display: flex; justify-content: space-between; align-items: center; gap: 8px; flex-wrap: wrap; }
  h2 { margin: 0; font-size: 1.15rem; }
  details { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  summary { cursor: pointer; font-weight: 650; }
  .message { white-space: pre-wrap; overflow-wrap: anywhere; }

  :focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
</style>
