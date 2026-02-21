import { ref, computed } from 'vue'
import type { AxiosError } from 'axios'
import { api } from '../api'

interface AuthUser {
  user: {
    activePlanId?: string
    [key: string]: any
  }
  [key: string]: any
}

const user = ref<AuthUser | null>(null)
const loading = ref(true)
const error = ref(null)


async function check() {
  loading.value = true
  try {
    const { data } = await api.get('/auth/me')
    user.value = data
    error.value = null
  } catch (e: AxiosError | any) {
    if (e?.response?.status === 401) user.value = null
    else error.value = e
  } finally { loading.value = false }
}


async function login(email: string, password: string) {
  await api.post('/auth/login', { email, password })
  await check()
}

async function register(email: string, password: string) {
  await api.post('/auth/register', { email, password })
  await check()
}


async function logout() {
  return api.post('/auth/logout').then(() => user.value = null)
}


function setActivePlanId(id: string | null) {
  if (user.value?.user) {
    user.value.user.activePlanId = id ?? undefined
  }
}

export function useAuth() {
  return { user, loading, error, isAuthed: computed(() => !!user.value), check, login, register, logout, setActivePlanId }
}
