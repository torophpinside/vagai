import { ref } from 'vue'
import { api } from '../services/api'

const user = ref(null)
const isAuthenticated = ref(false)

const TOKEN_KEY = 'vagai_token'
const USER_KEY = 'vagai_user'
const STALE_KEYS = ['vagai_resume_draft']

function quotaError() {
  const err = new Error('Sessão autenticada, mas não foi possível salvá-la: armazenamento local cheio. Limpe os dados do site e tente novamente.')
  err.code = 'QUOTA_EXCEEDED'
  return err
}

function persistSession(token, userData) {
  const payload = JSON.stringify(userData)
  try {
    localStorage.setItem(TOKEN_KEY, token)
    localStorage.setItem(USER_KEY, payload)
    return
  } catch (err) {
    const cleaned = STALE_KEYS.filter((key) => {
      try {
        if (localStorage.getItem(key) !== null) {
          localStorage.removeItem(key)
          return true
        }
      } catch {
        /* storage bloqueado */
      }
      return false
    })
    if (cleaned.length > 0) {
      try {
        localStorage.setItem(TOKEN_KEY, token)
        localStorage.setItem(USER_KEY, payload)
        return
      } catch {
        /* quota ainda excedida */
      }
    }
    console.warn('Falha ao persistir sessão no localStorage:', err)
    throw quotaError()
  }
}

function clearStoredSession() {
  try {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  } catch {
    /* storage bloqueado */
  }
}

function initAuth() {
  let token = null
  let rawUser = null
  try {
    token = localStorage.getItem(TOKEN_KEY)
    rawUser = localStorage.getItem(USER_KEY)
  } catch {
    return
  }
  if (!token || !rawUser) return

  try {
    user.value = JSON.parse(rawUser)
    isAuthenticated.value = true
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`
  } catch {
    clearStoredSession()
  }
}

async function login(email, password) {
  const response = await api.post('/auth/login', { email, password })
  const { token, user: userData, organization } = response.data

  persistSession(token, { ...userData, organization })
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`

  user.value = { ...userData, organization }
  isAuthenticated.value = true

  return response.data
}

async function register(name, email, password, organization) {
  const response = await api.post('/auth/register', { name, email, password, organization })
  const { token, user: userData, organization: orgData } = response.data

  persistSession(token, { ...userData, organization: orgData })
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`

  user.value = { ...userData, organization: orgData }
  isAuthenticated.value = true

  return response.data
}

function logout() {
  clearStoredSession()
  delete api.defaults.headers.common['Authorization']
  user.value = null
  isAuthenticated.value = false
}

function updateUser(userData) {
  const current = user.value || {}
  const updated = { ...current, ...userData }
  try {
    localStorage.setItem(USER_KEY, JSON.stringify(updated))
  } catch (err) {
    console.warn('Não foi possível persistir o perfil atualizado:', err)
  }
  user.value = updated
}

export function useAuth() {
  return {
    user,
    isAuthenticated,
    initAuth,
    login,
    register,
    logout,
    updateUser,
  }
}
