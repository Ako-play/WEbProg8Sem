import { useEffect, useState } from 'react';
import type { Competition, EsportsArticle, GameDigest, User } from '../lib/api';
import { getEsportsDigest } from '../lib/api';
import { userDisplayName } from '../lib/displayName';
import { esportsDigestFallback } from '../data/esportsDigestFallback';

type DashboardProps = {
  user: User;
  onLogout: () => Promise<void>;
};

type TournamentOverlay = { gameLabel: string; competition: Competition };

type ArticleSectionKind = 'patch' | 'news' | 'mode';

type ArticleReadOverlay = {
  gameLabel: string;
  article: EsportsArticle;
  sectionKind: ArticleSectionKind;
};

const sectionEyebrow: Record<ArticleSectionKind, string> = {
  patch: 'патчноут',
  news: 'новость',
  mode: 'режим игры',
};

function ClickableArticleRows({
  items,
  onOpen,
  empty,
  hint,
}: {
  items: EsportsArticle[];
  onOpen: (a: EsportsArticle) => void;
  empty: string;
  hint: string;
}) {
  if (!items.length) {
    return <p className="muted esports-empty">{empty}</p>;
  }
  return (
    <ul className="esports-patch-click-list">
      {items.map((item) => (
        <li key={item.id}>
          <button type="button" className="esports-patch-row-btn" onClick={() => onOpen(item)}>
            <span className="esports-article-date">{item.date}</span>
            <span className="esports-patch-row-title">{item.title}</span>
            <span className="esports-patch-row-hint">{hint}</span>
          </button>
        </li>
      ))}
    </ul>
  );
}

function TournamentOverlayView({
  detail,
  onClose,
}: {
  detail: TournamentOverlay;
  onClose: () => void;
}) {
  const { competition, gameLabel } = detail;
  const matches = competition.matches ?? [];

  return (
    <div className="esports-overlay" role="dialog" aria-modal="true" aria-labelledby="tournament-overlay-title">
      <div className="esports-overlay-backdrop" onClick={onClose} aria-hidden="true" />
      <div className="card esports-overlay-panel">
        <div className="esports-overlay-toolbar">
          <button type="button" className="primary-button esports-overlay-back" onClick={onClose}>
            ← Назад к ленте
          </button>
          <button type="button" className="ghost-button" onClick={onClose}>
            Закрыть
          </button>
        </div>
        <p className="eyebrow">{gameLabel}</p>
        <h2 id="tournament-overlay-title">{competition.title}</h2>
        <p className="muted esports-overlay-meta">
          {competition.region} · {competition.dates} · {competition.prizePool} · <span className="esports-pill esports-pill-inline">{competition.status}</span>
        </p>
        <h3 className="esports-overlay-section-title">Результаты матчей (кто кого обыграл)</h3>
        {matches.length === 0 ? (
          <p className="muted">Сетка и матчи для этого турнира ещё не опубликованы — следите за анонсами.</p>
        ) : (
          <div className="esports-matches-table-wrap">
            <table className="esports-matches-table">
              <thead>
                <tr>
                  <th>Этап</th>
                  <th>Игрок 1 / команда</th>
                  <th>Игрок 2 / команда</th>
                  <th>Победитель</th>
                  <th>Счёт</th>
                </tr>
              </thead>
              <tbody>
                {matches.map((m, i) => (
                  <tr key={`${competition.id}-${i}-${m.phase}`}>
                    <td>{m.phase}</td>
                    <td>{m.teamA}</td>
                    <td>{m.teamB}</td>
                    <td className="esports-winner-cell">{m.winner}</td>
                    <td>{m.score}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function ArticleReadOverlayView({ detail, onClose }: { detail: ArticleReadOverlay; onClose: () => void }) {
  const { article, gameLabel, sectionKind } = detail;
  const hasBody = Boolean(article.body?.trim());
  const kindLabel = sectionEyebrow[sectionKind];

  return (
    <div className="esports-overlay" role="dialog" aria-modal="true" aria-labelledby="article-read-title">
      <div className="esports-overlay-backdrop" onClick={onClose} aria-hidden="true" />
      <div className="card esports-overlay-panel esports-overlay-panel-article">
        <div className="esports-overlay-toolbar">
          <button type="button" className="primary-button esports-overlay-back" onClick={onClose}>
            ← Назад к ленте
          </button>
          <button type="button" className="ghost-button" onClick={onClose}>
            Закрыть
          </button>
        </div>
        <p className="eyebrow">
          {gameLabel} · {kindLabel}
        </p>
        <time className="esports-overlay-date" dateTime={article.date}>
          {article.date}
        </time>
        <h2 id="article-read-title">{article.title}</h2>
        <p className="esports-overlay-lead">{article.summary}</p>
        {hasBody ? (
          <div className="esports-overlay-body">{article.body}</div>
        ) : (
          <p className="muted">Развёрнутого текста пока нет — доступно только краткое описание.</p>
        )}
      </div>
    </div>
  );
}

function GameColumn({
  digest,
  onOpenTournament,
  onOpenArticle,
}: {
  digest: GameDigest;
  onOpenTournament: (c: Competition) => void;
  onOpenArticle: (a: EsportsArticle, kind: ArticleSectionKind) => void;
}) {
  const patches = digest.patchArticles;
  const featured = patches.slice(0, 2);
  const olderPatches = patches.slice(2);

  return (
    <div className={`card esports-game-column esports-game-${digest.game}`}>
      <div className="esports-game-head">
        <h2 className="esports-game-title">{digest.label}</h2>
        <p className="muted esports-game-lead">Соревнования, патчи, новости и режимы — нажмите на материал, чтобы открыть полный текст.</p>
      </div>

      <section className="esports-block">
        <h3 className="esports-block-title">Соревнования и лиги</h3>
        <p className="muted esports-block-note">Нажмите на строку турнира, чтобы открыть таблицу матчей и победителей.</p>
        <ul className="esports-comp-list">
          {digest.competitions.map((c) => (
            <li key={c.id}>
              <button type="button" className="esports-comp-row esports-comp-button" onClick={() => onOpenTournament(c)}>
                <div className="esports-comp-main">
                  <strong>{c.title}</strong>
                  <span className="muted">{c.region} · {c.dates}</span>
                </div>
                <div className="esports-comp-meta">
                  <span className="esports-pill">{c.status}</span>
                  <span className="esports-prize">{c.prizePool}</span>
                  <span className="esports-open-hint">Открыть →</span>
                </div>
              </button>
            </li>
          ))}
        </ul>
      </section>

      <section className="esports-block">
        <h3 className="esports-block-title">Последние патчи</h3>
        <p className="muted esports-block-note">Нажмите на карточку, чтобы прочитать материал целиком.</p>
        <div className="esports-featured-patches">
          {featured.map((p) => (
            <button key={p.id} type="button" className="esports-patch-featured esports-patch-featured-btn" onClick={() => onOpenArticle(p, 'patch')}>
              <header>
                <span className="esports-patch-badge">Патч</span>
                <time dateTime={p.date}>{p.date}</time>
              </header>
              <h4>{p.title}</h4>
              <p className="esports-patch-lead">{p.summary}</p>
              <span className="esports-patch-tap-hint">Открыть статью</span>
            </button>
          ))}
        </div>
      </section>

      <section className="esports-block">
        <h3 className="esports-block-title">Ранее: список патчей</h3>
        <ClickableArticleRows items={olderPatches} onOpen={(a) => onOpenArticle(a, 'patch')} empty="Нет записей." hint="Нажмите, чтобы открыть" />
      </section>

      <section className="esports-block">
        <h3 className="esports-block-title">Новости</h3>
        <p className="muted esports-block-note">Каждая новость открывается в отдельном окне — как патчи.</p>
        <ClickableArticleRows
          items={digest.newsArticles}
          onOpen={(a) => onOpenArticle(a, 'news')}
          empty="Новостей пока нет."
          hint="Нажмите, чтобы открыть новость"
        />
      </section>

      <section className="esports-block">
        <h3 className="esports-block-title">Типы и режимы игры</h3>
        <p className="muted esports-block-note">Материалы о форматах матчей тоже можно раскрыть и прочитать.</p>
        <ClickableArticleRows
          items={digest.gameTypeArticles}
          onOpen={(a) => onOpenArticle(a, 'mode')}
          empty="Материалов пока нет."
          hint="Нажмите, чтобы открыть"
        />
      </section>
    </div>
  );
}

export function Dashboard({ user, onLogout }: DashboardProps) {
  const [digest, setDigest] = useState<typeof esportsDigestFallback | null>(null);
  const [loading, setLoading] = useState(true);
  const [usedFallback, setUsedFallback] = useState(false);
  const [tournamentDetail, setTournamentDetail] = useState<TournamentOverlay | null>(null);
  const [articleRead, setArticleRead] = useState<ArticleReadOverlay | null>(null);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const data = await getEsportsDigest();
        if (!cancelled) {
          setDigest(data);
          setUsedFallback(false);
        }
      } catch {
        if (!cancelled) {
          setDigest(esportsDigestFallback);
          setUsedFallback(true);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!tournamentDetail && !articleRead) {
      return;
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setTournamentDetail(null);
        setArticleRead(null);
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [tournamentDetail, articleRead]);

  function closeOverlays() {
    setTournamentDetail(null);
    setArticleRead(null);
  }

  return (
    <div className="dashboard dashboard-esports">
      {tournamentDetail ? <TournamentOverlayView detail={tournamentDetail} onClose={closeOverlays} /> : null}
      {articleRead ? <ArticleReadOverlayView detail={articleRead} onClose={closeOverlays} /> : null}

      <header className="topbar esports-topbar">
        <section className="card topbar-hero">
          <div className="badge">Rift Pulse · Киберспорт</div>
          <p className="eyebrow">Dota 2 и Counter-Strike 2</p>
          <h2>Добро пожаловать, {userDisplayName(user)}</h2>
          <p className="muted hero-copy">
            Турниры с матчами, патчи, новости и режимы — всё открывается по нажатию; «Назад к ленте», «Закрыть» или Esc возвращают на главную ленту.
          </p>
          <div className="hero-tags">
            <span>Турниры</span>
            <span>Патчноуты</span>
            <span>Новости</span>
            <span>Режимы</span>
          </div>
        </section>

        <aside className="hero-side">
          <section className="card insight-card">
            <p className="eyebrow">Сессия</p>
            <p className="muted">Вы вошли в ленту киберспортивного контента.</p>
            <p className="muted esports-session-email" title={user.email}>
              Email: {user.email}
            </p>
            {usedFallback ? (
              <p className="esports-offline-hint">Показаны локальные данные: сервер API сейчас недоступен.</p>
            ) : null}
            <button className="ghost-button esports-logout" onClick={() => void onLogout()} type="button">
              Выйти
            </button>
          </section>
        </aside>
      </header>

      {loading ? (
        <div className="card esports-loading">Загрузка материалов…</div>
      ) : digest ? (
        <div className="esports-two-col">
          <GameColumn
            digest={digest.dota2}
            onOpenTournament={(c) => setTournamentDetail({ gameLabel: digest.dota2.label, competition: c })}
            onOpenArticle={(a, kind) => setArticleRead({ gameLabel: digest.dota2.label, article: a, sectionKind: kind })}
          />
          <GameColumn
            digest={digest.cs2}
            onOpenTournament={(c) => setTournamentDetail({ gameLabel: digest.cs2.label, competition: c })}
            onOpenArticle={(a, kind) => setArticleRead({ gameLabel: digest.cs2.label, article: a, sectionKind: kind })}
          />
        </div>
      ) : null}
    </div>
  );
}
