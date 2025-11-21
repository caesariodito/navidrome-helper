export type SearchItemType = 'album' | 'song'

export interface SearchItem {
  id: string
  type: SearchItemType
  title: string
  artist: string
  albumId?: string
  albumTitle?: string
  coverUrl: string
  tracks?: number
  duration?: number
}

export interface ImportRequestItem {
  id: string
  type: SearchItemType
  title: string
  artist: string
  albumId?: string
  albumTitle?: string
  coverUrl: string
}

export interface JobItem {
  sourceId: string
  sourceType: string
  title: string
  artist: string
  album: string
  coverUrl: string
  status: string
  message: string
}

export interface JobLog {
  jobId: string
  message: string
  createdAt: string
}

export interface Job {
  id: string
  status: string
  phase: string
  message: string
  progress: number
  artist?: string
  album?: string
  createdAt: string
  updatedAt?: string
  finishedAt?: string
  items?: JobItem[]
  logs?: JobLog[]
}

export interface JobListResponse {
  jobs: Job[]
}
