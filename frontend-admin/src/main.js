import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'
import Dashboard from './views/Dashboard.vue'
import Topics from './views/Topics.vue'
import TopicDetail from './views/TopicDetail.vue'
import Groups from './views/Groups.vue'
import GroupDetail from './views/GroupDetail.vue'
import Messages from './views/Messages.vue'
import Lab from './views/Lab.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Dashboard, meta: { title: '概览' } },
    { path: '/topics', component: Topics, meta: { title: 'Topics' } },
    { path: '/topics/:name', component: TopicDetail, meta: { title: 'Topic' } },
    { path: '/groups', component: Groups, meta: { title: '消费组' } },
    { path: '/groups/:group', component: GroupDetail, meta: { title: '消费组' } },
    { path: '/messages', component: Messages, meta: { title: '消息浏览' } },
    { path: '/lab', component: Lab, meta: { title: '实验室' } },
    { path: '/privacy', component: Dashboard, meta: { title: '隐私' } },
    { path: '/terms', component: Dashboard, meta: { title: '条款' } },
  ],
})

createApp(App).use(router).mount('#app')
