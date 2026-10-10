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

export const useAnalyticsMatching = (params = {}) => {
  return useQuery({
    queryKey: ['analytics', 'matching', toValue(params)],
    queryFn: () => api.get('/v1/analytics/matching-distribution', { params: toValue(params) }).then(res => res.data)
  })
}

export const useAnalyticsResume = () => {
  return useQuery({
    queryKey: ['analytics', 'resume'],
    queryFn: () => api.get('/v1/analytics/resume-insights').then(res => res.data)
  })
}

export const useAnalyticsJobTrends = (params = {}) => {
  return useQuery({
    queryKey: ['analytics', 'job-trends', toValue(params)],
    queryFn: () => api.get('/v1/analytics/job-trends', { params: toValue(params) }).then(res => res.data)
  })
}

export const useAnalyticsUserMetrics = (userId) => {
  return useQuery({
    queryKey: ['analytics', 'user-metrics', userId],
    queryFn: () => api.get(`/v1/analytics/user/${userId}`).then(res => res.data),
    enabled: userId != null
  })
}

export const useAnalyticsInterviewPrep = () => {
  return useQuery({
    queryKey: ['analytics', 'interview-prep'],
    queryFn: () => api.get('/v1/analytics/interview-prep').then(res => res.data)
  })
}

export const useAnalyticsAIQuality = () => {
  return useQuery({
    queryKey: ['analytics', 'ai-quality'],
    queryFn: () => api.get('/v1/analytics/ai-quality').then(res => res.data)
  })
}

export const useAnalyticsTimeToHire = () => {
  return useQuery({
    queryKey: ['analytics', 'time-to-hire'],
    queryFn: () => api.get('/v1/analytics/time-to-hire').then(res => res.data)
  })
}

export const useAnalyticsSkillsGap = (params = {}) => {
  return useQuery({
    queryKey: ['analytics', 'skills-gap', toValue(params)],
    queryFn: () => api.get('/v1/analytics/skills-gap', { params: toValue(params) }).then(res => res.data)
  })
}

export const useAnalyticsFunnel = () => {
  return useQuery({
    queryKey: ['analytics', 'funnel'],
    queryFn: () => api.get('/v1/analytics/funnel').then(res => res.data)
  })
}

export const useAnalyticsGeographic = () => {
  return useQuery({
    queryKey: ['analytics', 'geographic'],
    queryFn: () => api.get('/v1/analytics/geographic').then(res => res.data)
  })
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

export const useSites = (options = {}) => {
  return useQuery({ queryKey: ['sites'], queryFn: () => api.get('/sites').then(res => res.data), ...options })
}

export const useResumes = () => {
  return useQuery({ queryKey: ['resumes'], queryFn: () => api.get('/resumes').then(res => res.data) })
}

export const useResumeAnalyses = () => {
  return useQuery({ queryKey: ['resume-analyses'], queryFn: () => api.get('/resume-analyses').then(res => res.data) })
}

export const useInterviewPreps = () => {
  return useQuery({ queryKey: ['interview-preps'], queryFn: () => api.get('/interview-prep').then(res => res.data) })
}


export const useMe = () => {
  return useQuery({ queryKey: ['me'], queryFn: () => api.get('/me').then(res => res.data) })
}

export const usePlans = () => {
  return useQuery({ queryKey: ['plans'], queryFn: () => api.get('/plans').then(res => res.data) })
}

export const useAdminPlans = () => {
  return useQuery({ queryKey: ['plans', 'admin'], queryFn: () => api.get('/admin/plans').then(res => res.data) })
}

export const adminCreatePlan = (data) => api.post('/admin/plans', data)
export const adminUpdatePlan = (id, data) => api.patch(`/admin/plans/${id}`, data)

export const useNegativeKeywords = (options = {}) => {
  return useQuery({ queryKey: ['negative-keywords'], queryFn: () => api.get('/negative-keywords').then(res => res.data), ...options })
}

export const listSites = () => api.get('/sites').then(res => res.data)
export const addSite = (site) => api.post('/sites', site)
export const updateSite = (id, data) => api.patch(`/sites/${id}`, data)
export const deleteSite = (id) => api.delete(`/sites/${id}`)
export const uploadResume = (formData) => api.post('/resumes/upload', formData, {
  timeout: 240000
})
export const extractJob = (url) => api.post('/jobs/extract', { url }).then(res => res.data)
export const createJob = (data) => api.post('/jobs', data)
export const updateJobStatus = (id, status) => api.patch(`/jobs/${id}`, { status })
export const updateMatch = (id, applied) => api.patch(`/matches/${id}`, { applied })
export const deleteMatch = (id) => api.delete(`/matches/${id}`)
export const rematchMatches = () => api.post('/matches/rematch').then(res => res.data)
export const updateProfile = (data) => api.patch('/me', data)
export const changePassword = (data) => api.post('/me/change-password', data)
export const changePlan = (planSlug) => api.post('/me/plan', { plan_slug: planSlug })
export const updateNegativeKeywords = (keywords) => api.put('/negative-keywords', { keywords }).then(res => res.data)
export const deleteResumeAnalysis = (id) => api.delete(`/resume-analyses/${id}`)

export const parseResume = (formData) => api.post('/resumes/parse', formData, { timeout: 240000 })
export const createResume = (data) => api.post('/resumes', data)
export const deleteResume = (id) => api.delete(`/resumes/${id}`)
export const getResumeData = (id) => api.get(`/resumes/${id}/data`).then(res => res.data)
export const updateResumeData = (id, data) => api.put(`/resumes/${id}/data`, data)
export const generateResumePDF = (id) => api.post(`/resumes/${id}/generate-pdf`, {}, { responseType: 'blob' })

export const generateInterviewPrep = (matchId) => api.post('/interview-prep', { match_id: matchId }, { timeout: 240000 }).then(res => res.data)
export const generateSpontaneousPrep = (data) => api.post('/interview-prep/spontaneous', data, { timeout: 180000 }).then(res => res.data)
export const listInterviewPreps = (filter = {}) => api.get('/interview-prep', { params: filter }).then(res => res.data)
export const getInterviewPrep = (id) => api.get(`/interview-prep/${id}`).then(res => res.data)
export const updateQuestionStatus = (prepId, questionId, status) => api.patch(`/interview-prep/${prepId}/questions/${questionId}`, { status }).then(res => res.data)
export const saveQuestionAnswer = (prepId, questionId, answer, selfAssessment, timeSpentSeconds = 0) => api.post(`/interview-prep/${prepId}/questions/${questionId}/answers`, { answer, self_assessment: selfAssessment, time_spent_seconds: timeSpentSeconds }).then(res => res.data)
export const verifyInterviewPrep = (prepId) => api.post(`/interview-prep/${prepId}/verify`).then(res => res.data)
