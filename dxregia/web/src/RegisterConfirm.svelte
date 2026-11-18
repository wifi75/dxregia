<script lang="ts">
  import { onMount } from 'svelte'
  import { t } from './i18n.svelte'

  let state = $state<'working' | 'ok' | 'bad' | 'expired' | 'refused' | 'pending'>('working')
  let call = $state('')

  // Il link si conferma con una richiesta POST dalla pagina: i programmi che aprono i link nelle email non registrano nulla.
  onMount(async () => {
    const token = new URLSearchParams(location.search).get('token') ?? ''
    try {
      const response = await fetch('/api/register/confirm', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ Token: token }),
      })
      if (response.status === 410) { state = 'expired'; return }
      if (!response.ok) { state = 'bad'; return }
      const data = await response.json() as { Call: string; Result: string }
      call = data.Call
      state = data.Result === 'Registrato' ? 'ok' : data.Result ? 'refused' : 'pending'
    } catch {
      state = 'bad'
    }
  })
</script>

<main>
  <h1>{t('confTitle')}</h1>
  <p class="glass-panel box" role="status" data-state={state}>
    {#if state === 'working'}{t('confWorking')}
    {:else if state === 'ok'}{t('confOk', { call })}
    {:else if state === 'expired'}{t('confExpired')}
    {:else if state === 'refused'}{t('confRefused')}
    {:else if state === 'pending'}{t('confPending')}
    {:else}{t('confBad')}{/if}
  </p>
  <p><a href="/spots">{t('confBack')}</a></p>
</main>

<style>
  main { width: 100%; max-width: 560px; margin: 0 auto; padding: 22px clamp(14px, 2vw, 36px); }
  h1 { margin: 0 0 .8rem; font-size: 1.6rem; color: #143a60; }
  .box { padding: 18px 20px; margin: 0 0 1rem; font-size: 1.05rem; line-height: 1.5; }
  .box[data-state=ok] { border-left: 5px solid #2fa36b; }
  .box[data-state=bad], .box[data-state=expired], .box[data-state=refused] { border-left: 5px solid #c5483d; }
  a { color: #175c82; font-weight: 600; }
</style>
