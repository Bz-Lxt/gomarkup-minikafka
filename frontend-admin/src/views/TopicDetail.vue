<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, formatBytes } from '../api.js'

const route = useRoute()
const router = useRouter()
const detail = ref(null)
const err = ref('')
let timer

async function load() {
  try {
    detail.value = await api.topic(route.params.name)
    err.value = ''
  } catch (e) { err.value = e.message }
}
onMounted(() => { load(); timer = setInterval(load, 3000) })
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div>
    <div class="toolbar">
      <button class="btn ghost" type="button" @click="router.push('/topics')">返回列表</button>
    </div>
    <h2 class="page-title mono">{{ route.params.name }}</h2>
    <p class="sub">每个 Partition 的最早 Offset、LEO 与 Segment 条带</p>
    <p v-if="err" class="err">{{ err }}</p>
    <div class="panel table-wrap" v-if="detail">
      <table>
        <thead>
          <tr>
            <th>Partition</th><th>Earliest</th><th>LEO</th><th>磁盘</th><th>Segments</th><th>条带</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in detail.partitions" :key="p.id">
            <td class="mono">{{ p.id }}</td>
            <td class="mono">{{ p.earliest }}</td>
            <td class="mono">{{ p.leo }}</td>
            <td>{{ formatBytes(p.bytes) }}</td>
            <td>{{ p.segments }}</td>
            <td><div class="seg"><i v-for="n in Math.max(1, p.segments)" :key="n"></i></div></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
