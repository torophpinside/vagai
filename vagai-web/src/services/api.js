import { useQuery } from '@tanstack/vue-query'
import { toValue } from 'vue'
import axios from 'axios'

const API_URL = '/api'

export const api = axios.create({ baseURL: API_URL })

let isRedirecting = false

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !isRedirecting) {
      isRedirecting = true
      localStorage.removeItem('vagai_token')
      localStorage.removeItem('vagai_user')
      delete api.defaults.headers.common['Authorization']
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const useStats = () => {
  return useQuery({ queryKey: ['stats'], queryFn: () => api.get('/stats').then(res => res.data) })
}

export const useJobs = (filters = {}) => {
  return useQuery({
    queryKey: ['jobs', filters],
    queryFn: () => {
      const params = { ...toValue(filters) }
      if (Array.isArray(params.status)) {
        params.status = params.status.join(',')
      }
      if (Array.isArray(params.site)) {
        params.site = params.site.join(',')
      }
      return api.get('/jobs', { params }).then(res => res.data)
    }
  })
}

export const useMatches = (filters = {}) => {
  return useQuery({
    queryKey: ['matches', filters],
    queryFn: () => {
      const params = { ...toValue(filters) }
      if (Array.isArray(params.site) && params.site.length > 0) {
        params.site = params.site.join(',')
      }
      return api.get('/matches', { params }).then(res => res.data)
    }
  })
}

export const useSites = () => {
  return useQuery({ queryKey: ['sites'], queryFn: () => api.get('/sites').then(res => res.data) })
}

export const useResumes = () => {
  return useQuery({ queryKey: ['resumes'], queryFn: () => api.get('/resumes').then(res => res.data) })
}

export const useResumeAnalyses = () => {
  return useQuery({ queryKey: ['resume-analyses'], queryFn: () => api.get('/resume-analyses').then(res => res.data) })
}


export const useMe = () => {
  return useQuery({ queryKey: ['me'], queryFn: () => api.get('/me').then(res => res.data) })
}

export const usePlans = () => {
  return useQuery({ queryKey: ['plans'], queryFn: () => api.get('/plans').then(res => res.data) })
}

export const addSite = (site) => api.post('/sites', site)
export const updateSite = (id, data) => api.patch(`/sites/${id}`, data)
export const deleteSite = (id) => api.delete(`/sites/${id}`)
export const uploadResume = (formData) => api.post('/resumes/upload', formData, {
  timeout: 180000
})
export const updateJobStatus = (id, status) => api.patch(`/jobs/${id}`, { status })
export const updateMatch = (id, applied) => api.patch(`/matches/${id}`, { applied })
export const deleteMatch = (id) => api.delete(`/matches/${id}`)
export const updateProfile = (data) => api.patch('/me', data)
export const changePassword = (data) => api.post('/me/change-password', data)
export const changePlan = (planSlug) => api.post('/me/plan', { plan_slug: planSlug })
export const deleteResumeAnalysis = (id) => api.delete(`/resume-analyses/${id}`)
