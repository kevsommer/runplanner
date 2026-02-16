import { ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import type { AxiosError, AxiosResponse } from 'axios'

type UseApiOptions<T, A extends any[] = []> = {
  exec: (...args: A) => Promise<AxiosResponse<T>>
  onSuccess?: (response: AxiosResponse<T>) => void
  onError?: (error: unknown) => void
  showErrorToast?: boolean
  successToast?: string
}

export function useApi<T = any, A extends any[] = []>(options: UseApiOptions<T, A>) {
  const toast = useToast()
  const loading = ref(false)
  const error = ref<unknown>(null)
  const showErrorToast = options.showErrorToast ?? true

  async function exec(...args: A) {
    loading.value = true
    error.value = null
    try {
      const response = await options.exec(...args)
      if (options.successToast) {
        toast.add({
          severity: 'success',
          summary: 'Success',
          detail: options.successToast,
          life: 3000,
        })
      }
      options.onSuccess?.(response)
    } catch (e) {
      error.value = e
      if (showErrorToast) {
        const axiosError = e as AxiosError<{ message?: string }>
        toast.add({
          severity: 'error',
          summary: 'Error',
          detail: axiosError.response?.data?.message || axiosError.message || 'Something went wrong',
          life: 5000,
        })
      }
      options.onError?.(e)
    } finally {
      loading.value = false
    }
  }

  return { exec, loading, error }
}
