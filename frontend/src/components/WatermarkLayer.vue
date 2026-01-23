<template>
  <div
    v-if="watermarkText"
    class="watermark-layer"
    :style="watermarkStyle"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

const watermarkText = computed(() => {
  const email = userStore.user?.email?.trim()
  if (email) {
    return email
  }
  const username = userStore.user?.username?.trim()
  return username || ''
})

function escapeSvgText(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const watermarkStyle = computed(() => {
  if (!watermarkText.value) {
    return {}
  }

  const safeText = escapeSvgText(watermarkText.value.replace(/[\r\n]+/g, ' '))
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="220" viewBox="0 0 320 220">
  <text x="0" y="120" fill="rgba(0,0,0,0.13)" font-size="16" font-family="Segoe UI, Arial, sans-serif" transform="rotate(-20 0 120)">${safeText}</text>
</svg>`
  const encodedSvg = encodeURIComponent(svg)

  return {
    '--watermark-image': `url("data:image/svg+xml,${encodedSvg}")`,
  } as Record<string, string>
})
</script>

<style scoped>
.watermark-layer {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 9999;
  background-image: var(--watermark-image);
  background-repeat: repeat;
  background-size: 320px 220px;
}
</style>
