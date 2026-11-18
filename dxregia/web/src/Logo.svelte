<div class="logo" role="img" aria-label="DXRegia">
  <span class="mark" aria-hidden="true">
    <i class="wave"></i><i class="wave late"></i>
    <img src="/brand/dxregia-icon.svg" alt="" width="72" height="72">
  </span>
  <span class="word">DX<span>Regia</span></span>
</div>

<style>
  .logo { display: grid; justify-items: center; gap: 10px; margin-bottom: 22px; }
  .mark { position: relative; display: grid; place-items: center; width: 96px; height: 96px; }
  .mark img { position: relative; border-radius: 18px; box-shadow: 0 6px 16px #143a5940; }
  .wave { position: absolute; inset: 8px; border-radius: 22px; border: 2px solid #2f78b3; opacity: 0; animation: ripple 3.2s ease-out infinite; }
  .wave.late { animation-delay: 1.6s; }
  .word { font-size: 1.7rem; font-weight: 800; letter-spacing: .3px; color: #143a60; }
  .word span { color: #1fa86b; }
  @keyframes ripple { 0% { transform: scale(.8); opacity: .55; } 100% { transform: scale(1.35); opacity: 0; } }
  @media (prefers-reduced-motion: reduce) { .wave { animation: none; display: none; } }
</style>
