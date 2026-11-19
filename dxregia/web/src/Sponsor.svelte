<script lang="ts">
  import { t, lang } from './i18n.svelte'

  // Il banner si può chiudere: la scelta vale 7 giorni su questo browser.
  const key = 'dxregia-sponsor-closed'
  const week = 7 * 24 * 3600 * 1000
  const hiddenNow = () => {
    try { return Date.now() - Number(localStorage.getItem(key) ?? 0) < week } catch { return false }
  }
  let hidden = $state(hiddenNow())
  const close = () => {
    hidden = true
    try { localStorage.setItem(key, String(Date.now())) } catch { /* solo una comodità */ }
  }
  const modes = ['FT8', 'FT4', 'JT65', 'JT9', 'JT4', 'Q65', 'MSK144', 'FST4', 'FST4W', 'WSPR', 'PSK31', 'PSK63', 'RTTY', 'CW', 'SSTV']
  const site = $derived('https://aurora.iu3wjq.com/' + (['it', 'es', 'de', 'fr'].includes(lang()) ? lang() : 'en') + '/')
</script>

{#if !hidden}
  <aside class="sponsor" aria-label={t('sponsorLabel') + ': Aurora'}>
    <svg class="icon" viewBox="0 0 64 64" width="68" height="68" aria-hidden="true">
      <rect width="64" height="64" rx="12" fill="#0b2233" />
      <g>
        <rect x="9" y="10" width="7" height="34" rx="1.5" fill="#2f78d0" /><rect x="19" y="10" width="7" height="34" rx="1.5" fill="#26b5a0" />
        <rect x="29" y="10" width="7" height="34" rx="1.5" fill="#e6c13c" /><rect x="39" y="10" width="7" height="34" rx="1.5" fill="#26b5a0" />
        <rect x="49" y="10" width="7" height="34" rx="1.5" fill="#2f78d0" />
      </g>
      <text x="32" y="56" text-anchor="middle" font-family="system-ui, sans-serif" font-weight="800" font-size="7.2" fill="#fff" letter-spacing=".5">MULTIMODE</text>
    </svg>
    <div class="body">
      <p class="head"><strong class="name">Aurora</strong><span class="free">{t('sponsorFree')}</span><span class="tagline">{t('sponsorTitle')}</span></p>
      <p class="text">{t('sponsorText')}</p>
      <ul class="modes" aria-label={t('modes')}>
        {#each modes as mode}<li>{mode}</li>{/each}
        <li>{t('sponsorSatellites')}</li>
      </ul>
    </div>
    <span class="burst" aria-hidden="true"><span>{t('sponsorNew')}</span></span>
    <a class="get" href={site} target="_blank" rel="noopener noreferrer">
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12m0 0l-4.5-4.5M12 15l4.5-4.5M4 20h16" /></svg>
      {t('sponsorGet')}
    </a>
    <button class="close" type="button" aria-label={t('sponsorClose')} title={t('sponsorClose')} onclick={close}>×</button>
  </aside>
{/if}

<style>
  .sponsor { position: relative; display: flex; align-items: center; gap: 18px; margin: 0 0 14px; padding: 14px 54px 14px 18px; border-radius: 12px; color: #fff; background: linear-gradient(100deg, #0b4a52 0%, #1f7a6d 55%, #38a878 100%); border: 1px solid #ffffff33; box-shadow: 0 6px 18px #0b2a3a40; }
  .icon { flex: none; border-radius: 14px; box-shadow: 0 3px 10px #0006; }
  .body { flex: 1; min-width: 0; }
  .head { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 12px; margin: 0 0 4px; }
  .name { font-size: 1.5rem; line-height: 1; }
  .free { padding: 2px 10px; border-radius: 999px; background: #f6c33a; color: #3a2a00; font-size: .72rem; font-weight: 800; letter-spacing: .04em; text-transform: uppercase; }
  .tagline { font-size: 1.05rem; font-weight: 650; }
  .text { margin: 0 0 8px; font-size: .86rem; line-height: 1.4; opacity: .95; }
  .modes { display: flex; flex-wrap: wrap; gap: 5px; margin: 0; padding: 0; list-style: none; }
  .modes li { padding: 1px 9px; border-radius: 999px; border: 1px solid #ffffff55; background: #ffffff1f; font-size: .72rem; font-weight: 700; }
  .burst {
    --burst-shape: polygon(42% 0%, 52% 9%, 63% 2%, 70% 15%, 83% 13%, 85% 27%, 98% 32%, 93% 45%, 100% 55%, 89% 63%, 93% 76%, 80% 79%, 77% 92%, 64% 89%, 57% 100%, 46% 92%, 35% 99%, 29% 86%, 15% 88%, 14% 74%, 2% 69%, 8% 56%, 0% 47%, 11% 38%, 6% 25%, 19% 22%, 22% 8%, 35% 12%);
    flex: none;
    display: grid;
    place-items: center;
    width: 108px;
    aspect-ratio: 1;
    background: #f87810;
    color: #3a2a00;
    font-size: .82rem;
    font-weight: 900;
    line-height: 1;
    text-align: center;
    text-transform: uppercase;
    clip-path: var(--burst-shape);
  }
  .burst::before {
    content: '';
    grid-area: 1 / 1;
    place-self: stretch;
    margin: 5px;
    background: #ffd23e;
    clip-path: var(--burst-shape);
  }
  .burst span {
    grid-area: 1 / 1;
    position: relative;
    max-width: 65px;
    transform: rotate(-9deg);
  }
  @media (prefers-reduced-motion: no-preference) {
    .burst { animation: aurora-wave 5s ease-in-out 1s infinite; }
    @keyframes aurora-wave {
      0%, 30%, 100% { transform: rotate(0deg) scale(1); }
      9% { transform: rotate(7deg) scale(1.08); }
      18% { transform: rotate(-4deg) scale(1.03); }
    }
  }
  .get { flex: none; display: inline-flex; align-items: center; gap: 8px; padding: 12px 24px; border-radius: 999px; background: #fff; color: #0b3a4a; font-weight: 800; text-decoration: none; box-shadow: 0 3px 10px #0004; }
  .get:hover { background: #f1fbf7; }
  .close { position: absolute; top: 10px; right: 12px; display: grid; place-items: center; width: 30px; height: 30px; min-height: 0; padding: 0; border: 0; border-radius: 50%; background: transparent; box-shadow: none; color: #fff; font-size: 1.5rem; line-height: 1; cursor: pointer; }
  .close:hover:not(:disabled) { background: #ffffff2e; transform: none; filter: none; }
  .close:active:not(:disabled) { transform: none; box-shadow: none; }
  @media (max-width: 900px) {
    .burst, .text, .modes, .tagline { display: none; }
    .sponsor { gap: 12px; padding: 10px 44px 10px 12px; margin-bottom: 10px; }
    .icon { width: 44px; height: 44px; border-radius: 10px; }
    .name { font-size: 1.25rem; }
    .get { position: static; margin-left: auto; padding: 8px 14px; font-size: .9rem; }
    .get::after { content: ''; position: absolute; inset: 0; border-radius: 12px; }
    .close { z-index: 1; top: 50%; margin-top: -15px; right: 8px; }
  }
</style>
