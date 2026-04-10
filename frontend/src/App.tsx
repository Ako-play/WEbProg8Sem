import { useEffect, useState } from 'react';
import './App.css';
import { AuthForm } from './components/AuthForm';
import { Dashboard } from './components/Dashboard';
import { User, clearTokens, getCurrentUser, getTokens, login, logout, refresh, register } from './lib/api';

export default function App() {
  const [mode, setMode] = useState<'login' | 'register'>('register');
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    void bootstrap();
  }, []);

  async function bootstrap() {
    const { accessToken, refreshToken } = getTokens();
    if (!accessToken && !refreshToken) {
      setLoading(false);
      return;
    }

    try {
      if (!accessToken && refreshToken) {
        await refresh();
      }
      const currentUser = await getCurrentUser();
      setUser(currentUser);
    } catch {
      clearTokens();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }

  async function handleSubmit(params: { email: string; password: string; username?: string }) {
    const { email, password, username } = params;
    setLoading(true);
    setError('');
    try {
      const response =
        mode === 'login'
          ? await login(email, password)
          : await register(username!.trim(), email, password);
      setUser(response.user);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Ошибка входа в центр киберспорта';
      const low = message.toLowerCase();
      if (mode === 'register' && low.includes('username is already taken')) {
        setError('Этот логин уже занят. Выберите другой.');
      } else if (mode === 'register' && low.includes('email already exists')) {
        setMode('login');
        setError('Пользователь с таким email уже есть. Выполните вход.');
      } else if (mode === 'register' && low.includes('already exists')) {
        setMode('login');
        setError('Пользователь уже зарегистрирован. Выполните вход с тем же email и паролем.');
      } else {
        setError(message);
      }
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout() {
    try {
      await logout();
    } finally {
      clearTokens();
      setUser(null);
      setMode('login');
    }
  }

  if (loading && !user) {
    return (
      <div className="page-shell">
        <div className="page-aura page-aura-left" />
        <div className="page-aura page-aura-right" />
      <div className="wave-layer wave-top" />
      <div className="wave-layer wave-bottom" />
        <div className="wave-cycle" aria-hidden="true" />
        <div className="wave-cycle wave-cycle-delay" aria-hidden="true" />
        <div className="app-frame">
          <div className="card loading-card">Загрузка турнирного центра...</div>
        </div>
      </div>
    );
  }

  return (
    <main className="page-shell">
      <div className="page-aura page-aura-left" />
      <div className="page-aura page-aura-right" />
      <div className="wave-layer wave-top" />
      <div className="wave-layer wave-bottom" />
      <div className="wave-cycle" aria-hidden="true" />
      <div className="wave-cycle wave-cycle-delay" aria-hidden="true" />
      <div className="page-grid" />
      <div className="app-frame">
        {user ? (
          <Dashboard user={user} onLogout={handleLogout} />
        ) : (
          <AuthForm mode={mode} loading={loading} error={error} onModeChange={setMode} onSubmit={handleSubmit} />
        )}
      </div>
    </main>
  );
}
