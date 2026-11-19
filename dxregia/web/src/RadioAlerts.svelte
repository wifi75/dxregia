<script lang="ts">
  import { onMount, untrack } from 'svelte'
  import { locale, t } from './i18n.svelte'
  import { alertStorageKey, bands, localDay, spotKey, type Spot } from './spotTools'
  import { collectAlerts, emptyAlertBook, emptyRule, inQuietHours,
    parseWatchCalls, readAlertBook, validRule, type AlertBook, type AlertRule, type RadioAlert } from './radioAlerts'

  let { spots, available }: { spots: Spot[]; available?: boolean } = $props()
  let book = $state(emptyAlertBook())
  let draft = $state<AlertRule>(emptyRule())
  let ready = $state(false)
  let storageError = $state(false)
  let invalid = $state(false)
  let notificationError = $state(false)
  let previous: Set<string> | undefined
  let summaryDay = $state(localDay(new Date()))
  let summary = $derived(book.days[summaryDay] ?? { groups: 0, spots: 0 })
  let now = $state(Date.now())
  let today = $derived(localDay(new Date(now)))
  const time = $derived(new Intl.DateTimeFormat(locale(), {dateStyle: 'short', timeStyle: 'short'}))
  const number = $derived(new Intl.NumberFormat(locale(), {maximumFractionDigits: 3}))
  const canNotify = typeof Notification !== 'undefined' && window.isSecureContext
  let permission = $state<NotificationPermission>(canNotify ? Notification.permission : 'denied')
  const modes = ['CW','SSB','DIGI','FT8','FT4','RTTY','PSK31']

  onMount(() => {
    book = readAlertBook()
    ready = true
    const sync = (event: StorageEvent) => {
      if (event.key === alertStorageKey || event.key === null) book = readAlertBook()
    }
    const clock = setInterval(() => { now = Date.now() }, 30_000)
    const syncPermission = () => { if (canNotify) permission = Notification.permission }
    window.addEventListener('storage', sync)
    window.addEventListener('focus', syncPermission)
    return () => {
      clearInterval(clock)
      window.removeEventListener('storage', sync)
      window.removeEventListener('focus', syncPermission)
    }
  })

  // Read the latest revision under the browser’s same-origin lock when available.
  async function editBook<T>(edit: (current: AlertBook) => { changed: boolean; value: T }): Promise<T | undefined> {
    const apply = () => {
      const current = storageError ? $state.snapshot(book) : readAlertBook()
      const result = edit(current)
      if (!result.changed) return result.value
      try {
        localStorage.setItem(alertStorageKey, JSON.stringify(current))
        storageError = false
      } catch { storageError = true }
      book = current
      return result.value
    }
    if (navigator.locks) return navigator.locks.request(alertStorageKey, apply)
    return apply()
  }

  async function deliver(fresh: RadioAlert[]) {
    if (!fresh.length || !book.notifications || !canNotify || Notification.permission !== 'granted') return
    const current = new Date()
    if (inQuietHours(current, book.quietStart, book.quietEnd)) return
    const [latest] = fresh
    const body = fresh.length === 1
      ? [latest.call, number.format(latest.frequency) + ' kHz', latest.mode].filter(Boolean).join(' · ')
      : t('radioBatch', { n: fresh.length })
    try {
      const worker = await navigator.serviceWorker?.getRegistration()
      if (worker?.active) await worker.showNotification('DXRegia · ' + t('radioAlerts'), { body, tag: 'dxregia-radio-alerts' })
      else {
        const notification = new Notification('DXRegia · ' + t('radioAlerts'), { body, tag: 'dxregia-radio-alerts' })
        notification.onclick = () => { window.focus(); window.location.assign('/spots?view=alerts') }
      }
      notificationError = false
    } catch { notificationError = true }
  }

  $effect(() => {
    if (!ready || !available) return
    const old = previous
    previous = new Set(spots.map(spotKey))
    if (!old) return
    const incoming = spots.filter(spot => !old.has(spotKey(spot))).reverse()
    if (incoming.length) untrack(() => { void recordIncoming(incoming) })
  })

  async function recordIncoming(incoming: Spot[]) {
    const fresh = await editBook(current => {
      const date = new Date()
      const update = collectAlerts(current, incoming, date)
      const notify = update.fresh.length > 0 && current.notifications && canNotify && Notification.permission === 'granted' &&
        !inQuietHours(date, current.quietStart, current.quietEnd) && date.getTime() - (current.lastNotification ?? 0) >= 60_000
      if (notify) current.lastNotification = date.getTime()
      return { changed: update.updated, value: notify ? update.fresh : [] }
    })
    if (fresh?.length) await deliver(fresh)
  }

  function saveRule(event: SubmitEvent) {
    event.preventDefault()
    if (!validRule(draft)) { invalid = true; return }
    const next = $state.snapshot(draft)
    next.label = next.label.trim()
    next.frequency = next.frequency.trim().replace(',', '.')
    next.calls = parseWatchCalls(next.calls).join(', ')
    void editBook(current => {
      const existing = current.rules.find(rule => rule.id === next.id)
      if (existing) {
        current.rules = current.rules.map(rule => rule.id === next.id ? { ...next, known: rule.known } : rule)
      } else {
        if (current.rules.length >= 20) return { changed: false, value: false }
        current.rules.push({ ...next, id: crypto.randomUUID?.() ?? String(Date.now()) })
      }
      return { changed: true, value: true }
    }).then(saved => {
      invalid = !saved
      if (saved) draft = emptyRule()
    })
  }

  function removeRule(id: string) {
    void editBook(current => {
      current.rules = current.rules.filter(rule => rule.id !== id)
      return { changed: true, value: true }
    })
    if (draft.id === id) draft = emptyRule()
  }

  function toggleRule(id: string) {
    void editBook(current => {
      const rule = current.rules.find(candidate => candidate.id === id)
      if (!rule) return { changed: false, value: false }
      rule.enabled = !rule.enabled
      if (draft.id === id) draft.enabled = rule.enabled
      return { changed: true, value: true }
    })
  }

  function quiet(part: 'quietStart' | 'quietEnd', value: string) {
    void editBook(current => {
      current[part] = value
      return { changed: true, value: true }
    })
  }

  async function toggleNotifications() {
    if (!canNotify) return
    try {
      permission = book.notifications ? Notification.permission : await Notification.requestPermission()
      const enabled = !book.notifications && permission === 'granted'
      await editBook(current => {
        current.notifications = enabled
        return { changed: true, value: true }
      })
      notificationError = false
    } catch { notificationError = true }
  }

  function readAll() {
    void editBook(current => {
      for (const alert of current.alerts) alert.read = true
      return { changed: true, value: true }
    })
  }

  function criteria(rule: AlertRule) {
    return [rule.calls, rule.band, rule.mode, rule.frequency && rule.frequency + ' kHz',
      rule.portable && t('radioPortable'), rule.newDXCC && t('radioNewDXCC')].filter(Boolean).join(' · ')
  }
</script>

<div class="radio-grid">
  <section class="glass-panel radio-panel">
    <h2>{t('radioRules')}</h2>
    <form class="rule-form" onsubmit={saveRule}>
      <label>{t('radioName')}<input bind:value={draft.label} maxlength="60" required /></label>
      <label>{t('radioWatchCalls')}<input bind:value={draft.calls} maxlength="400" placeholder="IU3WJQ, JA1ABC" /></label>
      <label>{t('bands')}<select bind:value={draft.band}><option value="">{t('allShort')}</option>{#each bands as [band]}<option>{band}</option>{/each}</select></label>
      <label>{t('modes')}<select bind:value={draft.mode}><option value="">{t('allShort')}</option>{#each modes as mode}<option>{mode}</option>{/each}</select></label>
      <label>{t('radioExactFrequency')}<input inputmode="decimal" bind:value={draft.frequency} placeholder="14074" /></label>
      <label>{t('radioExpiry')}<input type="date" bind:value={draft.expires} /></label>
      <label class="check"><input type="checkbox" bind:checked={draft.newDXCC} /> {t('radioNewDXCC')}</label>
      <label class="check"><input type="checkbox" bind:checked={draft.portable} /> {t('radioPortable')}</label>
      <div class="actions">
        <button type="submit">{t('radioSave')}</button>
        {#if draft.id}<button type="button" onclick={() => { draft = emptyRule(); invalid = false }}>{t('clear')}</button>{/if}
      </div>
      {#if invalid}<p class="notice" role="status">{t('radioInvalid')}</p>{/if}
      {#if storageError}<p class="notice" role="status">{t('spotStorageUnavailable')}</p>{/if}
    </form>

    {#if book.rules.length}
      <ul class="rules">
        {#each book.rules as rule (rule.id)}
          <li>
            <div>
              <strong>{rule.label}</strong>
              <small>{criteria(rule)}{#if rule.expires} · {t('radioExpiry')} {rule.expires}{/if}</small>
              {#if rule.expires && rule.expires < today}<small>{t('radioExpired')}</small>{/if}
            </div>
            <div class="actions">
              <button type="button" aria-pressed={rule.enabled} onclick={() => toggleRule(rule.id)}>{rule.enabled ? t('radioActive') : t('radioOff')}</button>
              <button type="button" onclick={() => { draft = $state.snapshot(rule); invalid = false }}>{t('radioEdit')}</button>
              <button type="button" onclick={() => removeRule(rule.id)}>{t('spotRemoveFilter')}</button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="glass-panel radio-panel">
    <h2>{t('radioDelivery')}</h2>
    <p>{t('radioLocalNote')}</p>
    <button type="button" disabled={!canNotify || permission === 'denied' && !book.notifications}
      onclick={() => void toggleNotifications()}>{book.notifications ? t('radioDisableBrowser') : t('radioEnableBrowser')}</button>
    {#if !canNotify || permission === 'denied' || notificationError}
      <p role="status" class="notice">{t('radioBrowserUnavailable')}</p>
    {/if}
    <h3>{t('radioQuietHours')}</h3>
    <div class="actions">
      <label>{t('radioFrom')}<input type="time" value={book.quietStart} onchange={event => quiet('quietStart', event.currentTarget.value)} /></label>
      <label>{t('radioTo')}<input type="time" value={book.quietEnd} onchange={event => quiet('quietEnd', event.currentTarget.value)} /></label>
    </div>
    <h3>{t('radioDaily')}</h3>
    <label>{t('radioDay')}<input type="date" bind:value={summaryDay} max={today} /></label>
    <p aria-live="polite">{t('radioSummary', { groups: summary.groups, spots: summary.spots })}</p>
  </section>

  <section class="glass-panel radio-panel inbox">
    <header>
      <h2>{t('radioInbox')}</h2>
      {#if book.alerts.some(alert => !alert.read)}<button type="button" onclick={readAll}>{t('radioReadAll')}</button>{/if}
    </header>
    {#if book.alerts.length}
      <ul class="inbox-list">
        {#each book.alerts as alert (alert.id)}
          <li class:unread={!alert.read}>
            <div><strong>{alert.call}</strong> · {number.format(alert.frequency)} kHz · {alert.mode || '—'}{#if alert.entity} · {alert.entity}{/if}
              <small>{alert.rules.join(' · ')}{#if alert.count > 1} · {t('radioRepeated', { n: alert.count })}{/if}</small>
            </div>
            <time datetime={new Date(alert.last).toISOString()}>{time.format(alert.last)}</time>
          </li>
        {/each}
      </ul>
    {:else}<p>{t('radioEmpty')}</p>{/if}
  </section>
</div>

<style>
  .radio-grid { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 12px; align-items: start; }
  .radio-panel { padding: clamp(12px, 2vw, 20px); min-width: 0; }
  h2 { margin: 0 0 12px; font-size: 1.08rem; }
  h3 { margin: 14px 0 7px; font-size: 1rem; }
  p { margin: 8px 0; }
  .rule-form { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
  label { display: grid; gap: 3px; min-width: 0; font-weight: 650; font-size: .87rem; }
  label input:not([type=checkbox]), select { width: 100%; min-width: 0; min-height: 38px; }
  label.check { display: flex; align-items: center; gap: 6px; }
  .notice, .rule-form .actions { grid-column: 1 / -1; }
  .actions { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
  .rules, .inbox-list { list-style: none; padding: 0; margin: 12px 0 0; }
  li, header { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 8px; align-items: center; }
  li { padding: 8px 0; border-top: 1px solid #7798ae55; overflow-wrap: anywhere; }
  small { display: block; }
  .inbox { grid-column: 1 / -1; }
  .inbox .unread { border-inline-start: 3px solid #2d8dbc; padding-inline-start: 8px; }
  @media (max-width: 800px) { .radio-grid { grid-template-columns: minmax(0, 1fr); } }
  @media (max-width: 450px) { .rule-form { grid-template-columns: minmax(0, 1fr); } }
</style>
