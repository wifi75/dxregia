<script lang="ts">
  import { onMount } from 'svelte'
  import land from './propagationLand.json'
  import { locale, t } from './i18n.svelte'

  const regions = {
    NA: [-102, 44, 'propNorthAmerica'], SA: [-61, -20, 'propSouthAmerica'], EU: [16, 51, 'propEurope'],
    AF: [20, 7, 'propAfrica'], AS: [103, 42, 'propAsia'], OC: [142, -25, 'propOceania'], AN: [0, -74, 'propAntarctica'],
  } as const
  type Region = keyof typeof regions
  type Route = { From: Region; To: Region; Band: string; Count: number }
  type Report = { At: string; Minutes: number; Routes: Route[] }
  const regionCodes = Object.keys(regions) as Region[]
  const width = 900
  const height = 450
  const point = (lon: number, lat: number) => [(lon + 180) * width / 360, (90 - lat) * height / 180] as const

  let report = $state<Report>()
  let failed = $state(false)
  let from = $state<Region | ''>('')
  let band = $state('')
  let canvas = $state<HTMLCanvasElement>()
  let request: AbortController | undefined

  const clock = $derived(new Intl.DateTimeFormat(locale(), { timeZone: 'UTC', hour: '2-digit', minute: '2-digit' }))
  const number = $derived(new Intl.NumberFormat(locale()))
  const regionName = (region: Region) => t(regions[region][2])
  const scoped = $derived((report?.Routes ?? []).filter(route => !from || route.From === from))
  const bandTotals = $derived.by(() => {
    const count = new Map<string, number>()
    for (const route of scoped) count.set(route.Band, (count.get(route.Band) ?? 0) + route.Count)
    return [...count].sort((a, b) => b[1] - a[1])
  })
  const corridors = $derived.by(() => {
    const groups = new Map<string, { from: Region; to: Region; count: number; bands: Route[] }>()
    for (const route of scoped) {
      if (band && band !== route.Band) continue
      const key = route.From + route.To
      let group = groups.get(key)
      if (!group) {
        group = { from: route.From, to: route.To, count: 0, bands: [] }
        groups.set(key, group)
      }
      group.count += route.Count
      group.bands.push(route)
    }
    return [...groups.values()].sort((a, b) => b.count - a.count)
  })
  const plotted = $derived(corridors.slice(0, 8))
  const activeRegions = $derived(new Set(plotted.flatMap(route => [route.from, route.to])))

  function setOrigin(region: Region | '') {
    from = region
    if (!scoped.some(route => route.Band === band)) band = ''
  }
  function draw() {
    const ctx = canvas?.getContext('2d')
    if (!ctx) return
    ctx.fillStyle = '#0d2638'
    ctx.fillRect(0, 0, width, height)
    ctx.strokeStyle = '#305063'
    ctx.lineWidth = 1
    for (let lon = -150; lon < 180; lon += 30) {
      const [x] = point(lon, 0)
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, height); ctx.stroke()
    }
    for (let lat = -60; lat <= 60; lat += 30) {
      const [, y] = point(0, lat)
      ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(width, y); ctx.stroke()
    }
    ctx.fillStyle = '#284d5f'
    ctx.strokeStyle = '#4a7181'
    for (const ring of land) {
      ctx.beginPath()
      ring.forEach(([longitude, latitude], index) => {
        const [x, y] = point(longitude, latitude)
        if (index === 0) ctx.moveTo(x, y)
        else ctx.lineTo(x, y)
      })
      ctx.closePath(); ctx.fill(); ctx.stroke()
    }
    ctx.strokeStyle = '#8eafb9'
    ctx.setLineDash([5, 7])
    ctx.beginPath(); ctx.moveTo(0, height / 2); ctx.lineTo(width, height / 2); ctx.stroke()
    ctx.setLineDash([])
    for (const route of [...plotted].reverse()) {
      const [x, y] = point(regions[route.from][0], regions[route.from][1])
      const [targetX, endY] = point(regions[route.to][0], regions[route.to][1])
      let dx = targetX - x
      if (Math.abs(dx) > width / 2) dx -= Math.sign(dx) * width
      const dy = endY - y
      const distance = Math.hypot(dx, dy)
      ctx.strokeStyle = '#57ded7'
      ctx.fillStyle = '#57ded7'
      ctx.lineWidth = 1.5 + 2.5 * Math.sqrt(route.count / (plotted[0]?.count || 1))
      if (distance === 0) {
        ctx.beginPath(); ctx.arc(x, y, 23, 0, 2 * Math.PI); ctx.stroke()
        continue
      }
      const bend = Math.min(54, distance * .23)
      const controlX = x + dx / 2 + dy / distance * bend
      const controlY = y + dy / 2 - dx / distance * bend
      for (const offset of (x + dx < 0 || x + dx > width ? [-width, 0, width] : [0])) {
        const startX = x + offset
        const endX = startX + dx
        const middleX = controlX + offset
        ctx.beginPath(); ctx.moveTo(startX, y)
        ctx.quadraticCurveTo(middleX, controlY, endX, endY); ctx.stroke()
        const u = .68
        const v = 1 - u
        const px = v * v * startX + 2 * u * v * middleX + u * u * endX
        const py = v * v * y + 2 * u * v * controlY + u * u * endY
        const angle = Math.atan2(v * (controlY - y) + u * (endY - controlY), v * (middleX - startX) + u * (endX - middleX))
        ctx.beginPath()
        ctx.moveTo(px + 8 * Math.cos(angle), py + 8 * Math.sin(angle))
        for (const side of [-1, 1]) ctx.lineTo(px + 8 * Math.cos(angle + side * 2.45), py + 8 * Math.sin(angle + side * 2.45))
        ctx.closePath(); ctx.fill()
      }
    }
  }

  $effect(draw)
  async function refresh() {
    if (request) return
    const controller = new AbortController()
    request = controller
    try {
      const response = await fetch('/api/propagation', { signal: AbortSignal.any([controller.signal, AbortSignal.timeout(8000)]) })
      if (!response.ok) throw new Error()
      report = await response.json() as Report
      if (band && !scoped.some(route => route.Band === band)) band = ''
      failed = false
    } catch {
      if (!controller.signal.aborted) failed = true
    } finally { request = undefined }
  }

  onMount(() => {
    void refresh()
    const timer = setInterval(() => { if (!document.hidden) void refresh() }, 30000)
    return () => { clearInterval(timer); request?.abort() }
  })
</script>

<section class="glass-panel observed" aria-labelledby="observed-title">
  <header>
    <div>
      <h2 id="observed-title">{t('propTitle')}</h2>
      <p>{t('propWindow', { minutes: report?.Minutes ?? 30 })}</p>
    </div>
    {#if report}<time datetime={report.At}>{clock.format(new Date(report.At))} UTC</time>{/if}
  </header>
  <div class="controls">
    <label>{t('propOrigin')}
      <select class="button-select" value={from} onchange={event => setOrigin(event.currentTarget.value as Region | '')}>
        <option value="">{t('propAllRegions')}</option>
        {#each regionCodes as region}<option value={region}>{regionName(region)}</option>{/each}
      </select>
    </label>
    <label>{t('bands')}
      <select class="button-select" bind:value={band}>
        <option value="">{t('allShort')}</option>
        {#each bandTotals as [name]}<option value={name}>{name}</option>{/each}
      </select>
    </label>
  </div>
  {#if failed}<p role="status" class="message">{t('propError')}</p>{/if}
  <div class="overview">
    <div class="map-box">
      <div class="propagation-map">
        <canvas bind:this={canvas} {width} {height} aria-hidden="true"></canvas>
        <span class="equator" aria-hidden="true">0°</span>
        {#each regionCodes as region}
          {@const [x, y] = point(regions[region][0], regions[region][1])}
          <button type="button" class:active={activeRegions.has(region)} aria-pressed={from === region}
            title={regionName(region)} aria-label={regionName(region)}
            style:left={(x * 100 / width) + '%'} style:top={(y * 100 / height) + '%'}
            onclick={() => setOrigin(from === region ? '' : region)}>{region}</button>
        {/each}
      </div>
      <p class="map-note"><span aria-hidden="true">↗</span> {t('propLegend')}</p>
    </div>
    <div class="bands">
      <h3>{t('propActiveBands')}</h3>
      {#if bandTotals.length}
        <ul>
          {#each bandTotals as [name, count]}
            <li>
              <button type="button" aria-pressed={band === name} aria-label={name + ' · ' + number.format(count) + ' spot'}
                onclick={() => band = band === name ? '' : name}>
                <strong>{name}</strong><span class="bar" aria-hidden="true"><i style:width={(100 * count / bandTotals[0][1]) + '%'}></i></span>
                <span>{number.format(count)}</span>
              </button>
            </li>
          {/each}
        </ul>
      {:else if !failed}
        <p role="status">{report ? t('propEmpty') : t('loading')}</p>
      {/if}
    </div>
  </div>
  {#if plotted.length}
    <h3>{t('propRoutes')}</h3>
    <div class="routes">
      {#each plotted as route}
        <article>
          <div><strong>{regionName(route.from)}</strong> <span aria-hidden="true">→</span><span class="screen-reader"> {t('propTowards')} </span> <strong>{regionName(route.to)}</strong></div>
          <div class="route-bands">
            {#each [...route.bands].sort((a, b) => b.Count - a.Count) as part}
              <span>{part.Band} <b>{number.format(part.Count)}</b></span>
            {/each}
          </div>
        </article>
      {/each}
    </div>
  {/if}
  <p class="map-note">{t('propLimit')}</p>
</section>

<style>
  .observed { padding: clamp(12px, 2vw, 22px); margin-bottom: 14px; }
  header { display: flex; justify-content: space-between; align-items: start; flex-wrap: wrap; gap: 6px 16px; }
  h2 { font-size: clamp(1.2rem, 3vw, 1.55rem); margin: 0; color: #153b57; }
  h3 { margin: 0 0 8px; font-size: 1rem; }
  p { margin: 3px 0 0; color: #456074; }
  time { font-size: .8rem; font-variant-numeric: tabular-nums; color: #456074; }
  .controls { display: flex; flex-wrap: wrap; gap: 8px 18px; margin: 14px 0; }
  label { display: grid; gap: 4px; font-size: .83rem; font-weight: 650; }
  select { min-height: 38px; }
  .overview { display: grid; grid-template-columns: minmax(0, 2.8fr) minmax(185px, 1fr); align-items: start; gap: 16px; margin-bottom: 16px; }
  .map-box { min-width: 0; }
  .propagation-map { position: relative; background: #0d2638; border-radius: 12px; overflow: hidden; }
  :global(html[data-theme=dark]) .propagation-map { filter: invert(1) hue-rotate(180deg); }
  .equator { position: absolute; top: 50%; right: 5px; transform: translateY(-100%); color: #b3c9d0; font-size: .68rem; }
  canvas { display: block; width: 100%; height: auto; }
  .propagation-map button { position: absolute; transform: translate(-50%, -50%); min-height: 26px; min-width: 28px; padding: 0 4px; border-radius: 8px; font-size: .69rem; --button-face: #0d2638; --button-ink: #d8e6eb; --button-edge: #83a8b7; box-shadow: none; }
  .propagation-map button:hover { transform: translate(-50%, -50%); }
  .propagation-map button.active { --button-edge: #72ede7; --button-ink: #b9fff9; }
  .propagation-map button[aria-pressed=true] { --button-face: #7ef3e7; --button-ink: #0d2638; --button-edge: #bafbf7; }
  .map-note { font-size: .82rem; margin-top: 8px; }
  .map-note span { color: #06766e; font-weight: 800; margin-right: 4px; }
  .bands ul { list-style: none; margin: 0; padding: 0; }
  .bands button { width: 100%; min-height: 32px; display: grid; grid-template-columns: 3.4em minmax(12px, 1fr) 3.25em; align-items: center; text-align: left; gap: 5px; padding: 3px 6px; font-size: .8rem; --button-face: transparent; --button-ink: #193c54; --button-edge: transparent; box-shadow: none; }
  .bands button[aria-pressed=true] { --button-face: #d7efef; --button-edge: #4fa19d; }
  .bands button span:last-child { text-align: right; font-variant-numeric: tabular-nums; }
  .bar, .bar i { display: block; height: 7px; border-radius: 10px; }
  .bar { background: #d5e5ed; }
  .bar i { background: #2c8a83; }
  .routes { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr)); gap: 8px; }
  article { padding: 9px 10px; border: 1px solid #c3d9e4; border-radius: 8px; background: #ffffff63; }
  .route-bands { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 5px; }
  .route-bands span { border-radius: 4px; background: #dceceb; padding: 1px 5px; font-size: .74rem; }
  .route-bands b { margin-inline-start: 4px; font-variant-numeric: tabular-nums; color: #295f59; }
  .message { margin-bottom: 10px; }
  .screen-reader { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
  @media (max-width: 700px) {
    .overview { grid-template-columns: 1fr; gap: 12px; }
    .bands ul { display: grid; grid-template-columns: repeat(auto-fit, minmax(135px, 1fr)); }
    .propagation-map button { min-height: 24px; min-width: 25px; font-size: .61rem; }
  }
</style>
