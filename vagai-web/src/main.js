import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import router from './router'
import './style.css'
import { useAuth } from './composables/auth'

const app = createApp(App)

const { initAuth } = useAuth()
initAuth()

app.use(VueQueryPlugin)
app.use(router)
app.mount('#app')
