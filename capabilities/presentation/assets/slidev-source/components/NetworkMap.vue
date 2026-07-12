<template>
  <div class="network-map" aria-label="三度人脉关系图">
    <svg viewBox="0 0 980 390" role="img">
      <defs>
        <filter id="network-glow"><feGaussianBlur stdDeviation="4" result="blur" /><feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge></filter>
      </defs>
      <path class="network-route" d="M100 196 C210 70, 270 310, 400 196 S620 68, 760 196 S870 315, 930 196" />
      <path class="network-route network-route--soft" d="M100 196 C210 70, 270 310, 400 196 S620 68, 760 196 S870 315, 930 196" />
      <path class="network-route network-route--branch" d="M400 196 C460 108, 500 105, 560 145" />
      <path class="network-route network-route--branch" d="M760 196 C815 110, 860 110, 914 138" />
      <g class="network-node network-node--you" transform="translate(100 196)"><circle r="31" /><circle class="network-node__core" r="10" /><text y="62">你 / 0 度</text></g>
      <g class="network-node network-node--one" transform="translate(300 196)"><circle r="28" /><circle class="network-node__core" r="8" /><text y="58">直接熟人 / 1 度</text></g>
      <g class="network-node network-node--two" transform="translate(520 196)"><circle r="30" /><circle class="network-node__core" r="9" /><text y="60">朋友的朋友 / 2 度</text></g>
      <g class="network-node network-node--three" transform="translate(760 196)"><circle r="34" /><circle class="network-node__core" r="11" /><text y="66">三度人脉 / 3 度</text></g>
      <g class="network-node network-node--satellite" transform="translate(915 138)"><circle r="13" /><text y="34">客户</text></g>
      <g class="network-node network-node--satellite" transform="translate(560 145)"><circle r="13" /><text y="34">商家</text></g>
    </svg>
    <div class="network-status"><span class="status-dot"></span><b>BLUE OCEAN CHANNEL OPEN</b><small>隔着两层中间人，仍然可以被连接</small></div>
  </div>
</template>
