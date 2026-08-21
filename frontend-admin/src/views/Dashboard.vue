<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { api, formatBytes } from '../api.js'

const metrics = ref(null)
const err = ref('')
const produceHist = ref(Array(40).fill(0))
const consumeHist = ref(Array(40).fill(0))
let timer

function path(values, color) {
  const w = 600, h = 160, pad = 8
  const max = Math.max(1, ...values)
  const pts = values.map((v, i) => {
    const x = pad + (i / (values.length - 1)) * (w - pad * 2)
    const y = h - pad - (v / max) * (h - pad * 2)
    return `${x},${y}`
  })
  return { d: `M ${pts.join(' L ')}`, color, w, h }
}

async function load() {
  try {
    const m = await api.metrics()
    metrics.value = m
    produceHist.value = [...produceHist.value.slice(1), m.produce_rate || 0]
    consumeHist.value = [...consumeHist.value.slice(1), m.consume_rate || 0]
    err.value = ''
  } catch (e) {
    err.value = e.message
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 3000)
})
onUnmounted(() => clearInterval(timer))

const p = () => path(produceHist.value, '#3ee0c5')
const c = () => path(consumeHist.value, '#8b7cff')
</script>

<template>
  <div>
    <h2 class="page-title">声呐概览</h2>
    <p class="sub">Broker 吞吐、磁盘与活跃连接 · 每 3 秒刷新</p>
    <p v-if="err" class="err">{{ err }}</p>
    <div v-if="metrics" class="grid-4">
      <div class="tile" v-for="(item, i) in [
        { label: '消息总量', value: metrics.messages_total, hint: '跨所有 Partition' },
        { label: '生产速率', value: Number(metrics.produce_rate).toFixed(1), hint: 'msg / s' },
        { label: '消费速率', value: Number(metrics.consume_rate).toFixed(1), hint: 'msg / s' },
        { label: '磁盘占用', value: formatBytes(metrics.bytes_on_disk), hint: `${metrics.topics} Topics · ${metrics.partitions} Partitions` },
      ]" :key="i" :style="{ animationDelay: i * 80 + 'ms' }">
        <div class="label">{{ item.label }}</div>
        <div class="value">{{ item.value }}</div>
        <div class="hint">{{ item.hint }}</div>
      </div>
    </div>
    <div class="grid-2" style="margin-top:16px">
      <div class="panel">
        <div class="label">生产曲线</div>
        <svg class="chart" viewBox="0 0 600 160" preserveAspectRatio="none">
          <polyline :points="p().d.replace('M ','').replaceAll(' L ',' ')" fill="none" :stroke="p().color" stroke-width="2" />
        </svg>
      </div>
      <div class="panel">
        <div class="label">消费曲线</div>
        <svg class="chart" viewBox="0 0 600 160" preserveAspectRatio="none">
          <polyline :points="c().d.replace('M ','').replaceAll(' L ',' ')" fill="none" :stroke="c().color" stroke-width="2" />
        </svg>
      </div>
    </div>
    <div class="grid-3" style="margin-top:16px">
      <div class="tile" v-if="metrics">
        <div class="label">活跃连接</div>
        <div class="value">{{ metrics.active_connections }}</div>
      </div>
      <div class="tile" v-if="metrics">
        <div class="label">消费组</div>
        <div class="value">{{ metrics.groups }}</div>
      </div>
      <div class="tile" v-if="metrics">
        <div class="label">运行时间</div>
        <div class="value">{{ metrics.uptime_sec }}s</div>
      </div>
    </div>
  </div>
</template>
