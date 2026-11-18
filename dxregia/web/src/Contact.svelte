<script lang="ts">
  import { dxccFlags } from './dxccFlags'
  import { t, lang, locale } from './i18n.svelte'

  let kind = $state<'message' | 'partner'>('message')
  let call = $state('')
  let contact = $state('')
  let email = $state('')
  let country = $state('')
  let name = $state('')
  let host = $state('')
  let port = $state<number>()
  let message = $state('')
  let pending = $state(false)
  let sent = $state(false)
  let problem = $state('')

  // Elenco dei paesi con il nome nella lingua scelta. Serve solo al sysop per rispondere nella lingua giusta.
  const codes = [...new Set(Object.values(dxccFlags))].filter(code => /^[a-z]{2}$/.test(code) && code !== 'un')
  let countries = $derived.by(() => {
    let names: Intl.DisplayNames | undefined
    try { names = new Intl.DisplayNames([locale()], { type: 'region' }) } catch { /* ripiega sul codice */ }
    return codes
      .map(code => ({ code, label: names?.of(code.toUpperCase()) ?? code.toUpperCase() }))
      .sort((a, b) => a.label.localeCompare(b.label, locale()))
  })

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (pending) return
    pending = true
    problem = ''
    try {
      const result = await fetch('/api/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          Kind: kind, Call: call, Contact: contact, Email: email, Country: country.toUpperCase(), Lang: lang(),
          Name: kind === 'partner' ? name : '', Host: kind === 'partner' ? host : '', Port: kind === 'partner' ? port : 0, Message: message,
        }),
      })
      if (!result.ok) {
        problem = result.status === 404 ? t('errUnavailable') : (await result.text()).trim()
        return
      }
      sent = true
    } catch {
      problem = t('errServer')
    } finally {
      pending = false
    }
  }
</script>

<main>
  <h1>{t('cTitle')}</h1>
  {#if sent}
    <div class="glass-panel done" role="status">
      <h2>{t('sentTitle')}</h2>
      <p>{t('sentText')}</p>
      <a href="/spots">{t('back')}</a>
    </div>
  {:else}
    <form class="glass-panel glass-enter" onsubmit={submit}>
      <p class="lead">{t('cLead')}</p>

      <div class="kinds" role="radiogroup">
        <label class="kind" class:on={kind === 'message'}>
          <input type="radio" name="kind" value="message" bind:group={kind} />
          <strong>{t('cKindMsg')}</strong><span>{t('cKindMsgHint')}</span>
        </label>
        <label class="kind" class:on={kind === 'partner'}>
          <input type="radio" name="kind" value="partner" bind:group={kind} />
          <strong>{t('cKindPartner')}</strong><span>{t('cKindPartnerHint')}</span>
        </label>
      </div>

      <fieldset>
        <legend>{t('secYou')}</legend>
        <div class="grid four">
          <label>{t('fCall')} <input bind:value={call} required maxlength="12" autocapitalize="characters" spellcheck="false" /></label>
          <label>{t('fName')} <input bind:value={contact} maxlength="120" autocomplete="name" /></label>
          <label>{t('fEmail')} <input bind:value={email} type="email" required maxlength="254" autocomplete="email" /></label>
          <label>{t('fCountry')}
            <span class="country">
              {#if country && codes.includes(country)}<img src={'/flags/' + country + '.svg'} alt="" width="26" height="19" />{/if}
              <select bind:value={country} autocomplete="country">
                <option value="">{t('fCountryNone')}</option>
                {#each countries as item (item.code)}<option value={item.code}>{item.label}</option>{/each}
              </select>
            </span>
          </label>
        </div>
        <p class="hint">{t('fCountryHelp')}</p>
      </fieldset>

      {#if kind === 'partner'}
        <fieldset>
          <legend>{t('secNode')}</legend>
          <div class="grid three">
            <label>{t('fNodeName')} <input bind:value={name} maxlength="120" required /></label>
            <label>{t('fHost')} <input bind:value={host} maxlength="253" required spellcheck="false" /></label>
            <label>{t('fPort')} <input type="number" bind:value={port} min="1" max="65535" required /></label>
          </div>
          <p class="hint">{t('nodeHelp')}</p>
        </fieldset>
      {/if}

      <label>{kind === 'partner' ? t('fNotes') : t('fMessage')}
        <textarea bind:value={message} maxlength="3000" rows={kind === 'partner' ? 2 : 4} required={kind === 'message'}></textarea>
      </label>
      {#if problem}<p role="alert">{problem}</p>{/if}
      <button disabled={pending}>{pending ? t('sending') : t('send')}</button>
    </form>
  {/if}
</main>

<style>
  main { width: 100%; max-width: 900px; margin: 0 auto; padding: 22px clamp(14px, 2vw, 36px); }
  h1 { margin: 0 0 14px; font-size: 1.65rem; }
  h2 { margin: 0 0 6px; font-size: 1.2rem; }
  p { margin: 0; }
  form, .done { display: grid; gap: 14px; padding: 18px 20px; }
  .lead { color: #485b6d; }
  .kinds { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 260px), 1fr)); gap: 10px; }
  .kind { position: relative; display: grid; gap: 2px; padding: 12px 14px 12px 40px; border: 1.5px solid #b9c9d7; border-radius: 10px; background: #ffffffa6; cursor: pointer; font-weight: 450; }
  .kind input { position: absolute; left: 14px; top: 15px; min-height: auto; width: 16px; height: 16px; accent-color: #214e80; }
  .kind span { color: #485b6d; font-size: .9rem; }
  .kind.on { border-color: #214e80; background: #e8f1f9; box-shadow: 0 0 0 3px #214e8022; }
  fieldset { display: grid; gap: 8px; margin: 0; padding: 0; border: 0; min-width: 0; }
  legend { padding: 0; margin-bottom: 8px; font-weight: 700; color: #143a60; }
  .grid { display: grid; gap: 10px 12px; }
  .grid.four { grid-template-columns: repeat(auto-fit, minmax(min(100%, 190px), 1fr)); }
  .grid.three { grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); }
  .hint { color: #485b6d; font-size: .88rem; }
  label { display: grid; gap: 4px; font-weight: 600; min-width: 0; }
  .country { display: flex; align-items: center; gap: 8px; }
  .country img { border: 1px solid #15263826; border-radius: 2px; object-fit: cover; }
  .country select { flex: 1; }
  input:not([type=radio]), select, textarea { padding: 8px; min-height: 42px; border: 1px solid #aab8c5; border-radius: 6px; font: inherit; font-weight: 450; min-width: 0; }
  button { justify-self: start; }
  :focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
  [role=alert] { color: #9b2424; }
</style>
