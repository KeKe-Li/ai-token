package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Channel struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Provider      string     `json:"provider"`
	BaseURL       string     `json:"base_url"`
	APIKeyEnc     string     `json:"-"`
	Models        []string   `json:"models"`
	Status        int        `json:"status"`
	Priority      int        `json:"priority"`
	Weight        int        `json:"weight"`
	RateLimit     int        `json:"rate_limit"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ChannelStore struct {
	db *pgxpool.Pool
}

func NewChannelStore(db *pgxpool.Pool) *ChannelStore {
	return &ChannelStore{db: db}
}

func (s *ChannelStore) Create(ctx context.Context, ch *Channel) error {
	err := s.db.QueryRow(ctx,
		`INSERT INTO channels (name, provider, base_url, api_key_enc, models, status, priority, weight, rate_limit)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		ch.Name, ch.Provider, ch.BaseURL, ch.APIKeyEnc, ch.Models,
		ch.Status, ch.Priority, ch.Weight, ch.RateLimit,
	).Scan(&ch.ID, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	return nil
}

func (s *ChannelStore) List(ctx context.Context) ([]Channel, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, provider, base_url, models, status, priority, weight, rate_limit, cooldown_until, created_at, updated_at
		 FROM channels ORDER BY priority DESC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Provider, &ch.BaseURL,
			&ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.RateLimit,
			&ch.CooldownUntil, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (s *ChannelStore) GetByID(ctx context.Context, id int64) (*Channel, error) {
	var ch Channel
	err := s.db.QueryRow(ctx,
		`SELECT id, name, provider, base_url, api_key_enc, models, status, priority, weight, rate_limit, cooldown_until, created_at, updated_at
		 FROM channels WHERE id = $1`, id,
	).Scan(&ch.ID, &ch.Name, &ch.Provider, &ch.BaseURL, &ch.APIKeyEnc,
		&ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.RateLimit,
		&ch.CooldownUntil, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}
	return &ch, nil
}

func (s *ChannelStore) Update(ctx context.Context, ch *Channel) error {
	_, err := s.db.Exec(ctx,
		`UPDATE channels SET name=$1, provider=$2, base_url=$3, api_key_enc=$4, models=$5, status=$6, priority=$7, weight=$8, rate_limit=$9, updated_at=NOW()
		 WHERE id=$10`,
		ch.Name, ch.Provider, ch.BaseURL, ch.APIKeyEnc, ch.Models,
		ch.Status, ch.Priority, ch.Weight, ch.RateLimit, ch.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}
	return nil
}

func (s *ChannelStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `DELETE FROM channels WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}
