<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  index: number
  chapter: string
  label?: string
  kind?: 'hero' | 'chapter' | 'insight' | 'compare' | 'map' | 'workflow' | 'closing'
}>(), {
  label: '',
  kind: 'insight',
})

const progress = computed(() => `${Math.max(0, Math.min(100, (props.index / 21) * 100))}%`)
const formattedIndex = computed(() => String(props.index).padStart(2, '0'))
</script>

<template>
  <div class="live-stage" :data-kind="kind">
    <BoardAtmosphere />
    <header class="stage-header">
      <div class="stage-brand">
        <span class="stage-brand__glyph">W1</span>
        <span class="stage-brand__name">高顿 ComfyUI 实战课</span>
      </div>
      <div class="stage-chapter">{{ chapter }}</div>
    </header>

    <main class="stage-shell" :class="`stage-shell--${kind}`">
      <span class="stage-corner stage-corner--tl"></span>
      <span class="stage-corner stage-corner--tr"></span>
      <span class="stage-corner stage-corner--bl"></span>
      <span class="stage-corner stage-corner--br"></span>
      <div class="stage-content"><slot /></div>
    </main>

    <footer class="stage-footer">
      <span>{{ label || chapter }}</span>
      <div class="stage-progress" aria-hidden="true"><i :style="{ width: progress }"></i></div>
      <strong>{{ formattedIndex }} / 21</strong>
    </footer>
  </div>
</template>
