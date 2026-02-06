export interface User {
  id: string
  email: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  user_id: string
  content: string
  media_urls: string[]
  created_at: string
  updated_at: string
}

export interface Reply {
  id: string
  message_id: string
  user_id: string
  content: string
  media_urls: string[]
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}
