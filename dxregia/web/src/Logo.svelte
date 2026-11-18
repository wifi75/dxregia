<div class="logo" role="img" aria-label="DXRegia">
  <svg viewBox="0 0 316 96" aria-hidden="true" focusable="false">
    <circle class="ring" cx="83" cy="48" r="26" fill="none" stroke="#f0a23a" stroke-width="2"/>
    <circle class="ring late" cx="83" cy="48" r="26" fill="none" stroke="#f0a23a" stroke-width="2"/>
    <g fill="none" stroke-width="13" stroke-linecap="butt" stroke-linejoin="miter">
  <path transform="translate(22 26)" d="M4 4V40H14A18 18 0 0 0 14 4Z" stroke="#143a60"/>
  <path transform="translate(66 26)" d="M4 4L30 40M30 4L4 40" class="x" stroke="#f0a23a"/>
  <path transform="translate(112 26)" d="M4 40V4H15A10 10 0 0 1 15 24H4M16 24L28 40" stroke="#143a60"/>
  <path transform="translate(154 26)" d="M24 4H4V40H24M4 22H20" stroke="#143a60"/>
  <path transform="translate(192 26)" d="M30 12A18 18 0 1 0 30 32V22H17" stroke="#143a60"/>
  <path transform="translate(238 26)" d="M4 4V40" stroke="#143a60"/>
  <path transform="translate(258 26)" d="M4 40L18 4L32 40M9 28H27" stroke="#143a60"/>
  </g>
  </svg>
</div>

<style>
  .logo { display: grid; justify-items: center; margin: 4px 0 24px; }
  svg { width: min(280px, 80%); height: auto; overflow: visible; color: #143a60; }
  svg g { stroke: currentColor; }
  svg g .x { stroke: #f0a23a; }
  .ring { transform-box: fill-box; transform-origin: center; opacity: 0; animation: ring 3.4s ease-out infinite; }
  .ring.late { animation-delay: 1.7s; }
  @keyframes ring { 0% { transform: scale(.35); opacity: .65; } 100% { transform: scale(2.1); opacity: 0; } }
  @media (prefers-reduced-motion: reduce) { .ring { animation: none; display: none; } }
</style>
