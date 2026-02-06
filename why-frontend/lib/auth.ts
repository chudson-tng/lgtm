import { AuthResponse } from './types'

const TOKEN_KEY = 'why_token'

export function setToken(token: string) {
  if (typeof window !== 'undefined') {
    localStorage.setItem(TOKEN_KEY, token)
  }
}

export function getToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(TOKEN_KEY)
  }
  return null
}

export function removeToken() {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(TOKEN_KEY)
  }
}

export function isAuthenticated(): boolean {
  return getToken() !== null
}

export async function signup(email: string, password: string): Promise<AuthResponse> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Signup failed')
  }

  const data = await response.json()
  setToken(data.token)
  return data
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Login failed')
  }

  const data = await response.json()
  setToken(data.token)
  return data
}

export function logout() {
  removeToken()
}
