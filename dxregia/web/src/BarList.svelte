<script lang="ts">
  // Elenco con barre orizzontali: nome, barra proporzionale, numero e percentuale.
  type Item = { name: string; count: number; color?: string; flag?: string }
  let { items, total, empty = 'Ancora nessun dato.' }: { items: Item[]; total?: number; empty?: string } = $props()
  let sum = $derived(total ?? items.reduce((acc, item) => acc + item.count, 0))
  let max = $derived(Math.max(1, ...items.map(item => item.count)))
  const format = new Intl.NumberFormat('it-IT')
</script>

{#if items.length}
  <ul class="list">
    {#each items as item (item.name)}
      <li>
        <span class="name">{#if item.flag}<img src={'/flags/' + item.flag + '.svg'} alt="" width="22" height="16" loading="lazy" />{/if}{item.name}</span>
        <span class="track"><i style:width={(item.count / max) * 100 + '%'} style:background={item.color ?? '#2f78b3'}></i></span>
        <span class="num">{format.format(item.count)}</span>
        <span class="pct">{sum ? Math.round((item.count / sum) * 100) : 0}%</span>
      </li>
    {/each}
  </ul>
{:else}
  <p class="empty">{empty}</p>
{/if}

<style>
  .list { margin: 0; padding: 0; list-style: none; display: grid; gap: 7px; }
  li { display: grid; grid-template-columns: minmax(90px, 1.1fr) minmax(60px, 2fr) auto 38px; align-items: center; gap: 10px; font-size: .9rem; }
  .name { display: flex; align-items: center; gap: 8px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #1c3447; font-weight: 600; }
  .name img { border: 1px solid #15263826; border-radius: 2px; object-fit: cover; flex: 0 0 auto; }
  .track { height: 11px; border-radius: 6px; background: #dbe7f0; overflow: hidden; }
  .track i { display: block; height: 100%; border-radius: 6px; transition: width .5s ease; }
  .num { font-variant-numeric: tabular-nums; font-weight: 700; color: #14293a; text-align: right; }
  .pct { color: #4d6578; font-size: .8rem; text-align: right; font-variant-numeric: tabular-nums; }
  .empty { margin: 4px 0; color: #4d6578; }
  @media (prefers-reduced-motion: reduce) { .track i { transition: none; } }
</style>
