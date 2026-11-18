<script lang="ts">
  // Piccolo punto interrogativo con spiegazione: compare al passaggio del mouse e alla tastiera (focus).
  let { text }: { text: string } = $props()
</script>

<span class="hint" tabindex="0" role="button" aria-label={text}>?<span class="tip" role="tooltip">{text}</span></span>

<style>
  .hint { position: relative; display: inline-grid; place-items: center; width: 1.25rem; height: 1.25rem; margin-left: .35rem; border-radius: 50%; background: #2f78b3; color: #fff; font: 700 .75rem/1 system-ui, sans-serif; cursor: help; vertical-align: middle; }
  .tip { display: none; position: absolute; z-index: 30; left: 50%; bottom: calc(100% + 8px); transform: translateX(-50%); width: max-content; max-width: min(300px, 80vw); padding: .55rem .7rem; border-radius: 8px; background: #143a60; color: #fff; font-weight: 400; font-size: .82rem; line-height: 1.4; text-align: left; box-shadow: 0 6px 18px #0004; }
  .hint:hover .tip, .hint:focus .tip { display: block; }
</style>
