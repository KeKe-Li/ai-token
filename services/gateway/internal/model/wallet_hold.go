package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
