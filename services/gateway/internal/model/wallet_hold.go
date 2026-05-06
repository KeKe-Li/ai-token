package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	WalletHoldStatusHeld     = "held"
	WalletHoldStatusCaptured = "captured"
	WalletHoldStatusReleased = "released"
	WalletHoldStatusFailed   = "failed"

	WalletHoldReasonAPIRequest = "api_request"
)

const (
	BillingEventPreAuthHeld   = "preauth_held"
	BillingEventPreAuthFailed = "preauth_failed"
	BillingEventModelUnpriced = "model_unpriced"
	BillingEventPreAuthError  = "preauth_error"
	BillingEventHoldCaptured  = "hold_captured"
	BillingEventHoldReleased  = "hold_released"
	BillingEventCaptureFailed = "capture_failed"
)

type WalletHold struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	APIKeyID       int64      `json:"api_key_id"`
	Model          string     `json:"model"`
	Amount         int64      `json:"amount"`
	CapturedAmount int64      `json:"captured_amount"`
	ReleasedAmount int64      `json:"released_amount"`
	Status         string     `json:"status"`
	Reason         string     `json:"reason"`
	RequestLogID   *int64     `json:"request_log_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type BillingEvent struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	APIKeyID     int64     `json:"api_key_id"`
	RequestLogID *int64    `json:"request_log_id,omitempty"`
	WalletHoldID *int64    `json:"wallet_hold_id,omitempty"`
	EventType    string    `json:"event_type"`
	Model        string    `json:"model,omitempty"`
	Amount       int64     `json:"amount"`
	Balance      int64     `json:"balance"`
	Reserved     int64     `json:"reserved_balance"`
	Available    int64     `json:"available_balance"`
	Status       string    `json:"status"`
	Note         string    `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ReserveBalanceDecision struct {
	Allowed        bool
	RequiredAmount int64
	Balance        int64
	Reserved       int64
	Available      int64
	Note           string
}

func CheckReserveBalance(balance, reservedBalance, requiredAmount int64) ReserveBalanceDecision {
	available := balance - reservedBalance
	if available < 0 {
		available = 0
	}
	decision := ReserveBalanceDecision{
		Allowed:        true,
		RequiredAmount: requiredAmount,
		Balance:        balance,
		Reserved:       reservedBalance,
		Available:      available,
		Note:           "预授权 hold 可创建",
	}
	if requiredAmount <= 0 {
		decision.Note = "无需创建预授权 hold"
		return decision
	}
	if available < requiredAmount {
		decision.Allowed = false
		decision.Note = "可用余额低于预授权金额"
	}
	return decision
}

type HoldCapturePlan struct {
	Allowed        bool
	HoldAmount     int64
	CaptureAmount  int64
	ReleaseAmount  int64
	BalanceBefore  int64
	BalanceAfter   int64
	ReservedBefore int64
	ReservedAfter  int64
	Note           string
}

func PlanHoldCapture(holdAmount, actualCost, balance int64) HoldCapturePlan {
	plan := HoldCapturePlan{
		Allowed:        true,
		HoldAmount:     holdAmount,
		CaptureAmount:  actualCost,
		BalanceBefore:  balance,
		BalanceAfter:   balance - actualCost,
		ReservedBefore: holdAmount,
		ReservedAfter:  0,
		Note:           "hold capture 成功",
	}
	if holdAmount > actualCost {
		plan.ReleaseAmount = holdAmount - actualCost
	}
	if actualCost <= 0 {
		plan.CaptureAmount = 0
		plan.ReleaseAmount = holdAmount
		plan.BalanceAfter = balance
		plan.Note = "实际成本为 0，释放 hold"
		return plan
	}
	if balance < actualCost {
		plan.Allowed = false
		plan.CaptureAmount = 0
		plan.ReleaseAmount = holdAmount
		plan.BalanceAfter = balance
		plan.Note = "余额不足以 capture 实际成本"
	}
	return plan
}

func BuildExpiredHoldReleaseEvent(hold WalletHold, balance, reservedBalance int64) BillingEvent {
	holdID := hold.ID
	reservedAfter := maxInt64(reservedBalance-hold.Amount, 0)
	return BillingEvent{
		UserID:       hold.UserID,
		APIKeyID:     hold.APIKeyID,
		WalletHoldID: &holdID,
		EventType:    BillingEventHoldReleased,
		Model:        hold.Model,
		Amount:       hold.Amount,
		Balance:      balance,
		Reserved:     reservedAfter,
		Available:    balance - reservedAfter,
		Status:       WalletHoldStatusReleased,
		Note:         "过期预授权 hold 自动释放",
	}
}

func NormalizeWalletHoldStatusFilter(raw string) (string, bool) {
	status := strings.TrimSpace(raw)
	if status == "" {
		return "", true
	}
	switch status {
	case WalletHoldStatusHeld, WalletHoldStatusCaptured, WalletHoldStatusReleased, WalletHoldStatusFailed:
		return status, true
	default:
		return "", false
	}
}

func BuildManualHoldReleaseNote(reason string) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return "管理员手动释放异常 hold"
	}
	return "管理员手动释放异常 hold: " + trimmed
}

type WalletHoldStore struct {
	db *pgxpool.Pool
}

func NewWalletHoldStore(db *pgxpool.Pool) *WalletHoldStore {
	return &WalletHoldStore{db: db}
}

func (s *WalletHoldStore) ListAll(ctx context.Context, userID *int64, status string, limit, offset int) ([]WalletHold, error) {
	if userID != nil {
		return s.ListByUser(ctx, *userID, status, limit, offset)
	}
	if status != "" {
		rows, err := s.db.Query(ctx,
			`SELECT id, user_id, api_key_id, request_log_id, model, amount, captured_amount, released_amount, status, reason, created_at, updated_at, expires_at
			 FROM wallet_holds
			 WHERE status = $1
			 ORDER BY created_at DESC, id DESC
			 LIMIT $2 OFFSET $3`,
			status, limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list wallet holds: %w", err)
		}
		defer rows.Close()
		return scanWalletHolds(rows)
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, request_log_id, model, amount, captured_amount, released_amount, status, reason, created_at, updated_at, expires_at
		 FROM wallet_holds
		 ORDER BY created_at DESC, id DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallet holds: %w", err)
	}
	defer rows.Close()
	return scanWalletHolds(rows)
}

func (s *WalletHoldStore) ListByUser(ctx context.Context, userID int64, status string, limit, offset int) ([]WalletHold, error) {
	if status != "" {
		rows, err := s.db.Query(ctx,
			`SELECT id, user_id, api_key_id, request_log_id, model, amount, captured_amount, released_amount, status, reason, created_at, updated_at, expires_at
			 FROM wallet_holds
			 WHERE user_id = $1 AND status = $2
			 ORDER BY created_at DESC, id DESC
			 LIMIT $3 OFFSET $4`,
			userID, status, limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list user wallet holds: %w", err)
		}
		defer rows.Close()
		return scanWalletHolds(rows)
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, request_log_id, model, amount, captured_amount, released_amount, status, reason, created_at, updated_at, expires_at
		 FROM wallet_holds
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list user wallet holds: %w", err)
	}
	defer rows.Close()
	return scanWalletHolds(rows)
}

type walletHoldRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanWalletHolds(rows walletHoldRows) ([]WalletHold, error) {
	var holds []WalletHold
	for rows.Next() {
		var hold WalletHold
		var requestLogID pgtype.Int8
		var modelValue pgtype.Text
		var expiresAt pgtype.Timestamptz
		if err := rows.Scan(
			&hold.ID,
			&hold.UserID,
			&hold.APIKeyID,
			&requestLogID,
			&modelValue,
			&hold.Amount,
			&hold.CapturedAmount,
			&hold.ReleasedAmount,
			&hold.Status,
			&hold.Reason,
			&hold.CreatedAt,
			&hold.UpdatedAt,
			&expiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan wallet hold: %w", err)
		}
		if requestLogID.Valid {
			value := requestLogID.Int64
			hold.RequestLogID = &value
		}
		if modelValue.Valid {
			hold.Model = modelValue.String
		}
		if expiresAt.Valid {
			value := expiresAt.Time
			hold.ExpiresAt = &value
		}
		holds = append(holds, hold)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate wallet holds: %w", err)
	}
	return holds, nil
}

type BillingEventStore struct {
	db *pgxpool.Pool
}

func NewBillingEventStore(db *pgxpool.Pool) *BillingEventStore {
	return &BillingEventStore{db: db}
}

func (s *BillingEventStore) ListAll(ctx context.Context, userID *int64, limit, offset int) ([]BillingEvent, error) {
	if userID != nil {
		return s.ListByUser(ctx, *userID, limit, offset)
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, request_log_id, wallet_hold_id, event_type, model, amount, balance, reserved_balance, available_balance, status, note, created_at
		 FROM billing_events
		 ORDER BY created_at DESC, id DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list billing events: %w", err)
	}
	defer rows.Close()

	return scanBillingEvents(rows)
}

func (s *BillingEventStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]BillingEvent, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, request_log_id, wallet_hold_id, event_type, model, amount, balance, reserved_balance, available_balance, status, note, created_at
		 FROM billing_events
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list user billing events: %w", err)
	}
	defer rows.Close()

	return scanBillingEvents(rows)
}

func (s *BillingEventStore) ListByWalletHold(ctx context.Context, walletHoldID int64, limit, offset int) ([]BillingEvent, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, api_key_id, request_log_id, wallet_hold_id, event_type, model, amount, balance, reserved_balance, available_balance, status, note, created_at
		 FROM billing_events
		 WHERE wallet_hold_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		walletHoldID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallet hold billing events: %w", err)
	}
	defer rows.Close()

	return scanBillingEvents(rows)
}

type billingEventRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanBillingEvents(rows billingEventRows) ([]BillingEvent, error) {
	var events []BillingEvent
	for rows.Next() {
		var event BillingEvent
		var requestLogID pgtype.Int8
		var walletHoldID pgtype.Int8
		var modelValue pgtype.Text
		var note pgtype.Text
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.APIKeyID,
			&requestLogID,
			&walletHoldID,
			&event.EventType,
			&modelValue,
			&event.Amount,
			&event.Balance,
			&event.Reserved,
			&event.Available,
			&event.Status,
			&note,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan billing event: %w", err)
		}
		if requestLogID.Valid {
			value := requestLogID.Int64
			event.RequestLogID = &value
		}
		if walletHoldID.Valid {
			value := walletHoldID.Int64
			event.WalletHoldID = &value
		}
		if modelValue.Valid {
			event.Model = modelValue.String
		}
		if note.Valid {
			event.Note = note.String
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate billing events: %w", err)
	}
	return events, nil
}

func insertBillingEvent(ctx context.Context, tx pgx.Tx, event BillingEvent) (int64, error) {
	var requestLogID any
	if event.RequestLogID != nil {
		requestLogID = *event.RequestLogID
	}
	var walletHoldID any
	if event.WalletHoldID != nil {
		walletHoldID = *event.WalletHoldID
	}

	var id int64
	err := tx.QueryRow(ctx,
		`INSERT INTO billing_events (user_id, api_key_id, request_log_id, wallet_hold_id, event_type, model, amount, balance, reserved_balance, available_balance, status, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id`,
		event.UserID, event.APIKeyID, requestLogID, walletHoldID,
		event.EventType, event.Model, event.Amount, event.Balance,
		event.Reserved, event.Available, event.Status, event.Note,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert billing event: %w", err)
	}
	return id, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
