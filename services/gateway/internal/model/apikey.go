package model

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKey struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"`
	KeyPrefix string     `json:"key_prefix"`
	Status    int        `json:"status"`
	Models    []string   `json:"models"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyStore struct {
	db *pgxpool.Pool
}

func NewAPIKeyStore(db *pgxpool.Pool) *APIKeyStore {
	return &APIKeyStore{db: db}
}

func (s *APIKeyStore) Create(ctx context.Context, userID int64, name string, models []string) (*APIKey, string, error) {
	rawKey := generateAPIKey()
	keyHash := hashKey(rawKey)
	keyPrefix := rawKey[:12]

	var apiKey APIKey
	err := s.db.QueryRow(ctx,
		`INSERT INTO api_keys (user_id, name, key_hash, key_prefix, models)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, key_prefix, status, models, expires_at, created_at`,
		userID, name, keyHash, keyPrefix, models,
	).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
		&apiKey.Status, &apiKey.Models, &apiKey.ExpiresAt, &apiKey.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create api key: %w", err)
	}

	return &apiKey, rawKey, nil
}

func (s *APIKeyStore) ListByUser(ctx context.Context, userID int64) ([]APIKey, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, name, key_prefix, status, models, expires_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		if err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
			&key.Status, &key.Models, &key.ExpiresAt, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *APIKeyStore) Delete(ctx context.Context, id, userID int64) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM api_keys WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}

func (s *APIKeyStore) ValidateKey(ctx context.Context, rawKey string) (*APIKey, error) {
	keyHash := hashKey(rawKey)
	var key APIKey
	err := s.db.QueryRow(ctx,
		`SELECT ak.id, ak.user_id, ak.name, ak.key_prefix, ak.status, ak.models, ak.expires_at, ak.created_at
		 FROM api_keys ak
		 WHERE ak.key_hash = $1 AND ak.status = 1`,
		keyHash,
	).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
		&key.Status, &key.Models, &key.ExpiresAt, &key.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid api key")
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, fmt.Errorf("api key expired")
	}

	return &key, nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "sk-" + hex.EncodeToString(bytes)
}

func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
