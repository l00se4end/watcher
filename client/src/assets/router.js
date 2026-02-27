import { createMemoryHistory, createRouter } from 'vue-router'

import Index from '../views/Index.vue'
import Auth from '../views/Auth.vue'
import Profile from '../views/Profile.vue'

const routes = [
    { path: '/', component: Index },
    { path: '/auth', component: Auth },
    { path: '/me', component: Profile }
]

export const router = createRouter({
  history: createMemoryHistory(),
  routes,
})