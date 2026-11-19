<script lang="ts">
  import { t, lang } from './i18n.svelte'

  let call = $state('')
  let name = $state('')
  let qth = $state('')
  let locator = $state('')
  let email = $state('')
  let password = $state('')
  let repeat = $state('')
  let website = $state('') // campo trappola: le persone non lo vedono
  let busy = $state(false)
  let sent = $state(false)
  let error = $state('')

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (busy) return
    error = ''
    if (password !== repeat) {
      error = t('regMismatch')
      return
    }
    busy = true
    try {
      const response = await fetch('/api/register', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Call: call, Name: name, QTH: qth, Locator: locator, Email: email, Password: password, Lang: lang(), Website: website }),
      })
      if (response.ok) {
        sent = true
        password = repeat = ''
      } else {
        error = ({ 400: t('regInvalid'), 409: t('regTaken'), 429: t('regLimit'), 502: t('regMailFail'), 503: t('regDown') } as Record<number, string>)[response.status] ?? t('regError')
      }
    } catch {
      error = t('regError')
    } finally {
      busy = false
    }
  }
</script>

<main>
  <h1>{t('regTitle')}</h1>
  {#if sent}
    <p class="done glass-panel" role="status">{t('regSent')}</p>
    <p><a href="/spots">{t('confBack')}</a></p>
  {:else}
    <p class="intro">{t('regIntro')}</p>
    <form class="glass-panel glass-enter" onsubmit={submit}>
      <label>{t('regCall')}
        <input bind:value={call} required maxlength="10" autocapitalize="characters" autocomplete="username" spellcheck="false" style:text-transform="uppercase" disabled={busy} />
      </label>
      <label>{t('regName')}
        <input bind:value={name} required maxlength="100" autocomplete="given-name" disabled={busy} />
      </label>
      <div class="location">
        <label>{t('regQTH')}
          <input bind:value={qth} required maxlength="100" autocomplete="address-level2" disabled={busy} />
        </label>
        <label>{t('regLocator')}
          <input bind:value={locator} required maxlength="6" pattern="[A-Ra-r]{2}[0-9]{2}([A-Xa-x]{2})?" placeholder="JN55TE" autocapitalize="characters" spellcheck="false" style:text-transform="uppercase" disabled={busy} />
        </label>
      </div>
      <label>{t('regEmail')}
        <input type="email" bind:value={email} required maxlength="254" autocomplete="email" disabled={busy} />
        <small>{t('regNote')}</small>
      </label>
      <label>{t('regPassword')}
        <input type="password" bind:value={password} required minlength="8" maxlength="72" pattern={'[\\x21-\\x7e]{8,72}'} autocomplete="new-password" disabled={busy} />
        <small>{t('regPasswordHint')}</small>
      </label>
      <label>{t('regPassword2')}
        <input type="password" bind:value={repeat} required minlength="8" maxlength="72" autocomplete="new-password" disabled={busy} />
      </label>
      <label class="trap" aria-hidden="true">Website
        <input bind:value={website} tabindex="-1" autocomplete="off" />
      </label>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button disabled={busy}>{busy ? t('regSending') : t('regSubmit')}</button>
    </form>
  {/if}
</main>

<style>
  main { width: 100%; max-width: 560px; margin: 0 auto; padding: 22px clamp(14px, 2vw, 36px); }
  h1 { margin: 0 0 .6rem; font-size: 1.6rem; color: #143a60; }
  .intro { margin: 0 0 1rem; color: #3b566d; }
  form, .done { display: grid; gap: 14px; padding: 18px 20px; }
  label { display: grid; gap: 4px; font-weight: 600; min-width: 0; }
  input { padding: 8px 10px; min-height: 42px; border: 1px solid #aab8c5; border-radius: 6px; font: inherit; font-weight: 450; min-width: 0; }
  .location { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 12px; }
  @media (max-width: 440px) { .location { grid-template-columns: 1fr; } }
  small { color: #5b6e80; font-weight: 400; }
  .error { margin: 0; color: #b3261e; font-weight: 600; }
  .trap { position: absolute; left: -9999px; width: 1px; height: 1px; overflow: hidden; }
  form button { justify-self: start; }
  a { color: #175c82; font-weight: 600; }
</style>
