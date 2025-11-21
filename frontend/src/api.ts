import type { SearchItem, ImportRequestItem, Job, JobListResponse } from './types'

const API_BASE = import.meta.env.VITE_API_BASE ?? ''

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `Request failed: ${res.status}`)
  }
  return res.json()
}

export async function search(query: string): Promise<SearchItem[]> {
  const data = await request<{ items: SearchItem[] }>(`/api/search?q=${encodeURIComponent(query)}`)
  return data.items ?? []
}

export async function createImport(items: ImportRequestItem[]): Promise<{ jobId: string }> {
  return request('/api/import', {
    method: 'POST',
    body: JSON.stringify({ items }),
  })
}

export async function getJob(id: string): Promise<Job> {
  return request<Job>(`/api/jobs/${id}`)
}

export async function listJobs(): Promise<JobListResponse> {
  return request<JobListResponse>('/api/jobs')
}
