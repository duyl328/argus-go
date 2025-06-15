import './assets/main.css'
import "./assets/icon/ali/iconfont.css"

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)

import httpPlugin from './plugins/http';
app.use(httpPlugin);


app.use(createPinia())
app.use(router)

app.mount('#app')
