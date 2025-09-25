import './assets/main.css'
import "./assets/icon/ali/iconfont.css"
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import VueVirtualScroller from 'vue-virtual-scroller'

import App from './App.vue'
import router from './router'

const app = createApp(App)

import httpPlugin from './plugins/http';
app.use(httpPlugin);

app.use(VueVirtualScroller)
app.use(createPinia())
app.use(router)

app.mount('#app')
