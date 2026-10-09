// Temporary SSR validation harness for specs/011-admin-hide-user-menu (T006/T009/T010/T011).
// Uses the REAL router.js (guards included) with browser-API stubs, and renders
// App.vue server-side with a mocked session, reporting nav links + final route.
import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { VueQueryPlugin } from '@tanstack/vue-query'
import { useAuth } from './src/composables/auth'
import App from './src/App.vue'

const locationStub = {
  pathname: '/',
  search: '',
  hash: '',
  host: 'localhost',
  hostname: 'localhost',
  protocol: 'http:',
  port: '',
  origin: 'http://localhost',
  href: 'http://localhost/',
  replace() {},
  assign() {},
}

globalThis.localStorage = {
  getItem(key) {
    if (key === 'vagai_token') return 'ssr-token'
    if (key === 'vagai_user')
      return JSON.stringify({ id: 1, name: 'SSR Tester', role: process.env.ROLE || 'user', plan: 'free' })
    return null
  },
  setItem() {},
  removeItem() {},
}

globalThis.window = {
  location: locationStub,
  history: { state: { forward: [], back: [] }, pushState() {}, replaceState() {}, go() {}, back() {} },
  addEventListener() {},
  removeEventListener() {},
  matchMedia: () => ({ matches: false, addEventListener() {}, removeEventListener() {} }),
  navigator: { userAgent: 'ssr' },
  document: undefined,
}
globalThis.document = {
  title: '',
  addEventListener() {},
  removeEventListener() {},
  querySelector: () => null,
  querySelectorAll: () => [],
  visibilityState: 'visible',
  hidden: false,
  baseURI: 'http://localhost/',
}
Object.defineProperty(globalThis, 'navigator', {
  value: { userAgent: 'ssr' },
  configurable: true,
  writable: true,
})
Object.defineProperty(globalThis, 'location', {
  value: locationStub,
  configurable: true,
  writable: true,
})
Object.defineProperty(globalThis, 'history', {
  value: globalThis.window.history,
  configurable: true,
  writable: true,
})
window.document = globalThis.document

const { initAuth } = useAuth()
initAuth()

console.error('[ssr-check] iniciando, ROLE =', process.env.ROLE, 'NAV_TO =', process.env.NAV_TO)

setTimeout(() => {
  console.error('SSR TIMEOUT: navegação nunca completou')
  process.exit(2)
}, 8000)

;(async () => {
  try {
    const { default: router } = await import('./src/router.js')

    const app = createSSRApp(App)
    app.use(VueQueryPlugin)
    app.use(router)

    const target = process.env.NAV_TO || '/'
    await router.push(target)
    const finalPath = router.currentRoute.value.path

    let renderOk = true
    let navItems = null
    let hrefs = []
    let pageMarkers = null
    try {
      const html = await renderToString(app)
      const navMatch = html.match(/<nav[\s\S]*?<\/nav>/)
      hrefs = navMatch ? [...navMatch[0].matchAll(/href="([^"]*)"/g)].map((m) => m[1]) : []
      navItems = hrefs.length
      const markers = ['Billing', 'Adicionar Fonte', 'Fontes Cadastradas', 'Palavras-chave de bloqueio', 'Seu Perfil']
      pageMarkers = Object.fromEntries(markers.map((m) => [m, html.includes(m)]))
    } catch (e) {
      renderOk = false
      hrefs = [`<render-error: ${e && e.message ? e.message : e}>`]
    }

    console.log(
      JSON.stringify({
        role: process.env.ROLE || 'user',
        nav_to: target,
        final_path: finalPath,
        redirected: finalPath !== target,
        nav_items: navItems,
        hrefs,
        page_markers: pageMarkers,
        render_ok: renderOk,
      }),
    )
    process.exit(0)
  } catch (e) {
    console.error('SSR ERROR:', e && e.stack ? e.stack : e)
    process.exit(1)
  }
})()
