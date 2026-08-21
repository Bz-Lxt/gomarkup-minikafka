<script setup>
import { inject, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api.js'

const route = useRoute()
const router = useRouter()
const toast = inject('toast')
const confirmModal = inject('confirmModal')
const detail = ref(null)
const err = ref('')
const form = reactive({ topic: '', to: 'earliest' })
const errors = reactive({ topic: '' })
let timer

async function load() {
  try {
    detail.value = await api.group(route.params.group)
    if (!form.topic && detail.value.partitions?.length) form.topic = detail.value.partitions[0].topic
    err.value = ''
  } catch (e) { err.value = e.message }
}

function lagPct(p) {
  if (!p.leo) return 0
  return Math.min(100, (p.lag / p.leo) * 100)
}

async function reset() {
  errors.topic = ''
  if (!form.topic) {
    errors.topic = '请选择 Topic'
    toast('请修正表单错误', 'error')
    return
  }
  const ok = await confirmModal({
    title: '重置消费位点',
    body: `将 ${route.params.group} 在 ${form.topic} 的位点重置到 ${form.to}。此操作不可撤销。`,
  })
  if (!ok) return
  try {
    await api.reset(route.params.group, { topic: form.topic, to: form.to })
    toast('位点已重置')
    await load()
  } catch (e) { toast(e.message, 'error') }
}

onMounted(() => { load(); timer = setInterval(load, 3000) })
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div>
    <div class="toolbar">
      <button class="btn ghost" type="button" @click="router.push('/groups')">返回列表</button>
    </div>
    <h2 class="page-title mono">{{ route.params.group }}</h2>
    <p class="sub">每分区已提交 Offset、LEO 与 Lag</p>
    <p v-if="err" class="err">{{ err }}</p>
    <div class="panel" style="margin-bottom:16px">
      <div class="grid-2">
        <div class="field">
          <label>Topic *</label>
          <select v-model="form.topic">
            <option value="" disabled>选择 Topic</option>
            <option v-for="t in [...new Set((detail?.partitions || []).map(p => p.topic))]" :key="t" :value="t">{{ t }}</option>
          </select>
          <div class="err">{{ errors.topic }}</div>
        </div>
        <div class="field">
          <label>重置到</label>
          <select v-model="form.to">
            <option value="earliest">earliest</option>
            <option value="latest">latest</option>
          </select>
        </div>
      </div>
      <button class="btn warn" type="button" @click="reset">重置位点</button>
    </div>
    <div class="panel table-wrap" v-if="detail">
      <table>
        <thead>
          <tr>
            <th>Topic</th><th>P</th><th>Committed</th><th>LEO</th><th>Lag</th><th>进度</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in detail.partitions" :key="p.topic + p.partition">
            <td class="mono">{{ p.topic }}</td>
            <td>{{ p.partition }}</td>
            <td class="mono">{{ p.committed }}</td>
            <td class="mono">{{ p.leo }}</td>
            <td class="mono">{{ p.lag }}</td>
            <td><div class="lag"><span :style="{ width: lagPct(p) + '%' }"></span></div></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
