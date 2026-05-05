import { useQuery } from '@tanstack/vue-query'
import axios from 'axios'

const API_URL = '/api'

export const api = axios.create({ baseURL: API_URL })

api.interceptors.response.use(
  (response) => response,
  (error) => {
    return Promise.reject(error)
  }
)

export const useStats = () => {
  return useQuery({ queryKey: ['stats'], queryFn: () => api.get('/stats').then(res => res.data) })
}

export const useJobs = (filters = {}) => {
  return useQuery({ queryKey: ['jobs', filters], queryFn: () => api.get('/jobs', { params: filters }).then(res => res.data) })
}

export const useMatches = (filters = {}) => {
  return useQuery({ queryKey: ['matches', filters], queryFn: () => api.get('/matches', { params: filters }).then(res => res.data) })
}

export const useSites = () => {
  return useQuery({ queryKey: ['sites'], queryFn: () => api.get('/sites').then(res => res.data) })
}

export const useResumes = () => {
  return useQuery({ queryKey: ['resumes'], queryFn: () => api.get('/resumes').then(res => res.data) })
}

export const useMe = () => {
  return useQuery({ queryKey: ['me'], queryFn: () => api.get('/me').then(res => res.data) })
}

export const addSite = (site) => api.post('/sites', site)
export const deleteSite = (id) => api.delete(`/sites/${id}`)
export const uploadResume = (formData) => api.post('/resumes/upload', formData, {
  timeout: 180000
})
export const updateJobStatus = (id, status) => api.patch(`/jobs/${id}`, { status })
export const updateMatch = (id, applied) => api.patch(`/matches/${id}`, { applied })
export const deleteMatch = (id) => api.delete(`/matches/${id}`)
export const updateProfile = (data) => api.patch('/me', data)
export const changePassword = (data) => api.post('/me/change-password', data)
