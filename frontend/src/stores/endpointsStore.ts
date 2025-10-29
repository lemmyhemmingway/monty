import { create } from 'zustand'
import axios from 'axios'

interface Endpoint {
  id: string
  url: string
  check_type: string
  interval: number
  timeout: number
  created_at: string
  uptime?: number
}

interface EndpointsState {
  endpoints: Endpoint[]
  loading: boolean
  error: string | null
  fetchEndpoints: () => Promise<void>
  getHttpEndpoints: () => Endpoint[]
}

export const useEndpointsStore = create<EndpointsState>((set, get) => ({
  endpoints: [],
  loading: false,
  error: null,

  fetchEndpoints: async () => {
    set({ loading: true, error: null })
    try {
      const response = await axios.get('/api/endpoints')
      set({ endpoints: response.data, loading: false })
    } catch (error) {
      set({ error: 'Failed to fetch endpoints', loading: false })
    }
  },

  getHttpEndpoints: () => {
    return get().endpoints.filter(ep => ep.check_type === 'http')
  },
}))
