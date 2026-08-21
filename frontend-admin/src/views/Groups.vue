<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api.js'

const router = useRouter()
const list = ref([])
const err = ref('')
let timer

async function load() {
  try {
    list.value = await api.groups()
    err.value = ''
  } catch (e) { err.value = e.message }
}
onMounted(() => { load(); timer = setInterval(load, 3000) })
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div>
    <h2 class="page-title">消费组</h2>
    <p class="sub">位点进度与 Lag · 点击进入分区明细</p>
    <p v-if="err" class="err">{{ err }}</p>
    <div class="panel table-wrap">
      <table v-if="list.length">
        <thead>
          <tr>
            <th>Group</th><th>成员</th><th>Topics</th><th>Lag 合计</th>
          </tr>
        </thead>
        <tbody>
          <tr class="row-link" v-for="g in list" :key="g.group" @click="router.push(`/groups/${encodeURIComponent(g.group)}`)">
            <td class="mono">{{ g.group }}</td>
            <td>{{ (g.members || []).join(', ') || '—' }}</td>
            <td class="mono">{{ (g.topics || []).join(', ') }}</td>
            <td class="mono" :style="{ color: g.lag_total > 0 ? 'var(--amber)' : 'var(--cyan)' }">{{ g.lag_total }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">
        <div class="ring">∅</div>
        尚无消费组。到「实验室」拉取一次消息即可登记。
      </div>
    </div>
  </div>
</template>
