export type VideoStatus = 'pending' | 'processing' | 'ready' | 'failed'

export interface Video {
  id: string
  title: string
  description: string | null
  status: VideoStatus
  duration_seconds: number | null
  created_at: string
  renditions?: HLSRendition[]
}

export interface HLSRendition {
  resolution: '360p' | '480p' | '720p' | '1080p'
  playlist_url: string
  bandwidth: number
}

export interface User {
  id: string
  email: string
  role: 'user' | 'admin'
}

export interface ApiError {
  error: string
}

export interface WatchHistory {
  video_id: string
  title: string
  watched_at: string
  progress_seconds: number
  duration_seconds: number | null
  completed: boolean
}

export interface AdminVideo extends Video {
  user_email: string
}

export interface AdminStats {
  total_users: number
  total_videos: number
  videos_by_status: {
    pending: number
    processing: number
    ready: number
    failed: number
  }
}

export interface StreamURLResponse {
  master_playlist_url: string
  sas_token: string
}