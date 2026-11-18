<script lang="ts">
  import { onMount } from 'svelte'

  type Channel = { Call: string; Kind: string; State: string; Since: number }
  type Snapshot = { At: number; Channels: Channel[] | null }
  let snapshot = $state<Snapshot>()
  let requested = false
  let partners = $state<Record<string, { Name: string }>>({})
  let channels = $derived(snapshot?.Channels ?? [])
  let nodes = $derived(channels.filter(channel => channel.Kind === 'node'))
  let sessions = $derived(channels.filter(channel => channel.Kind === 'telnet' || channel.Kind === 'web'))
  let missing = $derived(Object.keys(partners).filter(call => !nodes.some(node => node.Call === call)))
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })

  async function refresh() {
    if (requested) return
    requested = true
    try {
      const [presence, directory] = await Promise.all([fetch('/api/presence'), fetch('/api/partners')])
      snapshot = presence.ok ? await presence.json() : undefined
      if (directory.ok) partners = await directory.json()
    } catch {
      snapshot = undefined
    } finally {
      requested = false
    }
  }

  onMount(() => {
    void refresh()
    const timer = setInterval(refresh, 5000)
    return () => clearInterval(timer)
  })
</script>

<section class="glass-panel" aria-labelledby="presence-title">
  <h2 id="presence-title">Attività del nodo</h2>
  {#if snapshot?.At}
    <p>Ultimo aggiornamento {date.format(new Date(snapshot.At * 1000))}</p>
    <h3>Nodi partner</h3>
    {#if nodes.length || missing.length}
      <div class="overflow">
        <table>
          <thead><tr><th>Indicativo</th><th>Stato</th><th>Sessione iniziata</th></tr></thead>
          <tbody>
            {#each nodes as node (node.Call)}
              <tr><td>{node.Call}</td><td>{node.State === 'normal' ? 'Attivo' : 'Sessione aperta · ' + (node.State || 'stato sconosciuto')}</td><td>{date.format(new Date(node.Since * 1000))}</td></tr>
            {/each}
            {#each missing as call (call)}
              <tr><td>{call}</td><td>Disconnesso</td><td>—</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p>Nessun collegamento attivo o partner registrato.</p>
    {/if}
    <h3>Radioamatori collegati ({sessions.length})</h3>
    {#if sessions.length}
      <div class="overflow">
        <table>
          <thead><tr><th>Indicativo</th><th>Accesso</th><th>Collegato dal</th></tr></thead>
          <tbody>
            {#each sessions as session (session.Call)}
              <tr><td>{session.Call}</td><td>{session.Kind === 'telnet' ? 'Telnet' : 'Web Mojo'}</td><td>{date.format(new Date(session.Since * 1000))}</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p>Nessuna sessione Telnet o Web Mojo attiva.</p>
    {/if}
  {:else}
    <p>Dati di presenza non disponibili. La dashboard si aggiorna quando riceve dati recenti da Mojo.</p>
  {/if}
</section>

<style>
  section { margin-top: 20px; padding: 16px; }
  h2 { margin: 0; font-size: 1.15rem; }
  h3 { margin: 16px 0 6px; font-size: 1rem; }
  p { color: #485b6d; margin: 6px 0; }
  .overflow { overflow: auto; }
  table { width: 100%; text-align: left; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 6px 8px; border-bottom: 1px solid #d7dfe7; }
  th { background: #edf3f9; }
</style>
