async function parse(res) {
  const text = await res.text()
  let body = {}
  try { body = text ? JSON.parse(text) : {} } catch { body = { error: { message: text } } }
  if (!res.ok) {
    const msg = body.error?.message || `请求失败 (${res.status})`
    const err = new Error(msg)
    err.details = body.error?.details
    throw err
  }
  return body.data !== undefined ? body.data : body
}

export const api = {
  health: () => fetch('/health').then(parse),
  metrics: () => fetch('/api/v1/metrics').then(parse),
  topics: () => fetch('/api/v1/topics').then(parse),
  topic: (name) => fetch(`/api/v1/topics/${encodeURIComponent(name)}`).then(parse),
  createTopic: (payload) => fetch('/api/v1/topics', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  produce: (payload) => fetch('/api/v1/produce', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  produceBatch: (payload) => fetch('/api/v1/produce/batch', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  consume: (payload) => fetch('/api/v1/consume', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  groups: () => fetch('/api/v1/groups').then(parse),
  group: (name) => fetch(`/api/v1/groups/${encodeURIComponent(name)}`).then(parse),
  commit: (group, payload) => fetch(`/api/v1/groups/${encodeURIComponent(group)}/commit`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  reset: (group, payload) => fetch(`/api/v1/groups/${encodeURIComponent(group)}/reset`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  }).then(parse),
  messages: (topic, partition, offset, limit) => fetch(
    `/api/v1/topics/${encodeURIComponent(topic)}/messages?partition=${partition}&offset=${offset}&limit=${limit}`,
  ).then(parse),
}

export function beijingNow() {
  const d = new Date(Date.now() + 8 * 3600 * 1000)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getUTCFullYear()}-${p(d.getUTCMonth() + 1)}-${p(d.getUTCDate())} ${p(d.getUTCHours())}:${p(d.getUTCMinutes())}:${p(d.getUTCSeconds())}`
}

export function formatBytes(n) {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(1)} MB`
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`
}
