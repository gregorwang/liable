import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './styles/design-system.css'
import './style.css'
import App from './App.vue'
import router from './router'
import { initTraceHotkey } from './utils/traceNotice'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ElementPlus)

initTraceHotkey()

app.mount('#app')
