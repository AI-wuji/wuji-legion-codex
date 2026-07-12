---
name: wuji-video-suite
description: Unified video Skill for generation, image-to-video, motion composition, timeline editing, rendering, and provider fallback. Use for any video task without exposing separate generation and composition systems.
---

# Wuji Video Suite

Choose one delivery path internally: direct generation, image-to-video, or composed motion/timeline output. Agnes is the default generator; an explicit user provider wins and failure falls back to GPT. Use HyperFrames/Remotion atoms only when composition, titles, captions, transitions, or deterministic rendering are required. Verify duration, dimensions, nonblank frames, playback, and the requested output format.
