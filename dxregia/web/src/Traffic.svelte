<script lang="ts">
  import { onMount } from 'svelte'

  type Day = { Views: number; Referrers: Record<string, number> }
  let data = $state<{ Views: number; Days: Record<string, Day> }>()
  let failed = $state(false)
  let days = $derived(Object.entries(data?.Days ?? {}).sort(([a], [b]) => b.localeCompare(a)))
  const count = new Intl.NumberFormat('it-IT')

  onMount(async () => {
    try {
      const response = await fetch('/api/traffic')
      if (!response.ok) throw new Error()
      data = await response.json()
    } catch {
      failed = true
    }
  })
</script>

<section aria-labelledby="traffic-title">
  <h2 id="traffic-title">Statistiche del portale</h2>
  {#if failed}
    <p role="status">Statistiche non disponibili.</p>
  {:else if data}
    <p><strong>{count.format(data.Views)}</strong> visualizzazioni totali della pagina degli spot.</p>
    {#if days.length}
      <p>Storico giornaliero (UTC) · ultimi 90 giorni</p>
      {#each days as [date, day] (date)}
        <details>
          <summary>{date} · {count.format(day.Views)} visualizzazioni</summary>
          <ul>
            {#each Object.entries(day.Referrers).sort((a, b) => b[1] - a[1]) as [domain, views] (domain)}
              <li>{domain} <strong>{count.format(views)}</strong></li>
            {/each}
          </ul>
        </details>
      {/each}
    {:else}
      <p>Nessuna visualizzazione registrata.</p>
    {/if}
  {:else}
    <p>Caricamento delle statistiche…</p>
  {/if}
</section>

<style>
  section { margin-top: 20px; padding: 16px; background: white; border: 1px solid #d7dfe7; }
  h2 { margin: 0; font-size: 1.15rem; }
  details { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  summary { cursor: pointer; color: #214e80; font-weight: 650; }
  ul { margin: 5px 0 0; padding: 0; list-style: none; }
  li { display: flex; justify-content: space-between; gap: 12px; padding: 5px 4px; overflow-wrap: anywhere; }
</style>
