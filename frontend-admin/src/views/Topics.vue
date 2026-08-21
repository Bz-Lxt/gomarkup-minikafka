<script setup>
import { inject, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, formatBytes } from '../api.js'

const toast = inject('toast')
const router = useRouter()
const list = ref([])
const err = ref('')
const show = ref(false)
const form = reactive({ name: '', partitions: 3 })
const errors = reactive({ name: '', partitions: '' })
let timer

function validate() {
  errors.name = ''
  errors.partitions = ''
  if (!form.name.trim()) errors.name = 'Topic 名称必填'
  else if (!/^[a-zA-Z0-9._-]{1,64}$/.test(form.name)) errors.name = '仅允许字母数字 . _ - ，最长 64'
  const n = Number(form.partitions)
  if (!Number.isInteger(n) || n < 1 || n > 16) errors.partitions = '分区数须为 1–16 的整数'
  return !errors.name && !errors.partitions
}

async function load() {
  try {
    list.value = await api.topics()
    err.value = ''
  } catch (e) { err.value = e.message }
}

async function create() {
  if (!validate()) {
    toast('请修正表单错误', 'error')
    return
  }
  try {
    await api.createTopic({ name: form.name.trim(), partitions: Number(form.partitions) })
    toast(`Topic ${form.name} 已创建`)
    show.value = false
    form.name = ''
    await load()
  } catch (e) { toast(e.message, 'error') }
}

onMounted(() => { load(); timer = setInterval(load, 3000) })
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div>
    <h2 class="page-title">Topics</h2>
    <p class="sub">磁带卷宗 · 创建后分区数不可变更</p>
    <div class="toolbar">
      <button class="btn" type="button" @click="show = !show">创建 Topic</button>
    </div>
    <div v-if="show" class="panel" style="margin-bottom:16px">
      <div class="grid-2">
        <div class="field">
          <label>名称 *</label>
          <input class="input" v-model="form.name" placeholder="orders" />
          <div class="err">{{ errors.name }}</div>
        </div>
        <div class="field">
          <label>分区数 * （1–16）</label>
          <input class="input" type="number" min="1" max="16" v-model.number="form.partitions" />
          <div class="err">{{ errors.partitions }}</div>
        </div>
      </div>
      <button class="btn" type="button" @click="create">保存</button>
    </div>
    <p v-if="err" class="err">{{ err }}</p>
    <div class="panel table-wrap">
      <table v-if="list.length">
        <thead>
          <tr>
            <th>名称</th><th>分区</th><th>消息数</th><th>磁盘</th><th>创建时间</th>
          </tr>
        </thead>
        <tbody>
          <tr class="row-link" v-for="t in list" :key="t.name" @click="router.push(`/topics/${encodeURIComponent(t.name)}`)">
            <td class="mono">{{ t.name }}</td>
            <td>{{ t.partitions }}</td>
            <td class="mono">{{ t.messages }}</td>
            <td>{{ formatBytes(t.bytes) }}</td>
            <td class="mono">{{ t.created_at }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">
        <div class="ring">∅</div>
        深海暂无回波，创建一个 Topic 开始投递。
      </div>
    </div>
  </div>
</template>
