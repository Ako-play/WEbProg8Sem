package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"currencyparser/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,24}$`)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")

type TokenPair struct {
	AccessToken  string           `json:"accessToken"`
	RefreshToken string           `json:"refreshToken"`
	User         repository.User  `json:"user"`
	ExpiresIn    int64            `json:"expiresIn"`
	TokenType    string           `json:"tokenType"`
}

type AuthService struct {
	repo          *repository.AuthRepository
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewAuthService(repo *repository.AuthRepository, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		repo:          repo,
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	if email == "" || password == "" {
		return TokenPair{}, fmt.Errorf("email and password are required")
	}
	if username == "" {
		return TokenPair{}, fmt.Errorf("username is required")
	}
	if !usernamePattern.MatchString(username) {
		return TokenPair{}, fmt.Errorf("username must be 3-24 characters: Latin letters, digits, underscore only")
	}
	if len(password) < 8 {
		return TokenPair{}, fmt.Errorf("password must contain at least 8 characters")
	}

	if _, err := s.repo.FindUserByNormalizedUsername(ctx, username); err == nil {
		return TokenPair{}, fmt.Errorf("this username is already taken")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return TokenPair{}, err
	}

	if _, err := s.repo.FindUserByEmail(ctx, email); err == nil {
		return TokenPair{}, fmt.Errorf("user with this email already exists")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, username, string(hash))
	if err != nil {
		return TokenPair{}, err
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (TokenPair, error) {
	user, err := s.repo.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.parseToken(refreshToken, s.refreshSecret)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}

	hash := hashToken(refreshToken)
	record, err := s.repo.FindRefreshToken(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return TokenPair{}, ErrInvalidToken
		}
		return TokenPair{}, err
	}
	if record.RevokedAt != nil || time.Now().After(record.ExpiresAt) {
		return TokenPair{}, ErrInvalidToken
	}

	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return TokenPair{}, err
	}

	userID, _ := claims["sub"].(string)
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return TokenPair{}, err
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repo.RevokeRefreshToken(ctx, hashToken(refreshToken))
}

func (s *AuthService) ParseAccessToken(token string) (string, error) {
	claims, err := s.parseToken(token, s.accessSecret)
	if err != nil {
		return "", ErrInvalidToken
	}
	userID, _ := claims["sub"].(string)
	if userID == "" {
		return "", ErrInvalidToken
	}
	return userID, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (repository.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *AuthService) issueTokens(ctx context.Context, user repository.User) (TokenPair, error) {
	if err := s.repo.DeleteExpiredRefreshTokens(ctx); err != nil {
		return TokenPair{}, err
	}

	now := time.Now()
	accessToken, err := s.createToken(user.ID, user.Email, now.Add(s.accessTTL), s.accessSecret)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := s.createToken(user.ID, user.Email, now.Add(s.refreshTTL), s.refreshSecret)
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.repo.StoreRefreshToken(ctx, user.ID, hashToken(refreshToken), now.Add(s.refreshTTL)); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) createToken(userID, email string, expiresAt time.Time, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	})
	result, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return result, nil
}

func (s *AuthService) parseToken(token string, secret []byte) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
