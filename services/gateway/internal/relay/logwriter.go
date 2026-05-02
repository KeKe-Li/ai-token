package relay

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBLogWriter struct {
	db *pgxpool.Pool
}

func NewDBLogWriter(db *pgxpool.Pool) *DBLogWriter {
	return &DBLogWriter{db: db}
}

func (w *DBLogWriter) WriteLog(ctx context.Context, record *UsageRecord) error {
	pricing := GetDefaultPricing(record.Model)
	record.Cost = CalculateCost(pricing, record.InputTokens, record.OutputTokens)

	_, err := w.db.Exec(ctx,
		`INSERT INTO request_logs (user_id, api_key_id, channel_id, model, request_method, request_path, status_code, input_tokens, output_tokens, cost, latency_ms, error_message, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		record.UserID, record.APIKeyID, record.ChannelID, record.Model,
		record.Method, record.Path, record.StatusCode,
		record.InputTokens, record.OutputTokens, record.Cost,
		record.LatencyMs, record.Error, record.IP,
	)
	if err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	if record.Cost > 0 && record.UserID > 0 {
		_, err := w.db.Exec(ctx,
			`UPDATE users SET balance = balance - $1, used_amount = used_amount + $1, request_count = request_count + 1, updated_at = NOW()
			 WHERE id = $2 AND balance >= $1`,
			record.Cost, record.UserID,
		)
		if err != nil {
			log.Printf("余额扣减失败 user=%d cost=%d: %v", record.UserID, record.Cost, err)
		}
	}

	return nil
}
