import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './pages/Dashboard.vue'
import Jobs from './pages/Jobs.vue'
import JobAdd from './pages/JobAdd.vue'
import Matches from './pages/Matches.vue'
import Applied from './pages/Applied.vue'
import Analysis from './pages/Analysis.vue'
import Settings from './pages/Settings.vue'
import Team from './pages/Team.vue'
import Billing from './pages/Billing.vue'
import Login from './pages/auth/Login.vue'
import Register from './pages/auth/Register.vue'

const routes = [
  { path: '/login', name: 'Login', component: Login, meta: { guest: true } },
  { path: '/register', name: 'Register', component: Register, meta: { guest: true } },
  { path: '/', name: 'Dashboard', component: Dashboard, meta: { requiresAuth: true } },
  { path: '/jobs', name: 'Jobs', component: Jobs, meta: { requiresAuth: true } },
  { path: '/jobs/new', name: 'JobAdd', component: JobAdd, meta: { requiresAuth: true } },
  { path: '/matches', name: 'Matches', component: Matches, meta: { requiresAuth: true } },
  { path: '/applied', name: 'Applied', component: Applied, meta: { requiresAuth: true } },
  { path: '/analysis', name: 'Analysis', component: Analysis, meta: { requiresAuth: true } },
  { path: '/settings', name: 'Settings', component: Settings, meta: { requiresAuth: true } },
  { path: '/settings/team', name: 'Team', component: Team, meta: { requiresAuth: true } },
  { path: '/settings/billing', name: 'Billing', component: Billing, meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('vagai_token')
  const isAuth = !!token

  if (to.meta.requiresAuth && !isAuth) {
    next({ name: 'Login' })
  } else if (to.meta.guest && isAuth) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
