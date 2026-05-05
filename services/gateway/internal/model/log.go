package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestLog struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	APIKeyID        int64     `json:"api_key_id"`
	ChannelID       *int64    `json:"channel_id"`
	WalletHoldID    *int64    `json:"wallet_hold_id,omitempty"`
	Model           string    `json:"model"`
	RequestMethod   string    `json:"request_method"`
	RequestPath     string    `json:"request_path"`
	StatusCode      int       `json:"status_code"`
	InputTokens     int       `json:"input_tokens"`
	OutputTokens    int       `json:"output_tokens"`
	Cost            int64     `json:"cost"`
	ReservedAmount  int64     `json:"reserved_amount"`
	LatencyMs       int       `json:"latency_ms"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	IPAddress       string    `json:"ip_address"`
	BillingStatus   string    `json:"billing_status"`
	BillingNote     string    `json:"billing_note,omitempty"`
	EstimatedTokens bool      `json:"estimated_tokens"`
	CreatedAt       time.Time `json:"created_at"`
}

type LogStore struct {
	db *pgxpool.Pool
}

func NewLogStore(db *pgxpool.Pool) *LogStore {
	return &LogStore{db: db}
}

func (s *LogStore) Create(ctx context.Context, log *RequestLog) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO request_logs (user_id, api_key_id, channel_id, wallet_hold_id, model, request_method, request_path, status_code, input_tokens, output_tokens, cost, reserved_amount, latency_ms, error_message, ip_address, billing_status, billing_note, estimated_tokens)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		log.UserID, log.APIKeyID, log.ChannelID, log.WalletHoldID, log.Model, log.RequestMethod, log.RequestPath,
		log.StatusCode, log.InputTokens, log.OutputTokens, log.Cost, log.ReservedAmount, log.LatencyMs,
		log.ErrorMessage, log.IPAddress, log.BillingStatus, log.BillingNote, log.EstimatedTokens,
	)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}
	return nil
}

func (s *LogStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]RequestLog, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, channel_id, wallet_hold_id, model, request_method, request_path, status_code, input_tokens, output_tokens, cost, reserved_amount, latency_ms, error_message, ip_address, billing_status, billing_note, estimated_tokens, created_at
				 FROM request_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.APIKeyID, &l.ChannelID, &l.WalletHoldID, &l.Model,
			&l.RequestMethod, &l.RequestPath, &l.StatusCode, &l.InputTokens,
			&l.OutputTokens, &l.Cost, &l.ReservedAmount, &l.LatencyMs, &l.ErrorMessage,
			&l.IPAddress, &l.BillingStatus, &l.BillingNote, &l.EstimatedTokens, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

type UsageSummary struct {
	Model         string `json:"model"`
	TotalRequests int64  `json:"total_requests"`
	InputTokens   int64  `json:"input_tokens"`
	OutputTokens  int64  `json:"output_tokens"`
	TotalCost     int64  `json:"total_cost"`
}

func (s *LogStore) ListAll(ctx context.Context, limit, offset int) ([]RequestLog, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, channel_id, wallet_hold_id, model, request_method, request_path, status_code, input_tokens, output_tokens, cost, reserved_amount, latency_ms, error_message, ip_address, billing_status, billing_note, estimated_tokens, created_at
				 FROM request_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list logs: %w", err)
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.APIKeyID, &l.ChannelID, &l.WalletHoldID, &l.Model,
			&l.RequestMethod, &l.RequestPath, &l.StatusCode, &l.InputTokens,
			&l.OutputTokens, &l.Cost, &l.ReservedAmount, &l.LatencyMs, &l.ErrorMessage,
			&l.IPAddress, &l.BillingStatus, &l.BillingNote, &l.EstimatedTokens, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

type GlobalStats struct {
	TotalRequests int64 `json:"total_requests"`
	TotalCost     int64 `json:"total_cost"`
	TotalInput    int64 `json:"total_input_tokens"`
	TotalOutput   int64 `json:"total_output_tokens"`
	UniqueUsers   int64 `json:"unique_users"`
}

func (s *LogStore) GetGlobalStats(ctx context.Context, since time.Time) (*GlobalStats, error) {
	var stats GlobalStats
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(cost), 0), COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0), COUNT(DISTINCT user_id)
		 FROM request_logs WHERE created_at >= $1`,
		since,
	).Scan(&stats.TotalRequests, &stats.TotalCost, &stats.TotalInput, &stats.TotalOutput, &stats.UniqueUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return &stats, nil
}

func (s *LogStore) GetUsageSummary(ctx context.Context, userID int64, since time.Time) ([]UsageSummary, error) {
	rows, err := s.db.Query(ctx,
		`SELECT model, COUNT(*) as total_requests, COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0), COALESCE(SUM(cost), 0)
		 FROM request_logs WHERE user_id = $1 AND created_at >= $2
		 GROUP BY model ORDER BY total_requests DESC`,
		userID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}
	defer rows.Close()

	var summaries []UsageSummary
	for rows.Next() {
		var s UsageSummary
		if err := rows.Scan(&s.Model, &s.TotalRequests, &s.InputTokens, &s.OutputTokens, &s.TotalCost); err != nil {
			return nil, fmt.Errorf("failed to scan summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}
