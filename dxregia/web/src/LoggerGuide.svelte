<script lang="ts">
  import { t } from './i18n.svelte'

  let { connection, call = '' }: { connection?: { Host: string; Port: number }; call?: string } = $props()
  const loggers = [
    { name: 'Log4OM 2', platform: 'Windows', path: 'Cluster → Connection → New', hint: 'regLoggerLog4om', url: 'https://www.log4om.com/l4ong/usermanual/Quick_Start_Guide_EN.pdf' },
    { name: 'Ham Radio Deluxe Logbook', platform: 'Windows', path: 'DX Cluster → Options → Connections → Add', hint: 'regLoggerHrd', url: 'https://support.hamradiodeluxe.com/support/solutions/articles/51000052680-setup-configuration-dx-cluster' },
    { name: 'N1MM Logger+', platform: 'Windows', path: 'Telnet → Clusters → Edit List', hint: 'regLoggerN1mm', url: 'https://n1mmwp.hamdocs.com/manual-windows/telnet-window/' },
    { name: 'DXLab SpotCollector', platform: 'Windows', path: 'Config → Spot Sources', hint: 'regLoggerDxlab', url: 'https://www.dxlabsuite.com/dxlabwiki/ManagingSpotSources' },
    { name: 'Swisslog', platform: 'Windows', path: 'Telnet → Edit Server Configuration', hint: 'regLoggerSwisslog', url: 'https://swisslogforwindows.com/english/Help/Internet%20and%20Telnet/Telnet%20Support.htm' },
    { name: 'CQRLOG', platform: 'Linux', path: 'DX Cluster → Telnet', hint: 'regLoggerCqrlog', url: 'https://www.cqrlog.com/help/h21.html' },
    { name: 'MacLoggerDX', platform: 'macOS', path: 'DX Cluster → dxclusters.dat', hint: 'regLoggerMac', url: 'https://dogparksoftware.com/MacLoggerDX%20Help/mldxfc_dxclusters.html' },
  ] as const
</script>

<aside class="glass-panel logger-guide" aria-labelledby="logger-guide-title">
  <h2 id="logger-guide-title">{t('regLoggerTitle')}</h2>
  <ol class="steps">
    <li>{t('regLoggerStepConfirm')}</li>
    <li>{t('regLoggerStepConnect')}</li>
    <li>{t('regLoggerStepSpot')}</li>
  </ol>
  <dl class="settings">
    <div><dt>{t('regLoggerServer')}</dt><dd>{#if connection}<code>{connection.Host}</code>{:else}{t('regLoggerAskSysop')}{/if}</dd></div>
    {#if connection}<div><dt>{t('regLoggerPort')}</dt><dd><code>{connection.Port}</code></dd></div>{/if}
    <div><dt>{t('regLoggerLogin')}</dt><dd>{#if call}<code>{call}</code>{:else}{t('regLoggerYourCall')}{/if}</dd></div>
    <div><dt>{t('regPassword')}</dt><dd>{t('regLoggerYourPassword')}</dd></div>
  </dl>
  <h3>{t('regLoggerCompatible')}</h3>
  <ul>
    {#each loggers as logger (logger.name)}
      <li>
        <details>
          <summary><strong>{logger.name}</strong><span>{logger.platform}</span></summary>
          <p><b>{logger.path}</b></p>
          <p>{t(logger.hint)}</p>
          <a href={logger.url} target="_blank" rel="noopener noreferrer">{t('regLoggerManual')}</a>
        </details>
      </li>
    {/each}
  </ul>
  <p class="note">{t('regLoggerOther')}</p>
</aside>

<style>
  .logger-guide { padding: clamp(14px, 2vw, 20px); }
  h2 { color: #143a60; font-size: 1.16rem; margin: 0 0 8px; }
  h3 { font-size: .98rem; margin: 14px 0 6px; }
  .steps { margin: 0 0 12px; padding-inline-start: 20px; color: #325369; }
  .steps li { padding-bottom: 4px; }
  .settings { display: grid; gap: 4px; margin: 0; padding: 9px 10px; background: #ffffff8c; border: 1px solid #b9ccdb; border-radius: 8px; }
  .settings div { display: grid; grid-template-columns: minmax(70px, 1fr) minmax(0, 2.4fr); gap: 8px; }
  dt { color: #456074; } dd { margin: 0; font-weight: 650; overflow-wrap: anywhere; }
  ul { list-style: none; padding: 0; margin: 0; }
  li details { border-bottom: 1px solid #c4d8e6; padding: 7px 0; }
  summary { cursor: pointer; }
  summary span { float: inline-end; font-size: .78rem; color: #466477; margin-inline-start: 6px; }
  p { margin: 5px 0; }
  a { color: #175c82; font-weight: 650; }
  .note { margin: 10px 0 0; color: #456074; font-size: .84rem; }
</style>
