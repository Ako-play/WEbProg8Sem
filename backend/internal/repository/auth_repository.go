package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (User, error) {
	user := User{}
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, username, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, password_hash, created_at
	`, uuid.NewString(), email, username, passwordHash).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) FindUserByNormalizedUsername(ctx context.Context, username string) (User, error) {
	user := User{}
	if err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE lower(trim(username)) = lower(trim($1))
	`, username).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by username: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	user := User{}
	if err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, id string) (User, error) {
	user := User{}
	if err := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.NewString(), userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) FindRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	record := RefreshTokenRecord{}
	if err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&record.ID, &record.UserID, &record.TokenHash, &record.ExpiresAt, &record.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshTokenRecord{}, ErrNotFound
		}
		return RefreshTokenRecord{}, fmt.Errorf("find refresh token: %w", err)
	}
	return record, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked_at IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return nil
}
