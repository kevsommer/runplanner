import { ref, computed } from 'vue'
import type { AxiosError } from 'axios'
import { api } from '../api'

const user = ref(null)
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
  return api.post('/auth/register', { email, password })
}


async function logout() {
  await api.post('/auth/logout')
  user.value = null
}


export function useAuth() {
  return { user, loading, error, isAuthed: computed(() => !!user.value), check, login, register, logout }
}
