const API_BASE = 'https://kanso.local/api'

interface AuthResult {
  jwt: string
  expiresIn: number
}

export interface User {
  sub: string
  email: string
  name: string
}

function getStoredJWT(): string | null {
  return localStorage.getItem('kanso_jwt')
}

function storeJWT(jwt: string): void {
  localStorage.setItem('kanso_jwt', jwt)
}

function clearJWT(): void {
  localStorage.removeItem('kanso_jwt')
}

async function exchangeGoogleToken(idToken: string): Promise<AuthResult> {
  const res = await fetch(`${API_BASE}/auth/google`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ idToken }),
    credentials: 'include',
  })
  if (!res.ok) throw new Error('Auth failed')
  return res.json()
}

async function getCurrentUser(): Promise<User | null> {
  const jwt = getStoredJWT()
  if (!jwt) return null
  const res = await fetch(`${API_BASE}/auth/me`, {
    headers: { Authorization: `Bearer ${jwt}` },
  })
  if (!res.ok) {
    clearJWT()
    return null
  }
  return res.json()
}

async function refreshJWT(): Promise<string | null> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  })
  if (!res.ok) return null
  const data = await res.json()
  storeJWT(data.jwt)
  return data.jwt
}

async function authenticatedFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const jwt = getStoredJWT()
  const headers = new Headers(options.headers)
  if (jwt) headers.set('Authorization', `Bearer ${jwt}`)

  let res = await fetch(url, { ...options, headers })
  if (res.status === 401) {
    const newJWT = await refreshJWT()
    if (newJWT) {
      headers.set('Authorization', `Bearer ${newJWT}`)
      res = await fetch(url, { ...options, headers })
    }
  }
  return res
}

async function logout(): Promise<void> {
  await fetch(`${API_BASE}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  })
  clearJWT()
}

export const authService = {
  getStoredJWT,
  storeJWT,
  clearJWT,
  exchangeGoogleToken,
  getCurrentUser,
  refreshJWT,
  authenticatedFetch,
  logout,
}

export { authenticatedFetch }
