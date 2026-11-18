<script lang="ts">
  import { onMount } from 'svelte'
  export type ContactRequest = {
    ID: string; Kind: string; Call: string; Name: string; Contact: string; Email: string;
    Host: string; Port: number; Message: string; MailStatus: string; ReceivedAt: string
  }

  let { onDraft }: { onDraft: (request: ContactRequest) => void } = $props()
  let requests = $state<ContactRequest[]>()
  let failed = $state(false)
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })

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
  onMount(() => { void refresh() })
</script>

<section aria-labelledby="requests-title">
  <div class="heading"><h2 id="requests-title">Richieste dal portale</h2><button type="button" onclick={refresh}>Aggiorna</button></div>
  {#if failed}<p role="status">Richieste non disponibili.</p>
  {:else if requests?.length}
    {#each requests as request (request.ID)}
      <details>
        <summary>{request.Call} · {request.Kind === 'partner' ? 'Peering' : 'Messaggio'} · {date.format(new Date(request.ReceivedAt))}</summary>
        <p>{request.Contact || 'Referente non indicato'} · <a href={'mailto:' + request.Email}>{request.Email}</a></p>
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
  section { margin-top: 20px; padding: 16px; background: white; border: 1px solid #d7dfe7; }
  .heading { display: flex; justify-content: space-between; align-items: center; gap: 8px; flex-wrap: wrap; }
  h2 { margin: 0; font-size: 1.15rem; }
  details { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  summary { cursor: pointer; font-weight: 650; }
  .message { white-space: pre-wrap; overflow-wrap: anywhere; }
  button { min-height: 44px; padding: 6px 12px; border: 1px solid #214e80; background: white; color: #214e80; border-radius: 4px; font: inherit; cursor: pointer; }
  :focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
</style>
