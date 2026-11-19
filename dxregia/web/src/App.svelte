<script lang="ts">
  import { onMount } from 'svelte'
  import Auth from './Auth.svelte'
  import Dashboard from './Dashboard.svelte'
  import { languages, lang, setLang, t } from './i18n.svelte'
  import Register from './Register.svelte'
  import RegisterConfirm from './RegisterConfirm.svelte'
  import { startAdminTranslation, translate } from './adminI18n'
  import { canInstall, needsManualInstall, installApp } from './install.svelte'
  import { soundOn, toggleSound } from './sound.svelte'
  import { currentTheme, toggleTheme } from './theme.svelte'
  let langOpen = $state(false)
  import Spots from './Spots.svelte'
  import ClusterStats from './ClusterStats.svelte'
  import Contact from './Contact.svelte'
  import { trackPortalPresence } from './portalPresence'
  import type { SpotView } from './spotTools'
  const contactPage = window.location.pathname === '/contatti'
  const registerPage = window.location.pathname === '/registrazione'
  const confirmPage = window.location.pathname === '/registrazione/conferma'
  const statsPage = window.location.pathname === '/statistiche'
  function viewFromURL(): SpotView {
    const view = new URLSearchParams(location.search).get('view')
    return view === 'research' || view === 'alerts' || view === 'ann' || view === 'propagation' ? view : 'feed'
  }
  let spotView = $state(viewFromURL())
  function openSpotView(event: MouseEvent, next: SpotView) {
    if (!publicPortal || toolPage || event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    if (spotView === next) return
    history.pushState(null, '', next === 'feed' ? '/spots' : '/spots?view=' + next)
    spotView = next
  }
  const toolPage = contactPage || registerPage || confirmPage || statsPage
  let registrationOpen = $state(false)
  if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/admin')) {
    fetch('/api/register/status').then(async response => { if (response.ok) registrationOpen = (await response.json()).Enabled === true }).catch(() => {})
  }
  // Il pannello sysop è solo su /admin; ogni altra pagina è il portale pubblico.
  const publicPortal = !(window.location.pathname === '/admin' || window.location.pathname.startsWith('/admin/'))
  if (publicPortal) trackPortalPresence()
  const titles = { feed: 'spotsTitle', research: 'spotResearch', alerts: 'radioAlerts', ann: 'navAnnouncements', propagation: 'navPropagation' } as const
  const portalTitle = $derived(t(titles[spotView]))

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

  let iosHint = $state(false)
  const onInstall = () => { if (needsManualInstall()) iosHint = !iosHint; else void installApp() }
  let adminTranslator: ReturnType<typeof startAdminTranslation> | undefined
  onMount(() => {
    if (publicPortal) return
    void refresh()
    adminTranslator = startAdminTranslation(document.getElementById('app')!)
    adminTranslator.setLanguage(lang())
    return () => adminTranslator?.stop()
  })
  $effect(() => { adminTranslator?.setLanguage(lang()) })

  $effect(() => {
    document.title = 'DXRegia – ' + (!publicPortal ? translate('Amministrazione', lang() as never) : contactPage ? t('cTitle') : registerPage ? t('regTitle') : confirmPage ? t('confTitle') : statsPage ? t('navStats') : portalTitle)
  })
</script>

<svelte:window onclick={() => { langOpen = false }} onpopstate={() => { spotView = viewFromURL() }} onkeydown={event => { if (event.key === 'Escape') langOpen = false }} />

<header>
  <div class="brandline">
    <a class="brand" href="/spots"><img class="brand-logo" src={currentTheme() === 'dark' ? '/brand/dxregia-logo-dark.svg' : '/brand/dxregia-logo.svg'} alt="DXRegia" width="112" height="34"></a>
    {#if publicPortal}
      <span class="sep" aria-hidden="true">–</span>
      {#if toolPage}<span class="page">{t('spotsTitle')}</span>
      {:else}<h1 id="spots-title" class="page">{t('spotsTitle')}</h1>{/if}
    {/if}
  </div>
  <div class="tools">
  <nav aria-label="Portale">
    <a href="/spots" aria-current={publicPortal && !toolPage && spotView === 'feed' ? 'page' : undefined} onclick={event => openSpotView(event, 'feed')}>{publicPortal ? t('navSpots') : 'Spot in diretta'}</a>
    {#if publicPortal}
      <a href="/spots?view=research" aria-current={!toolPage && spotView === 'research' ? 'page' : undefined} onclick={event => openSpotView(event, 'research')}>{t('spotResearch')}</a>
      <a href="/spots?view=alerts" aria-current={!toolPage && spotView === 'alerts' ? 'page' : undefined} onclick={event => openSpotView(event, 'alerts')}>{t('radioAlerts')}</a>
      <a href="/spots?view=ann" aria-current={!toolPage && spotView === 'ann' ? 'page' : undefined} onclick={event => openSpotView(event, 'ann')}>{t('navAnnouncements')}</a>
      <a href="/spots?view=propagation" aria-current={!toolPage && spotView === 'propagation' ? 'page' : undefined} onclick={event => openSpotView(event, 'propagation')}>{t('navPropagation')}</a>
    {/if}
    {#if publicPortal}<a class="stats-link" href="/statistiche" aria-current={statsPage ? 'page' : undefined}>{t('navStats')}</a>{/if}
    <a href="/contatti" aria-current={contactPage ? 'page' : undefined}>{publicPortal ? t('navContact') : 'Contatti'}</a>
    {#if publicPortal && registrationOpen}<a href="/registrazione" aria-current={registerPage || confirmPage ? 'page' : undefined}>{t('navRegister')}</a>{/if}
  </nav>
  {#if publicPortal && !toolPage}
    <button class="tool" type="button" aria-pressed={soundOn()} title={soundOn() ? t('soundOn') : t('soundOff')} aria-label={soundOn() ? t('soundOn') : t('soundOff')} onclick={toggleSound}>
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 9.5v5h3.5L12 18.5v-13L7.5 9.5z" fill="currentColor" stroke="none"/>
        {#if soundOn()}<path d="M15.5 9a4.2 4.2 0 0 1 0 6"/><path d="M18 6.5a8 8 0 0 1 0 11"/>{:else}<path d="M16 9.5l5 5M21 9.5l-5 5"/>{/if}
      </svg>
    </button>
  {/if}
  {#if publicPortal && canInstall()}
    <button class="tool" type="button" title={t('installApp')} aria-label={t('installApp')} onclick={onInstall}>
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="7" y="2.5" width="10" height="19" rx="2.5"/><path d="M12 8v6m0 0l-2.5-2.5M12 14l2.5-2.5"/></svg>
    </button>
  {/if}
  <button class="tool" type="button" title={currentTheme() === 'dark' ? t('themeLight') : t('themeDark')} aria-label={currentTheme() === 'dark' ? t('themeLight') : t('themeDark')} onclick={toggleTheme}>
    <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      {#if currentTheme() === 'dark'}<circle cx="12" cy="12" r="4"/><path d="M12 2.5v2.5M12 19v2.5M2.5 12H5M19 12h2.5M5.3 5.3l1.8 1.8M16.9 16.9l1.8 1.8M5.3 18.7l1.8-1.8M16.9 7.1l1.8-1.8"/>{:else}<path d="M20 14.5A8.5 8.5 0 0 1 9.5 4 8.5 8.5 0 1 0 20 14.5z"/>{/if}
    </svg>
  </button>
  {#if true}
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
{#if iosHint && publicPortal}
  <p class="install-hint" role="status">{t('installIos')}</p>
{/if}

{#if confirmPage}
  <RegisterConfirm />
{:else if registerPage}
  <Register />
{:else if contactPage}
  <Contact />
{:else if statsPage}
  <ClusterStats />
{:else if publicPortal}
  <Spots activePage={spotView} />
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
  :global(html) { color-scheme: light; scrollbar-gutter: stable; }
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
  .tool { display: inline-grid; place-items: center; width: 40px; min-height: 40px; padding: 0; border-radius: 50%; --button-face: #ffffffd9; --button-edge: #a4bccb; --button-ink: #214e80; box-shadow: none; }
  .tool[aria-pressed=false] { --button-ink: #7d8fa0; }
  .lang { position: relative; align-self: center; }
  .lang-btn { display: inline-grid; place-items: center; width: 40px; min-height: 40px; padding: 0; border-radius: 50%; --button-face: #ffffffd9; --button-edge: #a4bccb; --button-ink: #214e80; box-shadow: none; }
  .lang-menu { position: absolute; right: 0; top: calc(100% + 8px); z-index: 20; min-width: 170px; margin: 0; padding: 6px; list-style: none; border: 1px solid #b9c9d7; border-radius: 12px; background: #fffffff2; backdrop-filter: blur(10px); box-shadow: 0 12px 28px #143a5933; }
  .lang-menu button { display: flex; align-items: center; gap: 10px; width: 100%; min-height: 40px; padding: 4px 10px; border: 0; border-radius: 8px; box-shadow: none; background: none; color: #234968; font-weight: 600; text-align: left; }
  .lang-menu button:hover:not(:disabled) { background: #e7f0f7; transform: none; filter: none; }
  .lang-menu button[aria-checked=true] { background: #dcebf6; color: #143a60; }
  .lang-menu img { border: 1px solid #15263826; border-radius: 2px; object-fit: cover; }
  header .brand { display: inline-flex; gap: 9px; align-items: center; padding: 0; font-size: 1.25rem; font-weight: 800; line-height: 1; color: #143a60; }
  .brand-logo { display: block; height: 34px; width: auto; }
  header { justify-content: space-between; align-items: center; flex-wrap: wrap; }
  header nav { display: flex; flex-wrap: wrap; gap: .35rem; min-width: 0; }
  header nav a { white-space: nowrap; }
  header .tool, header .lang { flex: none; }
  header button:hover:not(:disabled), header button:active:not(:disabled) { transform: none; }
  :global(html.auth-card-view) .brandline { visibility: hidden; }
  .brandline { display: flex; align-items: center; gap: 10px; min-width: 0; }
  .sep { color: #8aa3b5; font-size: 1.25rem; line-height: 1; }
  .page { margin: 0; font-size: 1.25rem; font-weight: 650; line-height: 1; color: #2a516f; white-space: nowrap; }
  .tools { display: flex; align-items: center; gap: .5rem; margin-left: auto; min-width: 0; }
  @media (max-width: 640px) {
    header { gap: 4px; padding: 6px 12px; }
    header .brand { font-size: 1.1rem; }
    .sep, .page { font-size: 1.05rem; }
    .tools { width: 100%; justify-content: space-between; margin-left: 0; }
    header a { padding: .4rem .6rem; font-size: .92rem; }
    .lang-btn, .tool { width: 36px; min-height: 36px; }
  }
  @media (max-width: 420px) {
    header nav { flex-wrap: wrap; gap: .1rem; }
    header a { padding: .35rem .45rem; font-size: .85rem; white-space: nowrap; }
    .tools { gap: .3rem; }
    .lang-btn, .tool { width: 32px; min-height: 32px; }
  }
  header a { color: #234968; font-weight: 650; padding: .55rem .8rem; border-radius: .5rem; text-decoration: none; }
  header a:hover, header a[aria-current=page] { background: #e7f0f7; color: #143a60; }
  header .stats-link { background: #dcecf8; }
  main { width: 100%; padding: 24px clamp(14px, 2vw, 36px); }
  .install-hint { margin: 8px auto 0; max-width: 560px; padding: 10px 14px; border: 1px solid #9db8cc; border-radius: 12px; background: #ffffffee; color: #143a60; font-size: .92rem; text-align: center; }
</style>
