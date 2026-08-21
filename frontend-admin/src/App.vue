<script setup>
import { computed, onMounted, onUnmounted, provide, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { beijingNow } from './api.js'

const route = useRoute()
const router = useRouter()
const open = ref(false)
const clock = ref(beijingNow())
const toasts = ref([])
let timer

const title = computed(() => route.meta.title || '观测台')

function toast(message, type = 'ok') {
  const id = Date.now() + Math.random()
  toasts.value.push({ id, message, type })
  setTimeout(() => dismiss(id), 5000)
}
function dismiss(id) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

const modal = ref(null)
function confirmModal(opts) {
  return new Promise((resolve) => {
    modal.value = { ...opts, resolve }
  })
}
function closeModal(ok) {
  modal.value?.resolve(ok)
  modal.value = null
}

provide('toast', toast)
provide('confirmModal', confirmModal)

onMounted(() => {
  timer = setInterval(() => { clock.value = beijingNow() }, 1000)
})
onUnmounted(() => clearInterval(timer))

function goPrivacy() {
  toast('Coming Soon')
  router.push('/')
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar" :class="{ open }">
      <div class="brand">
        <div class="sonar"></div>
        <div>
          <h1>MiniKafka</h1>
          <p>DEEP-SEA OBSERVATORY</p>
        </div>
      </div>
      <nav class="nav" @click="open = false">
        <router-link to="/">概览</router-link>
        <router-link to="/topics">Topics</router-link>
        <router-link to="/groups">消费组</router-link>
        <router-link to="/messages">消息浏览</router-link>
        <router-link to="/lab">实验室</router-link>
      </nav>
      <div style="margin-top:auto;color:var(--muted);font-size:12px">
        <a href="/privacy" @click.prevent="goPrivacy">隐私</a>
        ·
        <a href="/terms" @click.prevent="goPrivacy">条款</a>
      </div>
    </aside>
    <div class="main">
      <header class="topbar">
        <button class="menu-btn" @click="open = !open">菜单</button>
        <div>
          <div class="display" style="font-size:18px">{{ title }}</div>
          <div class="clock mono">GMT+8 · {{ clock }}</div>
        </div>
        <div style="display:flex;align-items:center;gap:8px;color:var(--muted);font-size:12px">
          <span class="refresh-dot"></span> 3s 声呐扫描
        </div>
      </header>
      <div class="content">
        <router-view />
      </div>
    </div>
    <div class="toast-stack">
      <div v-for="t in toasts" :key="t.id" class="toast" :class="{ error: t.type === 'error' }">
        <span>{{ t.message }}</span>
        <button type="button" @click="dismiss(t.id)">×</button>
      </div>
    </div>
    <div v-if="modal" class="modal-mask" @click.self="closeModal(false)">
      <div class="modal">
        <h3>{{ modal.title }}</h3>
        <p>{{ modal.body }}</p>
        <div class="modal-actions">
          <button class="btn ghost" type="button" @click="closeModal(false)">取消</button>
          <button class="btn danger" type="button" @click="closeModal(true)">确认</button>
        </div>
      </div>
    </div>
  </div>
</template>
