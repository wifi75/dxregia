<div class="logo" role="img" aria-label="DXRegia">
  <svg viewBox="0 22 394 76" aria-hidden="true" focusable="false">
    <defs><clipPath id="logo-cap"><rect x="0" y="34" width="394" height="52"/></clipPath></defs>
    <circle class="ring" cx="92" cy="60" r="30" fill="none" stroke="#f0a23a" stroke-width="2"/>
    <circle class="ring late" cx="92" cy="60" r="30" fill="none" stroke="#f0a23a" stroke-width="2"/>
    <g clip-path="url(#logo-cap)" fill="none" stroke-width="12" stroke-linecap="butt" stroke-linejoin="miter" stroke-miterlimit="2.5">
    <path transform="translate(14.0 0)" d="M0 34V86M0 40H14A20 20 0 0 1 14 80H0" stroke="#143a60"/>
    <path transform="translate(74.0 0)" d="M-1.7 28L37.7 92M37.7 28L-1.7 92" class="x" stroke="#f0a23a"/>
    <path transform="translate(134.0 0)" d="M0 86V34M0 40H13A12 12 0 0 1 13 64H0M15 64L31.8 92" stroke="#143a60"/>
    <path transform="translate(188.0 0)" d="M26 40H0V80H26M0 60H20" stroke="#143a60"/>
    <path transform="translate(233.0 0)" d="M36.4 48.5A20 20 0 1 0 40 60H25" stroke="#143a60"/>
    <path transform="translate(298.0 0)" d="M0 34V86" stroke="#143a60"/>
    <path transform="translate(323.5 0)" d="M-2.7 92L21 40H35L58.7 92M7.3 70H48.7" stroke="#143a60"/>
  </g>
  </svg>
</div>

<style>
  .logo { display: grid; justify-items: center; margin: 4px 0 24px; }
  svg { width: min(300px, 80%); height: auto; overflow: visible; color: #143a60; }
  svg g { stroke: currentColor; }
  svg g .x { stroke: #f0a23a; }
  .ring { transform-box: fill-box; transform-origin: center; opacity: 0; animation: ring 3.4s ease-out infinite; }
  .ring.late { animation-delay: 1.7s; }
  @keyframes ring { 0% { transform: scale(.35); opacity: .65; } 100% { transform: scale(2.1); opacity: 0; } }
  @media (prefers-reduced-motion: reduce) { .ring { animation: none; display: none; } }
</style>
