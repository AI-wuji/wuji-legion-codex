<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const root = ref<HTMLDivElement | null>(null)
const canvas = ref<HTMLCanvasElement | null>(null)
let stop = () => {}

onMounted(() => {
  const host = root.value
  const element = canvas.value
  const context = element?.getContext('2d')
  if (!host || !element || !context) return

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const colors = [
    [132, 231, 255],
    [124, 141, 255],
    [86, 237, 196],
    [255, 217, 134],
  ]
  const plumes = Array.from({ length: 7 }, (_, index) => ({
    color: colors[index % colors.length],
    phase: index * 0.9,
    orbit: 0.14 + (index % 4) * 0.045,
    radius: 0.24 + (index % 3) * 0.045,
    speed: 0.00006 + (index % 4) * 0.000018,
  }))

  let width = 1
  let height = 1
  let frame = 0
  let visible = false

  const resize = () => {
    const rect = element.getBoundingClientRect()
    const dpr = Math.min(window.devicePixelRatio || 1, 1.5)
    width = Math.max(1, rect.width)
    height = Math.max(1, rect.height)
    element.width = Math.round(width * dpr)
    element.height = Math.round(height * dpr)
    context.setTransform(dpr, 0, 0, dpr, 0, 0)
  }

  const draw = (time: number) => {
    context.clearRect(0, 0, width, height)
    context.globalCompositeOperation = 'lighter'

    plumes.forEach((plume, index) => {
      const x = width * (0.5 + Math.sin(time * plume.speed + plume.phase) * plume.orbit)
      const y = height * (0.5 + Math.cos(time * plume.speed * 0.72 + plume.phase * 1.4) * plume.orbit)
      const radius = Math.max(width, height) * plume.radius
      const [r, g, b] = plume.color
      const gradient = context.createRadialGradient(x, y, 0, x, y, radius)
      gradient.addColorStop(0, `rgba(${r}, ${g}, ${b}, ${index === 0 ? 0.16 : 0.11})`)
      gradient.addColorStop(0.42, `rgba(${r}, ${g}, ${b}, 0.055)`)
      gradient.addColorStop(1, `rgba(${r}, ${g}, ${b}, 0)`)
      context.fillStyle = gradient
      context.fillRect(x - radius, y - radius, radius * 2, radius * 2)
    })

    context.globalCompositeOperation = 'source-over'
    for (let line = 0; line < 3; line += 1) {
      const offset = Math.sin(time * 0.00012 + line * 1.7) * height * 0.045
      context.strokeStyle = line === 1 ? 'rgba(86, 237, 196, 0.12)' : 'rgba(132, 231, 255, 0.1)'
      context.lineWidth = 1
      context.beginPath()
      context.moveTo(-width * 0.08, height * (0.31 + line * 0.17) + offset)
      context.bezierCurveTo(
        width * 0.24,
        height * (0.18 + line * 0.2) - offset,
        width * 0.66,
        height * (0.62 - line * 0.11) + offset,
        width * 1.08,
        height * (0.36 + line * 0.15) - offset,
      )
      context.stroke()
    }
  }

  const animate = (time: number) => {
    draw(time)
    if (visible && !reducedMotion) frame = window.requestAnimationFrame(animate)
  }

  const start = () => {
    if (visible || reducedMotion) return
    visible = true
    frame = window.requestAnimationFrame(animate)
  }

  const pause = () => {
    visible = false
    window.cancelAnimationFrame(frame)
  }

  resize()
  draw(0)
  const observer = new IntersectionObserver(([entry]) => {
    if (entry?.isIntersecting) start()
    else pause()
  }, { threshold: 0.05 })
  observer.observe(host)
  window.addEventListener('resize', resize)

  stop = () => {
    pause()
    observer.disconnect()
    window.removeEventListener('resize', resize)
  }
})

onBeforeUnmount(() => stop())
</script>

<template>
  <div ref="root" class="stage-fluid" aria-hidden="true">
    <canvas ref="canvas"></canvas>
    <div class="stage-fluid__grid"></div>
    <div class="stage-fluid__vignette"></div>
  </div>
</template>
