<script setup>
import { inject, onMounted, reactive, ref } from 'vue'
import { api } from '../api.js'

const toast = inject('toast')
const topics = ref([])
const produce = reactive({ topic: '', key: '', value: 'hello-from-lab', count: 10 })
const consume = reactive({ topic: '', group: 'lab-group', client_id: 'lab-client', max_messages: 20, auto_commit: true })
const pe = reactive({ topic: '', value: '', count: '' })
const ce = reactive({ topic: '', group: '', max_messages: '' })
const last = ref(null)

onMounted(async () => {
  topics.value = await api.topics().catch(() => [])
  if (topics.value[0]) {
    produce.topic = topics.value[0].name
    consume.topic = topics.value[0].name
  }
})

function vProduce() {
  pe.topic = pe.value = pe.count = ''
  if (!produce.topic) pe.topic = '请选择 Topic'
  if (!produce.value) pe.value = '消息内容必填'
  const n = Number(produce.count)
  if (!Number.isInteger(n) || n < 1 || n > 10000) pe.count = '条数 1–10000'
  return !pe.topic && !pe.value && !pe.count
}

function vConsume() {
  ce.topic = ce.group = ce.max_messages = ''
  if (!consume.topic) ce.topic = '请选择 Topic'
  if (!consume.group.trim()) ce.group = '消费组必填'
  else if (!/^[a-zA-Z0-9._-]{1,64}$/.test(consume.group)) ce.group = '仅允许字母数字 . _ -'
  const n = Number(consume.max_messages)
  if (!Number.isInteger(n) || n < 1 || n > 500) ce.max_messages = '1–500'
  return !ce.topic && !ce.group && !ce.max_messages
}

async function send() {
  if (!vProduce()) { toast('请修正生产表单', 'error'); return }
  try {
    const messages = Array.from({ length: Number(produce.count) }, (_, i) => ({
      key: produce.key || `k-${i}`,
      value: `${produce.value}#${i}`,
    }))
    last.value = await api.produceBatch({ topic: produce.topic, messages })
    toast(`已写入 ${messages.length} 条`)
  } catch (e) { toast(e.message, 'error') }
}

async function pull() {
  if (!vConsume()) { toast('请修正消费表单', 'error'); return }
  try {
    last.value = await api.consume({
      topic: consume.topic,
      group: consume.group.trim(),
      client_id: consume.client_id || 'lab-client',
      max_messages: Number(consume.max_messages),
      auto_commit: consume.auto_commit,
    })
    toast(`拉到 ${(last.value.messages || []).length} 条`)
  } catch (e) { toast(e.message, 'error') }
}
</script>

<template>
  <div>
    <h2 class="page-title">实验室</h2>
    <p class="sub">向 Broker 投递或拉取消息，观察 WAL 与位点变化</p>
    <div class="grid-2">
      <div class="panel">
        <h3 class="display">生产</h3>
        <div class="field"><label>Topic *</label>
          <select v-model="produce.topic">
            <option value="" disabled>选择</option>
            <option v-for="t in topics" :key="t.name" :value="t.name">{{ t.name }}</option>
          </select>
          <div class="err">{{ pe.topic }}</div>
        </div>
        <div class="field"><label>Key（可空，按哈希分区）</label><input class="input" v-model="produce.key" /></div>
        <div class="field"><label>Value *</label><input class="input" v-model="produce.value" /><div class="err">{{ pe.value }}</div></div>
        <div class="field"><label>条数 *</label><input class="input" type="number" min="1" v-model.number="produce.count" /><div class="err">{{ pe.count }}</div></div>
        <button class="btn" type="button" @click="send">批量发送</button>
      </div>
      <div class="panel">
        <h3 class="display">消费</h3>
        <div class="field"><label>Topic *</label>
          <select v-model="consume.topic">
            <option value="" disabled>选择</option>
            <option v-for="t in topics" :key="'c'+t.name" :value="t.name">{{ t.name }}</option>
          </select>
          <div class="err">{{ ce.topic }}</div>
        </div>
        <div class="field"><label>Group *</label><input class="input" v-model="consume.group" /><div class="err">{{ ce.group }}</div></div>
        <div class="field"><label>Client ID</label><input class="input" v-model="consume.client_id" /></div>
        <div class="field"><label>Max *</label><input class="input" type="number" v-model.number="consume.max_messages" /><div class="err">{{ ce.max_messages }}</div></div>
        <label style="display:flex;gap:8px;align-items:center;margin:10px 0">
          <input type="checkbox" v-model="consume.auto_commit" /> 自动提交 Offset
        </label>
        <button class="btn" type="button" @click="pull">拉取</button>
      </div>
    </div>
    <div class="panel" style="margin-top:16px" v-if="last">
      <div class="label">最近响应</div>
      <pre class="msg-body">{{ JSON.stringify(last, null, 2) }}</pre>
    </div>
  </div>
</template>
