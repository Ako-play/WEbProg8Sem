import { ChangeEvent, FormEvent, useState } from 'react';

type AuthFormProps = {
  mode: 'login' | 'register';
  loading: boolean;
  error: string;
  onModeChange: (mode: 'login' | 'register') => void;
  onSubmit: (params: { email: string; password: string; username?: string }) => Promise<void>;
};

export function AuthForm({ mode, loading, error, onModeChange, onSubmit }: AuthFormProps) {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  function handleUsernameChange(event: ChangeEvent<HTMLInputElement>) {
    setUsername(event.target.value);
  }

  function handleEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setEmail(event.target.value);
  }

  function handlePasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setPassword(event.target.value);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (mode === 'register') {
      await onSubmit({ email, password, username: username.trim() });
    } else {
      await onSubmit({ email, password });
    }
  }

  return (
    <section className="auth-layout">
      <div className="auth-intro">
        <div className="badge">Neon Rift Control</div>
        <p className="eyebrow">Esports command analytics</p>
        <h1>Rift Pulse</h1>
        <p className="auth-copy muted">
          Центр аналитики турнирных серий: сравнение матчапов, оценка темпа команды и мониторинг динамики результатов.
        </p>

        <div className="auth-ticker" aria-hidden="true">
          <span>Playoffs Heatmap</span>
          <span>Form Momentum</span>
          <span>Head-to-head Index</span>
        </div>

        <div className="auth-feature-grid">
          <article className="feature-tile">
            <span className="feature-index">01</span>
            <strong>Draft & map trends</strong>
            <p className="muted">Сводка по выбору карт, винрейтам и динамике состава в реальном времени.</p>
          </article>
          <article className="feature-tile">
            <span className="feature-index">02</span>
            <strong>Session-safe workspace</strong>
            <p className="muted">Защищенный вход, хранение истории запросов и быстрый доступ к данным аналитика.</p>
          </article>
        </div>
      </div>

      <section className="card auth-card">
        <div className="auth-card-header">
          <div>
            <p className="eyebrow">Control access</p>
            <h2>{mode === 'login' ? 'Вход в центр Rift Pulse' : 'Регистрация аналитика'}</h2>
          </div>
          <div className="auth-status-dot" aria-hidden="true" />
        </div>

        <div className="auth-switcher">
          <button className={mode === 'login' ? 'active' : ''} onClick={() => onModeChange('login')} type="button">
            Вход
          </button>
          <button className={mode === 'register' ? 'active' : ''} onClick={() => onModeChange('register')} type="button">
            Регистрация
          </button>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          {mode === 'register' ? (
            <label>
              Логин
              <input
                type="text"
                value={username}
                onChange={handleUsernameChange}
                autoComplete="username"
                placeholder="латиница, цифры, _ · 3–24 символа"
                minLength={3}
                maxLength={24}
                pattern="[a-zA-Z0-9_]{3,24}"
                title="3–24 символа: латинские буквы, цифры, подчёркивание"
                required
              />
            </label>
          ) : null}
          <label>
            Email
            <input type="email" value={email} onChange={handleEmailChange} required />
          </label>
          <label>
            Пароль
            <div className="password-field">
              <input type={showPassword ? 'text' : 'password'} value={password} onChange={handlePasswordChange} minLength={8} required />
              <button
                className="password-toggle"
                type="button"
                aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                onClick={() => setShowPassword((prev) => !prev)}
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M12 5C6 5 2.1 9 1 12c1.1 3 5 7 11 7s9.9-4 11-7c-1.1-3-5-7-11-7zm0 11a4 4 0 1 1 0-8 4 4 0 0 1 0 8z" />
                  <circle cx="12" cy="12" r="2.2" />
                </svg>
              </button>
            </div>
          </label>
          {error ? <div className="error-box">{error}</div> : null}
          <button className="primary-button" disabled={loading} type="submit">
            {loading ? 'Подождите...' : mode === 'login' ? 'Запустить панель' : 'Создать профиль'}
          </button>
        </form>
      </section>
    </section>
  );
}
