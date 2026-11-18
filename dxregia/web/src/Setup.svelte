<script lang="ts">
  import { onMount } from 'svelte'
  import { readSetupImport } from './setupImport'
  import type { SetupFields } from './setupImport'

  let available = $state(false)
  let alreadySaved = $state(false)
  let result = $state<{ started: boolean; port: number }>()
  let { onSaved }: { onSaved: () => void } = $props()
  let pending = $state(false)
  let error = $state('')
  let form = $state<SetupFields>({
    node: '', sysop: '', locator: '', port: 7300, name: '', qth: '', email: '', advertisedHost: '',
  })
  let profile = $state<SetupFields>()
  let profileDraft = $state<SetupFields>()
  let profileError = $state('')
  let profileMessage = $state('')
  let savingProfile = $state(false)
  let postponed = $state(false)
  let imported = $state('')
  let step = $state(1)

  async function importConfig(event: Event) {
    const input = event.currentTarget as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    error = ''
    imported = ''
    try {
      if (file.size > 256 * 1024) throw new Error('Il file supera 256 KB.')
      const fields = readSetupImport(await file.text())
      form = { ...form, ...fields }
      imported = 'Campi riconosciuti: ' + Object.keys(fields).length + '. Controlla i valori prima di applicare.'
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Impossibile leggere il file.'
    } finally {
      input.value = ''
    }
  }

  onMount(async () => {
    try {
      const [response, savedProfile] = await Promise.all([
        fetch('/api/setup', { cache: 'no-store' }), fetch('/api/node-settings', { cache: 'no-store' }),
      ])
      if (savedProfile.ok && savedProfile.status !== 204) profile = await savedProfile.json()
      if (response.ok && response.headers.get('content-type')?.includes('application/json')) {
        const setup = await response.json()
        available = setup.available === true
        if (setup.suggestedPort) form.port = setup.suggestedPort
        if (setup.saved) {
          form = { ...form, ...setup.saved }
          alreadySaved = true
        }
      }
    } catch {
      // The diagnostic page can run without the setup service.
    }
  })

  function exportSettings() {
    if (!profile) return
    const objectURL = URL.createObjectURL(new Blob([JSON.stringify(profile, null, 2)], { type: 'application/json' }))
    const link = document.createElement('a')
    link.href = objectURL
    link.download = 'dxregia-config.json'
    link.click()
    setTimeout(() => URL.revokeObjectURL(objectURL), 1000)
  }

  function editProfile() {
    if (!profile) return
    profileDraft = { ...profile }
    profileError = ''
    profileMessage = ''
  }

  async function updateProfile(event: SubmitEvent) {
    event.preventDefault()
    if (!profileDraft || savingProfile) return
    savingProfile = true
    profileError = ''
    try {
      const response = await fetch('/api/node-settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profileDraft),
      })
      if (!response.ok) {
        profileError = (await response.text()).trim() || 'Impossibile salvare il profilo.'
        return
      }
      profile = await response.json()
      profileDraft = undefined
      profileMessage = 'Profilo salvato. Mojo lo legge al prossimo aggiornamento.'
    } catch {
      profileError = 'Il servizio non risponde.'
    } finally {
      savingProfile = false
    }
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (pending) return
    if (step === 1 && !alreadySaved) {
      step = 2
      return
    }
    pending = true
    error = ''
    try {
      const response = await fetch('/api/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      if (!response.ok) {
        if (response.status === 503) alreadySaved = true
        error = (await response.text()).trim() || 'Salvataggio non riuscito.'
        return
      }
      result = await response.json()
      profile = { ...form }
      onSaved()
      available = false
    } catch {
      error = 'Il pannello non riesce a raggiungere il servizio di configurazione.'
    } finally {
      pending = false
    }
  }
</script>

{#if available || result || profile}
  <section aria-labelledby="setup-title">
    <h2 id="setup-title">{available || result ? "Prima configurazione" : "Configurazione sysop"}</h2>
    {#if result}
      <p role="status">{result.started ? 'Mojo è attivo su 127.0.0.1:' + result.port + '.' : 'Configurazione Mojo salvata.'}</p>
    {:else if profile && !available}
      <p>Nodo {profile.node} · Porta Telnet {profile.port}</p>
    {:else if postponed}
      <p>Configurazione rimandata. Puoi riprenderla da qui.</p>
      <button type="button" onclick={() => postponed = false}>Riprendi configurazione</button>
    {:else}
      {#if !alreadySaved}
        <details>
          <summary>Importa configurazione DXSpider</summary>
          <label>File di configurazione <input type="file" accept=".pm,.json,.conf,.txt" onchange={importConfig} /></label>
          <p>Puoi scegliere prima DXVars.pm e poi Listeners.pm. I dati importati compilano il modulo e puoi correggerli prima di salvare.</p>
        </details>
        {#if imported}<p role="status">{imported}</p>{/if}
      {/if}
      <form onsubmit={save}>
        {#if step === 1 && !alreadySaved}
          <h3>1. Identità del nodo e del sysop</h3>
          <div class="fields">
            <label>Nominativo del nodo <input bind:value={form.node} maxlength="12" required placeholder="IU3ABC-2" /></label>
            <label>Indicativo del sysop <input bind:value={form.sysop} maxlength="12" required placeholder="IU3ABC" /></label>
            <label>Nome del sysop <input bind:value={form.name} maxlength="120" /></label>
            <label>Email del sysop <input type="email" bind:value={form.email} maxlength="254" /></label>
            <label>Località (QTH) <input bind:value={form.qth} maxlength="120" /></label>
            <label>Locator <input bind:value={form.locator} maxlength="6" required placeholder="JN55AA" /></label>
            <label>Latitudine <input type="number" bind:value={form.latitude} min="-90" max="90" step="any" /></label>
            <label>Longitudine <input type="number" bind:value={form.longitude} min="-180" max="180" step="any" /></label>
          </div>
          <p>Se lasci vuote le coordinate, verrà usato il centro del locator.</p>
        {:else}
          <h3>2. Rete e avvio</h3>
          <p>Nodo {form.node} · Sysop {form.sysop} · Locator {form.locator}</p>
          <div class="fields">
            <label>Host pubblico <input bind:value={form.advertisedHost} maxlength="253" readonly={alreadySaved} placeholder="cluster.example.org" /></label>
            <label>Porta Telnet <input type="number" bind:value={form.port} min="1" max="65535" readonly={alreadySaved} required /></label>
          </div>
          <p>Il servizio web del pannello usa la porta {window.location.port || (window.location.protocol === 'https:' ? '443' : '80')} in questa sessione.</p>
        {/if}
        {#if error}<p class="error" role="alert">{error}</p>{/if}
        {#if step === 2 && !alreadySaved}<button class="secondary" type="button" onclick={() => step = 1}>Indietro</button>{/if}
        <button disabled={pending}>{pending ? 'Configurazione in corso…' : alreadySaved ? 'Riprova avvio' : step === 1 ? 'Continua' : 'Applica configurazione'}</button>
        {#if !alreadySaved}<button class="secondary" type="button" onclick={() => postponed = true}>Configura più tardi</button>{/if}
      </form>
    {/if}
    {#if profile && !available}
      <dl>
        <div><dt>Indicativo del sysop</dt><dd>{profile.sysop}</dd></div>
        <div><dt>Nome</dt><dd>{profile.name || '—'}</dd></div>
        <div><dt>Email</dt><dd>{profile.email || '—'}</dd></div>
        <div><dt>Località e locator</dt><dd>{[profile.qth, profile.locator].filter(Boolean).join(' · ')}</dd></div>
        <div><dt>Host pubblico</dt><dd>{profile.advertisedHost || '—'}</dd></div>
      </dl>
      {#if profileMessage}<p role="status">{profileMessage}</p>{/if}
      {#if profileDraft}
        <form onsubmit={updateProfile}>
          <div class="fields">
            <label>Nome del sysop <input bind:value={profileDraft.name} maxlength="120" /></label>
            <label>Email <input type="email" bind:value={profileDraft.email} maxlength="254" /></label>
            <label>Località (QTH) <input bind:value={profileDraft.qth} maxlength="120" /></label>
            <label>Locator <input bind:value={profileDraft.locator} maxlength="6" required /></label>
            <label>Latitudine <input type="number" bind:value={profileDraft.latitude} min="-90" max="90" step="any" /></label>
            <label>Longitudine <input type="number" bind:value={profileDraft.longitude} min="-180" max="180" step="any" /></label>
            <label>Host pubblico <input bind:value={profileDraft.advertisedHost} maxlength="253" /></label>
          </div>
          <p>Lascia vuote entrambe le coordinate per calcolarle dal locator.</p>
          {#if profileError}<p class="error" role="alert">{profileError}</p>{/if}
          <button disabled={savingProfile}>{savingProfile ? 'Salvataggio…' : 'Salva profilo'}</button>
          <button type="button" class="secondary" disabled={savingProfile} onclick={() => profileDraft = undefined}>Annulla</button>
        </form>
      {:else}
        <button type="button" onclick={editProfile}>Modifica profilo sysop</button>
        <button type="button" class="secondary" onclick={exportSettings}>Scarica configurazione JSON</button>
      {/if}
    {/if}
  </section>
{/if}

<style>
  section { margin-top: 20px; padding: 16px; background: #fff; border: 1px solid #d7dfe7; }
  h2 { margin: 0 0 12px; font-size: 1.15rem; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); gap: 12px 16px; }
  label { display: grid; gap: 5px; font-weight: 600; }
  input { color: #182636; border: 1px solid #aab8c5; border-radius: 4px; min-height: 44px; padding: 8px; font: inherit; }
  button { margin-top: 14px; min-height: 44px; padding: 7px 13px; border: 1px solid #214e80; border-radius: 4px; background: #214e80; color: #fff; font: inherit; font-weight: 650; cursor: pointer; }
  :is(button, input):focus-visible { outline: 2px solid #214e80; outline-offset: 3px; }
  button:disabled { opacity: .7; cursor: wait; }
  details { margin-bottom: 14px; padding: 10px; border: 1px solid #d7dfe7; }
  summary { cursor: pointer; font-weight: 650; }
  details label { margin-top: 8px; }
  .secondary { margin-inline: 4px; color: #214e80; background: #fff; }
  dl { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 170px),1fr)); gap: 12px; }
  dt { color: #485b6d; font-size: .85rem; }
  dd { margin: 0; overflow-wrap: anywhere; }
  .error { color: #9b2424; }
</style>
