<script lang="ts">
  import { onMount } from 'svelte'
  import Auth from './Auth.svelte'
  import Dashboard from './Dashboard.svelte'
  import Spots from './Spots.svelte'
  import Contact from './Contact.svelte'
  import { trackPortalPresence } from './portalPresence'
  const contactPage = window.location.pathname === '/contatti'
  const publicPortal = contactPage || window.location.pathname === '/spots'
  if (publicPortal) trackPortalPresence()

  type Status = { address: string; state: 'ready' | 'unexpected' | 'unreachable'; checkedAt: string }
  let status = $state<Status>()
  let busy = $state(false)
  let failed = $state(false)
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

  onMount(() => { if (!publicPortal) void refresh() })
</script>

<header>
  <a class="brand" href="/spots"><span class="signal-mark" aria-hidden="true"><i></i><i></i><i></i><i></i></span>DXRegia</a>
  <nav aria-label="Portale">
    <a href="/spots" aria-current={publicPortal && !contactPage ? 'page' : undefined}>Spot in diretta</a>
    <a href="/contatti" aria-current={contactPage ? 'page' : undefined}>Contatti</a>
    <a href="/" aria-current={!publicPortal ? 'page' : undefined}>Pannello sysop</a>
  </nav>
</header>

{#if contactPage}
  <Contact />
{:else if publicPortal}
  <Spots />
{:else}
<main>
  {#snippet nodeStatus()}
    <Dashboard {status} {failed} {busy} {refresh} />
  {/snippet}
  <Auth onSaved={refresh} {nodeStatus} />
</main>
{/if}

<style>
  :global(*) { box-sizing: border-box; }
  :global(html) { color-scheme: light; }
  :global(body) {
    margin: 0;
    min-width: 280px;
    min-height: 100vh;
    font: 15px/1.5 system-ui, sans-serif;
  }
  header {
    display: flex;
    align-items: baseline;
    gap: 18px;
    background: #ffffffb9;
    backdrop-filter: blur(14px);
    border-bottom: 1px solid #a9c3d088;
    padding: 8px clamp(12px, 2vw, 32px);
  }
  header .brand { display: inline-flex; gap: 9px; align-items: center; padding-inline: 0; font-size: 1.25rem; font-weight: 800; color: #143a60; }
  .signal-mark { display: inline-flex; align-items: end; gap: 2px; height: 18px; }
  .signal-mark i { display: block; width: 3px; height: 5px; background: #146690; border-radius: 1px; }
  .signal-mark i:nth-child(2) { height: 9px; }
  .signal-mark i:nth-child(3) { height: 13px; }
  .signal-mark i:nth-child(4) { height: 18px; background: #e39221; }
  p, dt { color: #485b6d; }
  header { justify-content: space-between; align-items: center; flex-wrap: wrap; }
  header nav { display: flex; flex-wrap: wrap; gap: .35rem; }
  header a { color: #234968; font-weight: 650; padding: .55rem .8rem; border-radius: .5rem; text-decoration: none; }
  header a:hover, header a[aria-current=page] { background: #e7f0f7; color: #143a60; }
  main { width: 100%; padding: 24px clamp(14px, 2vw, 36px); }
</style>
