<div class="logo" role="img" aria-label="DXRegia">
  <span class="mark" aria-hidden="true">
    <i class="wave"></i><i class="wave late"></i>
    <b class="bars"><i></i><i></i><i></i><i></i></b>
  </span>
  <span class="word">DX<span>Regia</span></span>
</div>

<style>
  .logo { display: grid; justify-items: center; gap: 10px; margin-bottom: 6px; }
  .mark { position: relative; display: grid; place-items: center; width: 86px; height: 86px; }
  .wave { position: absolute; inset: 14px; border: 2px solid #2f78b3; border-radius: 50%; opacity: 0; animation: ripple 3.2s ease-out infinite; }
  .wave.late { animation-delay: 1.6s; }
  .bars { position: relative; display: flex; align-items: end; gap: 5px; height: 46px; }
  .bars i { width: 9px; border-radius: 3px; background: linear-gradient(180deg, #2f78b3, #143a60); animation: level 2.4s ease-in-out infinite; transform-origin: bottom; }
  .bars i:nth-child(1) { height: 14px; animation-delay: 0s; }
  .bars i:nth-child(2) { height: 25px; animation-delay: .25s; }
  .bars i:nth-child(3) { height: 36px; animation-delay: .5s; }
  .bars i:nth-child(4) { height: 46px; animation-delay: .75s; background: linear-gradient(180deg, #f0a93a, #e08a1a); }
  .word { font-size: 1.7rem; font-weight: 800; letter-spacing: .3px; color: #143a60; }
  .word span { color: #1fa86b; }
  @keyframes ripple { 0% { transform: scale(.55); opacity: .55; } 100% { transform: scale(1.55); opacity: 0; } }
  @keyframes level { 0%, 100% { transform: scaleY(.55); } 50% { transform: scaleY(1); } }
  @media (prefers-reduced-motion: reduce) { .wave, .bars i { animation: none; } .wave { display: none; } }
</style>
