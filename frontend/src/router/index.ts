import LoginView from '@/views/LoginView.vue'
import LogoutView from '@/views/LogoutView.vue'
import RegisterView from '@/views/RegisterView.vue'
import { api } from '@/api'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', redirect: { name: 'dashboard' }, meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
  { path: '/register', name: 'register', component: RegisterView, meta: { public: true } },
  { path: '/logout', name: 'logout', component: LogoutView, meta: { public: true } },
  { path: '/dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { requiresAuth: true } },
  { path: '/plans/:id', name: 'plan', component: () => import('@/views/PlanView.vue'), meta: { requiresAuth: true } },
]

api.interceptors.response.use(r => r, err => {
  if (err?.response?.status === 401) {
    const { user } = useAuth()
    user.value = null
    if (router.currentRoute.value.meta.requiresAuth) router.push({ name: 'login' })
  }
  return Promise.reject(err)
})



const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const { isAuthed, loading, check } = useAuth()

  if (loading.value) await check()
  if (to.meta.requiresAuth && !isAuthed.value) return next({ name: 'login' })
  next()
})

export default router
