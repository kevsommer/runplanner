import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import Aura from '@primevue/themes/aura' // choose any theme you like
import { polyfill } from 'mobile-drag-drop'
import { scrollBehaviourDragImageTranslateOverride } from 'mobile-drag-drop/scroll-behaviour'
// global styles (PrimeIcons + optional PrimeFlex)
import 'primeicons/primeicons.css'
import 'primeflex/primeflex.css' // optional
import 'mobile-drag-drop/default.css'

polyfill({
  forceApply: true,
  dragImageTranslateOverride: scrollBehaviourDragImageTranslateOverride,
})

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
  },
  locale: {
    firstDayOfWeek: 1,
  },
})

app.use(ToastService)
app.use(ConfirmationService)

app.mount('#app')
