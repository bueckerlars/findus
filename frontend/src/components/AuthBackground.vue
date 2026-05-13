<script setup lang="ts">
/** Decorative full-viewport background for guest auth screens (uses `html` theme CSS variables). */
</script>

<template>
  <div class="fx-auth-bg" aria-hidden="true">
    <div class="fx-auth-bg__mesh" />
    <div class="fx-auth-bg__sweep" />
    <div class="fx-auth-bg__orb fx-auth-bg__orb--a" />
    <div class="fx-auth-bg__orb fx-auth-bg__orb--b" />
    <div class="fx-auth-bg__orb fx-auth-bg__orb--c" />
    <svg class="fx-auth-bg__svg" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1600 900" preserveAspectRatio="xMidYMid slice">
      <defs>
        <linearGradient id="fx-auth-bg-stroke" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="rgb(var(--fx-item-accent-from-rgb))" stop-opacity="0.95" />
          <stop offset="40%" stop-color="rgb(var(--fx-item-accent-via-rgb))" stop-opacity="0.75" />
          <stop offset="100%" stop-color="rgb(var(--fx-item-accent-to-rgb))" stop-opacity="0.55" />
        </linearGradient>
        <linearGradient id="fx-auth-bg-fill" x1="100%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stop-color="rgb(var(--fx-accent-muted-rgb))" stop-opacity="0.38" />
          <stop offset="45%" stop-color="rgb(var(--fx-accent-soft-rgb))" stop-opacity="0.48" />
          <stop offset="100%" stop-color="rgb(var(--fx-item-accent-to-rgb))" stop-opacity="0.28" />
        </linearGradient>
        <filter id="fx-auth-bg-blur" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur in="SourceGraphic" stdDeviation="18" />
        </filter>
      </defs>
      <g filter="url(#fx-auth-bg-blur)">
        <path
          fill="url(#fx-auth-bg-fill)"
          d="M-120 620 C280 420 520 780 920 520s520-200 880 40v420H-120V620Z"
          opacity="1"
        />
      </g>
      <g class="fx-auth-bg__strokes" fill="none" stroke-linecap="round" vector-effect="non-scaling-stroke">
        <path stroke="url(#fx-auth-bg-stroke)" stroke-width="3.5" d="M-80 380C360 180 640 520 1040 300s520-240 720-120" opacity="0.72" />
        <path stroke="url(#fx-auth-bg-stroke)" stroke-width="2.4" d="M200 760c280-200 480-120 720-280s400-320 760-240" opacity="0.58" />
        <path stroke="url(#fx-auth-bg-stroke)" stroke-width="1.85" d="M-40 120c320 120 560-80 880 60s520 200 840 80" opacity="0.48" />
        <path stroke="url(#fx-auth-bg-stroke)" stroke-width="1.4" d="M100 520c240-180 520-40 800-160s520-280 780-140" opacity="0.4" />
      </g>
    </svg>
    <div class="fx-auth-bg__grain" />
    <div class="fx-auth-bg__vignette" />
  </div>
</template>
