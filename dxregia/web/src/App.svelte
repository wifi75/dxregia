<script lang="ts">
  import { onMount } from 'svelte'
  import Auth from './Auth.svelte'
  import Dashboard from './Dashboard.svelte'
  import { languages, lang, setLang, t } from './i18n.svelte'
  let langOpen = $state(false)
  import Spots from './Spots.svelte'
  import Contact from './Contact.svelte'
  import { trackPortalPresence } from './portalPresence'
  const contactPage = window.location.pathname === '/contatti'
  // Il pannello sysop è solo su /admin; ogni altra pagina è il portale pubblico.
  const publicPortal = !(window.location.pathname === '/admin' || window.location.pathname.startsWith('/admin/'))
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

  $effect(() => {
    document.title = 'DXRegia – ' + (!publicPortal ? 'Amministrazione' : contactPage ? t('cTitle') : t('spotsTitle'))
  })
</script>

<svelte:window onclick={() => { langOpen = false }} onkeydown={event => { if (event.key === 'Escape') langOpen = false }} />

<header>
  <div class="brandline">
    <a class="brand" href="/spots"><span class="signal-mark" aria-hidden="true"><i></i><i></i><i></i><i></i></span>DXRegia</a>
    {#if publicPortal && !contactPage}
      <span class="sep" aria-hidden="true">–</span>
      <h1 id="spots-title" class="page">{t('spotsTitle')}</h1>
    {/if}
  </div>
  <div class="tools">
  <nav aria-label="Portale">
    <a href="/spots" aria-current={publicPortal && !contactPage ? 'page' : undefined}>{publicPortal ? t('navSpots') : 'Spot in diretta'}</a>
    <a href="/contatti" aria-current={contactPage ? 'page' : undefined}>{publicPortal ? t('navContact') : 'Contatti'}</a>
  </nav>
  {#if publicPortal}
    <div class="lang">
      <button class="lang-btn" type="button" aria-haspopup="menu" aria-expanded={langOpen} title={t('langLabel')} aria-label={t('langLabel')}
        onclick={event => { event.stopPropagation(); langOpen = !langOpen }}>
        <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3c2.6 2.6 3.9 5.6 3.9 9s-1.3 6.4-3.9 9c-2.6-2.6-3.9-5.6-3.9-9S9.4 5.6 12 3z"/>
        </svg>
      </button>
      {#if langOpen}
        <ul class="lang-menu" role="menu" aria-label={t('langLabel')}>
          {#each languages as language (language.code)}
            <li role="none">
              <button type="button" role="menuitemradio" aria-checked={lang() === language.code}
                onclick={() => { setLang(language.code); langOpen = false }}>
                <img src={'/flags/' + (language.code === 'en' ? 'gb' : language.code) + '.svg'} alt="" width="22" height="16" />
                {language.name}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {/if}
  </div>
</header>

{#if contactPage}
  <Contact />
{:else if publicPortal}
  <Spots />
{:else}
<main>
  {#snippet nodeStatus()}
    <Dashboard {status} {failed} {refresh} />
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
  .lang { position: relative; align-self: center; }
  .lang-btn { display: inline-grid; place-items: center; width: 40px; min-height: 40px; padding: 0; border-radius: 50%; --button-face: #ffffffd9; --button-edge: #a4bccb; --button-ink: #214e80; box-shadow: none; }
  .lang-menu { position: absolute; right: 0; top: calc(100% + 8px); z-index: 20; min-width: 170px; margin: 0; padding: 6px; list-style: none; border: 1px solid #b9c9d7; border-radius: 12px; background: #fffffff2; backdrop-filter: blur(10px); box-shadow: 0 12px 28px #143a5933; }
  .lang-menu button { display: flex; align-items: center; gap: 10px; width: 100%; min-height: 40px; padding: 4px 10px; border: 0; border-radius: 8px; box-shadow: none; background: none; color: #234968; font-weight: 600; text-align: left; }
  .lang-menu button:hover:not(:disabled) { background: #e7f0f7; transform: none; filter: none; }
  .lang-menu button[aria-checked=true] { background: #dcebf6; color: #143a60; }
  .lang-menu img { border: 1px solid #15263826; border-radius: 2px; object-fit: cover; }
  .visually-hidden { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
  header .brand { display: inline-flex; gap: 9px; align-items: center; padding-inline: 0; font-size: 1.25rem; font-weight: 800; color: #143a60; }
  .signal-mark { display: inline-flex; align-items: end; gap: 2px; height: 18px; }
  .signal-mark i { display: block; width: 3px; height: 5px; background: #146690; border-radius: 1px; }
  .signal-mark i:nth-child(2) { height: 9px; }
  .signal-mark i:nth-child(3) { height: 13px; }
  .signal-mark i:nth-child(4) { height: 18px; background: #e39221; }
  p, dt { color: #485b6d; }
  header { justify-content: space-between; align-items: center; flex-wrap: wrap; }
  header nav { display: flex; flex-wrap: wrap; gap: .35rem; }
  .brandline { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
  .sep { color: #8aa3b5; font-size: 1.25rem; }
  .page { margin: 0; font-size: 1.25rem; font-weight: 650; color: #2a516f; white-space: nowrap; }
  .tools { display: flex; align-items: center; gap: .5rem; margin-left: auto; }
  @media (max-width: 640px) {
    header { gap: 4px; padding: 6px 12px; }
    header .brand { font-size: 1.1rem; }
    .sep, .page { font-size: 1.05rem; }
    .tools { width: 100%; justify-content: space-between; margin-left: 0; }
    header a { padding: .4rem .6rem; font-size: .92rem; }
    .lang-btn { width: 36px; min-height: 36px; }
  }
  header a { color: #234968; font-weight: 650; padding: .55rem .8rem; border-radius: .5rem; text-decoration: none; }
  header a:hover, header a[aria-current=page] { background: #e7f0f7; color: #143a60; }
  main { width: 100%; padding: 24px clamp(14px, 2vw, 36px); }
</style>
