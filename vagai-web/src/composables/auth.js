import { ref } from 'vue'
import { api } from '../services/api'

const user = ref(null)
const isAuthenticated = ref(false)

function initAuth() {
  const token = localStorage.getItem('vagai_token')
  const userData = localStorage.getItem('vagai_user')
  if (token && userData) {
    user.value = JSON.parse(userData)
    isAuthenticated.value = true
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`
  }
}

async function login(email, password) {
  const response = await api.post('/auth/login', { email, password })
  const { token, user: userData, organization } = response.data

  localStorage.setItem('vagai_token', token)
  localStorage.setItem('vagai_user', JSON.stringify({ ...userData, organization }))
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`

  user.value = { ...userData, organization }
  isAuthenticated.value = true

  return response.data
}

async function register(name, email, password, organization) {
  const response = await api.post('/auth/register', { name, email, password, organization })
  const { token, user: userData, organization: orgData } = response.data

  localStorage.setItem('vagai_token', token)
  localStorage.setItem('vagai_user', JSON.stringify({ ...userData, organization: orgData }))
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`

  user.value = { ...userData, organization: orgData }
  isAuthenticated.value = true

  return response.data
}

function logout() {
  localStorage.removeItem('vagai_token')
  localStorage.removeItem('vagai_user')
  delete api.defaults.headers.common['Authorization']
  user.value = null
  isAuthenticated.value = false
}

export function useAuth() {
  return {
    user,
    isAuthenticated,
    initAuth,
    login,
    register,
    logout,
  }
}
