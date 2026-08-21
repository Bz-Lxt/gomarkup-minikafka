<script setup>
import { inject, onMounted, reactive, ref } from 'vue'
import { api } from '../api.js'

const toast = inject('toast')
const topics = ref([])
const result = ref(null)
const form = reactive({ topic: '', partition: 0, offset: 0, limit: 20 })
const errors = reactive({ topic: '', partition: '', offset: '', limit: '' })

onMounted(async () => {
  topics.value = await api.topics().catch(() => [])
  if (topics.value[0]) form.topic = topics.value[0].name
})

function validate() {
  errors.topic = errors.partition = errors.offset = errors.limit = ''
  if (!form.topic) errors.topic = '请选择 Topic'
  if (!Number.isInteger(Number(form.partition)) || form.partition < 0) errors.partition = '分区须为 ≥0 整数'
  if (!Number.isInteger(Number(form.offset)) || form.offset < 0) errors.offset = 'Offset 须为 ≥0 整数'
  if (!Number.isInteger(Number(form.limit)) || form.limit < 1 || form.limit > 200) errors.limit = 'limit 1–200'
  return !errors.topic && !errors.partition && !errors.offset && !errors.limit
}

async function query() {
  if (!validate()) {
    toast('请修正表单错误', 'error')
    return
  }
  try {
    result.value = await api.messages(form.topic, form.partition, form.offset, form.limit)
  } catch (e) { toast(e.message, 'error') }
}
</script>

<template>
  <div>
    <h2 class="page-title">消息浏览</h2>
    <p class="sub">按 Topic / Partition / Offset 窥探 WAL 中的记录</p>
    <div class="panel" style="margin-bottom:16px">
      <div class="grid-4">
        <div class="field">
          <label>Topic *</label>
          <select v-model="form.topic">
            <option value="" disabled>选择</option>
            <option v-for="t in topics" :key="t.name" :value="t.name">{{ t.name }}</option>
          </select>
          <div class="err">{{ errors.topic }}</div>
        </div>
        <div class="field">
          <label>Partition *</label>
          <input class="input" type="number" min="0" v-model.number="form.partition" />
          <div class="err">{{ errors.partition }}</div>
        </div>
        <div class="field">
          <label>Offset *</label>
          <input class="input" type="number" min="0" v-model.number="form.offset" />
          <div class="err">{{ errors.offset }}</div>
        </div>
        <div class="field">
          <label>Limit *</label>
          <input class="input" type="number" min="1" max="200" v-model.number="form.limit" />
          <div class="err">{{ errors.limit }}</div>
        </div>
      </div>
      <button class="btn" type="button" @click="query">查询</button>
    </div>
    <div class="panel table-wrap" v-if="result">
      <table v-if="result.messages?.length">
        <thead>
          <tr><th>Offset</th><th>时间</th><th>Key</th><th>Value</th></tr>
        </thead>
        <tbody>
          <tr v-for="m in result.messages" :key="m.offset">
            <td class="mono">{{ m.offset }}</td>
            <td class="mono">{{ m.timestamp }}</td>
            <td class="mono">{{ m.key }}</td>
            <td class="mono">{{ m.value }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">该范围内没有消息。next_offset = {{ result.next_offset }}</div>
    </div>
  </div>
</template>
