import axios from 'axios'

export const api = axios.create({ baseURL: '/api' })

export const PAGE_SIZE = 25

export async function fetchItems(query) {
  const params = {
    page: query.page,
    sort: query.sort,
    order: query.order,
  }
  if (query.q) params.q = query.q
  if (query.group) params.group = query.group
  if (query.woodType) params.woodType = query.woodType
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

export async function importData() {
  await api.post('/import')
}

export async function migrateDb() {
  await api.post('/migrate')
}

export function errorMessage(err) {
  const data = err?.response?.data
  if (typeof data === 'string' && data.trim()) return data.trim()
  if (err?.message === 'Network Error') return 'No hay conexión con el servidor'
  return 'No se pudo completar la operación'
}
