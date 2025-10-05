package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saimonsiddique/blog-api/internal/domain"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	tokenHash := hashToken(token)

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	tokenHash := hashToken(token)

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var rt domain.RefreshToken
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	return &rt, nil
}

func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

func (r *AuthRepository) DeleteUserRefreshTokens(ctx context.Context, userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	_, err := r.db.Exec(ctx, query)
	return err
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
