import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import 'element-plus/dist/index.css'
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles.css'

const app = createApp(App)
for (const [name, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(name, component)
}
app.use(ElementPlus, { locale: zhCn })
app.use(router)
app.mount('#app')
