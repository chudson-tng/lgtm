import { Message, Reply } from './types'
import { getToken } from './auth'

const API_BASE_URL = '/api/v1'

function getHeaders(): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }

  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return headers
}

export async function listMessages(): Promise<Message[]> {
  const response = await fetch(`${API_BASE_URL}/messages`, {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch messages')
  }

  return response.json()
}

export async function getMessage(id: string): Promise<Message> {
  const response = await fetch(`${API_BASE_URL}/messages/${id}`, {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch message')
  }

  return response.json()
}

export async function createMessage(content: string, mediaUrls: string[] = []): Promise<Message> {
  const response = await fetch(`${API_BASE_URL}/messages`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ content, media_urls: mediaUrls }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Failed to create message')
  }

  return response.json()
}

export async function listReplies(messageId: string): Promise<Reply[]> {
  const response = await fetch(`${API_BASE_URL}/messages/${messageId}/replies`, {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch replies')
  }

  return response.json()
}

export async function createReply(messageId: string, content: string, mediaUrls: string[] = []): Promise<Reply> {
  const response = await fetch(`${API_BASE_URL}/messages/${messageId}/replies`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ content, media_urls: mediaUrls }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Failed to create reply')
  }

  return response.json()
}

export async function uploadMedia(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)

  const token = getToken()
  const headers: HeadersInit = {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE_URL}/media`, {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('Failed to upload file')
  }

  const data = await response.json()
  return data.url
}
