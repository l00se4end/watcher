import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './assets/router'
import './assets/style.css'
import App from './App.vue'

const pinia = createPinia()
const app = createApp(App)


app.use(router)
app.use(pinia)
app.mount("#app")
