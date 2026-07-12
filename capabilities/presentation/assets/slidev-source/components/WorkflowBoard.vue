<script setup lang="ts">
import { computed, ref } from 'vue'

const activeIndex = ref(0)
const workflows = [
  {
    name: '单图头像 + 动态表情包',
    short: '头像 / 动图',
    base: 'AniWan + LightX64 + Radial',
    description: '先用统一的人设和视觉符号，让群友在每一次互动里都能认出你。',
    points: ['定制带业务标识的专属头像', '真人动态表情，增加群内互动频率', '把陌生群友转化为新一度人脉'],
    value: '日常高频露出',
    signal: '01 / DAILY SIGNAL',
  },
  {
    name: '九宫格 AI 古风表情包',
    short: '九宫格',
    base: 'Qwen-image-edit Q8 + Animate2.1',
    description: '一次生成一整套统一人设的问好、求助、分享与商务寒暄素材。',
    points: ['9 张成套国风卡通形象', '支持批量转动态动图', '反复出现，快速建立系列记忆'],
    value: '系列化记忆',
    signal: '02 / SERIES TAG',
  },
  {
    name: '地域国风立体艺术字',
    short: '艺术字',
    base: 'Nunchaku-Qwenimage + LLM 文字结构',
    description: '围绕城市、节日、行业和热点，主动发布适配同城场景的原创作品。',
    points: ['地标与国风建筑内嵌进 3D 字体', '适配同城、文旅、商家资源群', '把专业实力转成主动私信'],
    value: '热点借势',
    signal: '03 / LOCAL HOOK',
  },
  {
    name: '微观画卷九宫格',
    short: '微观画卷',
    base: '城市文旅 / 节日 / 品牌宣传',
    description: '用卷轴微缩造景做高质感作品集，拉开与普通同行的专业距离。',
    points: ['九宫格批量产出文旅宣传画面', '适配企业、门店与本地品牌', '形成三度人脉的核心爆款素材'],
    value: '高阶作品集',
    signal: '04 / PREMIUM PROOF',
  },
]

const current = computed(() => workflows[activeIndex.value])
</script>

<template>
  <div class="workflow-board">
    <div class="workflow-tabs" role="tablist" aria-label="ComfyUI 工作流选择">
      <button v-for="(workflow, index) in workflows" :key="workflow.name" class="workflow-tab" :class="{ 'workflow-tab--active': activeIndex === index }" role="tab" :aria-selected="activeIndex === index" @click="activeIndex = index">
        <span class="workflow-tab__index">0{{ index + 1 }}</span>
        <span>{{ workflow.short }}</span>
      </button>
    </div>
    <div class="workflow-detail">
      <div class="workflow-detail__head"><span>{{ current.signal }}</span><b>{{ current.value }}</b></div>
      <div class="workflow-detail__main">
        <div class="workflow-copy">
          <h3>{{ current.name }}</h3>
          <p>{{ current.description }}</p>
          <div class="workflow-pulse"><span></span><span></span><span></span><b>ACTIVE</b></div>
        </div>
        <div class="workflow-preview" :class="`workflow-preview--${activeIndex}`" aria-label="当前工作流产出示意">
          <div v-if="activeIndex === 0" class="preview-avatar">
            <div class="preview-avatar__halo"></div>
            <div class="preview-avatar__face"><span>AI</span></div>
            <div class="preview-avatar__signal">IDENTITY / 001</div>
            <i></i><i></i><i></i>
          </div>
          <div v-else-if="activeIndex === 1" class="preview-stickers">
            <div v-for="label in ['问', '好', '谢', '求', '赞', '发', '在', '聊', '约']" :key="label" class="preview-sticker"><span>{{ label }}</span></div>
            <div class="preview-stickers__stamp">9 FRAMES / LOOP</div>
          </div>
          <div v-else-if="activeIndex === 2" class="preview-type">
            <div class="preview-type__word">地域</div>
            <div class="preview-type__sub">LOCAL / FESTIVAL / BRAND</div>
            <div class="preview-type__beam"></div>
          </div>
          <div v-else class="preview-scene">
            <div v-for="cell in 9" :key="cell" class="preview-scene__cell"><i></i><span>{{ cell }}</span></div>
            <div class="preview-scene__stamp">MICRO WORLD / 9:9</div>
          </div>
        </div>
      </div>
      <div class="workflow-points"><div v-for="point in current.points" :key="point"><span class="pulse-mark"></span>{{ point }}</div></div>
      <div class="workflow-detail__foot"><span>适配底层 / {{ current.base }}</span><div class="detail-meter"><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div><strong>COMFYUI READY</strong></div>
    </div>
  </div>
</template>
