export type User = {
  id: string;
  email: string;
  /** Логин; у старых сессий может отсутствовать до обновления API */
  username?: string;
  createdAt: string;
};

export type AuthResponse = {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType: string;
  user: User;
};

export type EsportsArticle = {
  id: string;
  title: string;
  summary: string;
  date: string;
  body?: string;
};

export type MatchResult = {
  phase: string;
  teamA: string;
  teamB: string;
  winner: string;
  score: string;
};

export type Competition = {
  id: string;
  title: string;
  region: string;
  prizePool: string;
  dates: string;
  status: string;
  matches?: MatchResult[] | null;
};

export type GameDigest = {
  game: string;
  label: string;
  competitions: Competition[];
  patchArticles: EsportsArticle[];
  newsArticles: EsportsArticle[];
  gameTypeArticles: EsportsArticle[];
};

export type EsportsDigestResponse = {
  dota2: GameDigest;
  cs2: GameDigest;
};

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1';
const ACCESS_TOKEN_KEY = 'rift-pulse-access-token';
const REFRESH_TOKEN_KEY = 'rift-pulse-refresh-token';

export function getTokens() {
  return {
    accessToken: localStorage.getItem(ACCESS_TOKEN_KEY),
    refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY),
  };
}

export function setTokens(payload: Pick<AuthResponse, 'accessToken' | 'refreshToken'>) {
  localStorage.setItem(ACCESS_TOKEN_KEY, payload.accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, payload.refreshToken);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

async function rawRequest<T>(path: string, init: RequestInit = {}, auth = true): Promise<T> {
  const { accessToken } = getTokens();
  const headers = new Headers(init.headers);
  headers.set('Content-Type', 'application/json');
  if (auth && accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  let response: Response;
  try {
    response = await fetch(`${API_URL}${path}`, {
      ...init,
      headers,
    });
  } catch {
    throw new Error('Failed to fetch: backend недоступен. Проверьте, что API запущен на http://localhost:8080');
  }

  if (!response.ok) {
    const data = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(data.error ?? 'Request failed');
  }

  return response.json() as Promise<T>;
}

async function request<T>(path: string, init: RequestInit = {}, auth = true): Promise<T> {
  try {
    return await rawRequest<T>(path, init, auth);
  } catch (error) {
    if (!(error instanceof Error) || !auth || !error.message.toLowerCase().includes('invalid access token')) {
      throw error;
    }

    await refresh();
    return rawRequest<T>(path, init, auth);
  }
}

export async function register(username: string, email: string, password: string) {
  const response = await request<AuthResponse>(
    '/auth/register',
    {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    },
    false,
  );
  setTokens(response);
  return response;
}

export async function login(email: string, password: string) {
  const response = await request<AuthResponse>(
    '/auth/login',
    {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    },
    false,
  );
  setTokens(response);
  return response;
}

export async function refresh() {
  const { refreshToken } = getTokens();
  if (!refreshToken) {
    throw new Error('Refresh token not found');
  }

  const response = await rawRequest<AuthResponse>(
    '/auth/refresh',
    {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    },
    false,
  );
  setTokens(response);
  return response;
}

export async function logout() {
  const { refreshToken } = getTokens();
  if (refreshToken) {
    await request('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    });
  }
  clearTokens();
}

export function getCurrentUser() {
  return request<User>('/auth/me');
}

export function getEsportsDigest() {
  return request<EsportsDigestResponse>('/esports/digest');
}
