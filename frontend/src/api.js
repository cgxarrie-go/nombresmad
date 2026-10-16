import axios from 'axios'

const API_BASE_URL = (process.env.REACT_APP_API_URL || '/api').replace(/\/+$/, '')
const TOKEN_KEY = 'nombresmad_token'

export function getToken() {
  try {
    return localStorage.getItem(TOKEN_KEY) || ''
  } catch {
    return ''
  }
}

export function setToken(token) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    // storage can be blocked; the in-memory request still fails closed
  }
}

export const api = axios.create({ baseURL: API_BASE_URL })

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (err) => {
    const url = String(err?.config?.url || '')
    const isLogin = url.endsWith('/login')
    if (err?.response?.status === 401 && !isLogin) {
      setToken('')
      window.dispatchEvent(new Event('nombresmad-unauthorized'))
    }
    return Promise.reject(err)
  }
)

export const PAGE_SIZE = 25

export async function fetchItems(query) {
  const params = {
    page: query.page,
    sort: query.sort,
    order: query.order,
  }
  if (query.pageSize) params.pageSize = query.pageSize
  if (query.q) params.q = query.q
  if (query.name) params.name = query.name
  if (query.size) params.size = query.size
  if (query.group) params.group = query.group
  if (query.woodType) params.woodType = query.woodType
  if (query.hasPicture === 'true' || query.hasPicture === 'false') {
    params.hasPicture = query.hasPicture
  }
  if (query.delivered === 'true' || query.delivered === 'false') {
    params.delivered = query.delivered
  }
  if (query.deliveredWithBox === 'true' || query.deliveredWithBox === 'false') {
    params.deliveredWithBox = query.deliveredWithBox
  }
  const { data } = await api.get('/items', { params })
  return {
    items: data.items || [],
    page: data.page,
    pageSize: data.pageSize,
    total: data.total,
    totalPages: data.totalPages,
  }
}

export async function fetchOptions() {
  const { data } = await api.get('/options')
  return {
    woodTypes: data.woodTypes || [],
    groups: data.groups || [],
    sizes: data.sizes || [],
  }
}

export async function fetchItem(id) {
  const { data } = await api.get(`/items/${id}`)
  return data
}

export async function createItem(item) {
  const { data } = await api.post('/items', item)
  return data
}

export async function updateItem(id, item) {
  const { data } = await api.put(`/items/${id}`, item)
  return data
}

export async function deleteItem(id) {
  await api.delete(`/items/${id}`)
}

export function pictureSrc(item) {
  if (!item?.id || !item.picture) return ''
  return `${API_BASE_URL}/items/${item.id}/picture?v=${encodeURIComponent(item.picture)}`
}

export async function uploadPicture(id, file) {
  const body = new FormData()
  body.append('file', file)
  const { data } = await api.post(`/items/${id}/picture`, body)
  return data
}

export async function deletePicture(id) {
  const { data } = await api.delete(`/items/${id}/picture`)
  return data
}

export async function importData() {
  await api.post('/import')
}

export async function migrateDb() {
  await api.post('/migrate')
}

export async function fetchSession() {
  if (!getToken()) {
    const error = new Error('unauthorized')
    error.response = { status: 401 }
    throw error
  }
  const { data } = await api.get('/session')
  return { username: data.username }
}

export async function login(username, password) {
  const { data } = await api.post('/login', { username, password })
  if (!data?.token) {
    const error = new Error('missing token')
    error.response = { data: 'El servidor no devolvió un token de sesión' }
    throw error
  }
  setToken(data.token)
  return { username: data.username }
}

export async function logout() {
  setToken('')
  try {
    await api.post('/logout')
  } catch {
    // the token is already cleared locally
  }
}

export function errorMessage(err) {
  const data = err?.response?.data
  if (typeof data === 'string' && data.trim()) return data.trim()
  if (err?.message === 'Network Error') return 'No hay conexión con el servidor'
  return 'No se pudo completar la operación'
}
