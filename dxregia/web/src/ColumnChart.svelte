<script lang="ts">
  // Grafico a colonne: ogni colonna ha un valore principale e, se serve, uno secondario più scuro (ad esempio i visitatori unici).
  type Item = { label: string; value: number; second?: number; hint?: string }
  let { items, height = 170, color = '#7fb1d9', secondColor = '#1f6fa8', every = 1 }: {
    items: Item[]; height?: number; color?: string; secondColor?: string; every?: number
  } = $props()
  let max = $derived(Math.max(1, ...items.map(item => item.value)))
  const format = new Intl.NumberFormat('it-IT')
</script>

<div class="chart" style:--h={height + 'px'} role="img" aria-label="Grafico a colonne">
  <div class="axis" aria-hidden="true"><span>{format.format(max)}</span><span>{format.format(Math.round(max / 2))}</span><span>0</span></div>
  <div class="cols">
    {#each items as item, i (i)}
      <div class="col" title={item.hint ?? item.label + ': ' + format.format(item.value)}>
        <div class="bar" style:height={(item.value / max) * 100 + '%'} style:background={color}>
          {#if item.second}<i style:height={Math.min(100, (item.second / Math.max(1, item.value)) * 100) + '%'} style:background={secondColor}></i>{/if}
        </div>
        <span class="lab">{i % every === 0 ? item.label : ''}</span>
      </div>
    {/each}
  </div>
</div>

<style>
  .chart { display: grid; grid-template-columns: 34px minmax(0, 1fr); gap: 6px; }
  .axis { display: flex; flex-direction: column; justify-content: space-between; height: var(--h); padding-bottom: 20px; text-align: right; font-size: .72rem; color: #4d6578; font-variant-numeric: tabular-nums; }
  .cols { display: flex; align-items: stretch; gap: 3px; height: calc(var(--h) + 20px); border-bottom: 1px solid #b6c9d7;
    background: linear-gradient(#c9d8e3 1px, transparent 1px) 0 0 / 100% calc(var(--h) / 2) repeat-y; background-origin: content-box; }
  .col { flex: 1 1 0; min-width: 0; display: flex; flex-direction: column; justify-content: flex-end; align-items: center; }
  .bar { position: relative; width: 100%; max-width: 38px; min-height: 2px; border-radius: 4px 4px 0 0; overflow: hidden; transition: height .5s ease; }
  .bar i { position: absolute; left: 0; right: 0; bottom: 0; transition: height .5s ease; }
  .col:hover .bar { filter: brightness(.9); }
  .lab { height: 20px; margin-top: 2px; font-size: .7rem; line-height: 20px; color: #3d566a; white-space: nowrap; font-variant-numeric: tabular-nums; }
  @media (prefers-reduced-motion: reduce) { .bar, .bar i { transition: none; } }
</style>
