package relay

import (
	"context"
	"fmt"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type holdCaptureState struct {
	Hold            model.WalletHold
	Plan            model.HoldCapturePlan
	BalanceBefore   int64
	BalanceAfter    int64
	ReservedBefore  int64
	HasSnapshot     bool
	NeedsFinalizing bool
}

type DBLogWriter struct {
	db              *pgxpool.Pool
	pricingProvider PricingProvider
}

func NewDBLogWriter(db *pgxpool.Pool, pricingProvider ...PricingProvider) *DBLogWriter {
	var provider PricingProvider
	if len(pricingProvider) > 0 {
		provider = pricingProvider[0]
	}
	return &DBLogWriter{db: db, pricingProvider: provider}
}

func (w *DBLogWriter) WriteLog(ctx context.Context, record *UsageRecord) error {
	pricing, pricingStatus, pricingNote := ResolvePricing(ctx, w.pricingProvider, record.Model)
	record.Cost = CalculateCost(pricing, record.InputTokens, record.OutputTokens)

	billingStatus := pricingStatus
	billingNote := pricingNote
	var balanceBefore int64
	var balanceAfter int64
	hasBalanceSnapshot := false
	var holdState *holdCaptureState

	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin billing tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if billingStatus != BillingStatusUnpriced && record.WalletHoldID > 0 {
		state, status, note := w.prepareHoldCapture(ctx, tx, record)
		holdState = state
		billingStatus = status
		billingNote = note
		if state != nil && state.HasSnapshot {
			balanceBefore = state.BalanceBefore
			balanceAfter = state.BalanceAfter
			hasBalanceSnapshot = true
		}
	} else if billingStatus != BillingStatusUnpriced {
		var rowsAffected int64
		var chargeErr error
		if record.Cost > 0 && record.UserID > 0 {
			if err := tx.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1 FOR UPDATE`, record.UserID).Scan(&balanceBefore); err != nil {
				chargeErr = err
			} else {
				hasBalanceSnapshot = true
				balanceAfter = balanceBefore
				if balanceBefore >= record.Cost {
					result, err := tx.Exec(ctx,
						`UPDATE users SET balance = balance - $1, used_amount = used_amount + $1, request_count = request_count + 1, updated_at = NOW()
						 WHERE id = $2 AND balance >= $1`,
						record.Cost, record.UserID,
					)
					chargeErr = err
					if err == nil {
						rowsAffected = result.RowsAffected()
						if rowsAffected > 0 {
							balanceAfter = balanceBefore - record.Cost
						}
					}
				}
			}
		}

		billingStatus, billingNote, err = ClassifyChargeResult(record.Cost, record.UserID, rowsAffected, chargeErr)
		if err != nil {
			record.BillingStatus = billingStatus
			record.BillingNote = billingNote
		}
	}

	if record.EstimatedTokens && billingNote != "" {
		billingNote += "; 流式 token 为估算值"
	} else if record.EstimatedTokens {
		billingNote = "流式 token 为估算值"
	}
	record.BillingStatus = billingStatus
	record.BillingNote = billingNote

	var requestLogID int64
	insertErr := tx.QueryRow(ctx,
		`INSERT INTO request_logs (user_id, api_key_id, channel_id, wallet_hold_id, model, request_method, request_path, status_code, input_tokens, output_tokens, cost, reserved_amount, latency_ms, error_message, ip_address, billing_status, billing_note, estimated_tokens)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			 RETURNING id`,
		record.UserID, record.APIKeyID, record.ChannelID, nullableInt64(record.WalletHoldID), record.Model,
		record.Method, record.Path, record.StatusCode,
		record.InputTokens, record.OutputTokens, record.Cost, record.ReservedAmount,
		record.LatencyMs, record.Error, record.IP,
		record.BillingStatus, record.BillingNote, record.EstimatedTokens,
	).Scan(&requestLogID)
	if insertErr != nil {
		return fmt.Errorf("failed to write log: %w", insertErr)
	}

	if holdState != nil && holdState.NeedsFinalizing {
		if err := w.finalizeHoldCapture(ctx, tx, record, requestLogID, *holdState); err != nil {
			return err
		}
	}

	if hasBalanceSnapshot {
		walletTx, ok := BuildWalletTransaction(record, requestLogID, balanceBefore, balanceAfter)
		if ok {
			if err := w.insertWalletTransaction(ctx, tx, walletTx); err != nil {
				return err
			}
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("failed to commit billing tx: %w", commitErr)
	}

	if billingStatus == BillingStatusChargeFailed {
		return fmt.Errorf("billing charge failed: %s", billingNote)
	}
	if billingStatus == BillingStatusUnpriced {
		return fmt.Errorf("billing skipped because model is unpriced: %s", billingNote)
	}

	return nil
}

type holdCaptureTx interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (w *DBLogWriter) prepareHoldCapture(ctx context.Context, tx holdCaptureTx, record *UsageRecord) (*holdCaptureState, string, string) {
	var hold model.WalletHold
	if err := tx.QueryRow(ctx,
		`SELECT id, user_id, api_key_id, model, amount, status
		 FROM wallet_holds WHERE id=$1 FOR UPDATE`,
		record.WalletHoldID,
	).Scan(&hold.ID, &hold.UserID, &hold.APIKeyID, &hold.Model, &hold.Amount, &hold.Status); err != nil {
		return nil, BillingStatusChargeFailed, err.Error()
	}
	if hold.Status != model.WalletHoldStatusHeld {
		return nil, BillingStatusChargeFailed, "预授权 hold 状态不是 held"
	}

	var balance int64
	var reservedBalance int64
	if err := tx.QueryRow(ctx, `SELECT balance, reserved_balance FROM users WHERE id=$1 FOR UPDATE`, record.UserID).Scan(&balance, &reservedBalance); err != nil {
		return nil, BillingStatusChargeFailed, err.Error()
	}

	plan := model.PlanHoldCapture(hold.Amount, record.Cost, balance)
	state := &holdCaptureState{
		Hold:            hold,
		Plan:            plan,
		BalanceBefore:   plan.BalanceBefore,
		BalanceAfter:    plan.BalanceAfter,
		ReservedBefore:  reservedBalance,
		HasSnapshot:     true,
		NeedsFinalizing: true,
	}
	if !plan.Allowed {
		return state, BillingStatusChargeFailed, plan.Note
	}
	if record.Cost <= 0 {
		return state, BillingStatusZeroCost, plan.Note
	}
	return state, BillingStatusCharged, "扣费成功，已 capture 预授权 hold"
}

func (w *DBLogWriter) finalizeHoldCapture(ctx context.Context, tx holdCaptureTx, record *UsageRecord, requestLogID int64, state holdCaptureState) error {
	status := model.WalletHoldStatusCaptured
	eventType := model.BillingEventHoldCaptured
	if !state.Plan.Allowed {
		status = model.WalletHoldStatusFailed
		eventType = model.BillingEventCaptureFailed
	} else if record.Cost <= 0 {
		status = model.WalletHoldStatusReleased
		eventType = model.BillingEventHoldReleased
	}

	if state.Plan.Allowed && record.Cost > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE users
			 SET balance = balance - $1,
			     used_amount = used_amount + $1,
			     request_count = request_count + 1,
			     reserved_balance = GREATEST(reserved_balance - $2, 0),
			     updated_at = NOW()
			 WHERE id=$3`,
			record.Cost, state.Hold.Amount, record.UserID,
		); err != nil {
			return fmt.Errorf("failed to capture wallet hold: %w", err)
		}
	} else {
		if _, err := tx.Exec(ctx,
			`UPDATE users
			 SET reserved_balance = GREATEST(reserved_balance - $1, 0),
			     updated_at = NOW()
			 WHERE id=$2`,
			state.Hold.Amount, record.UserID,
		); err != nil {
			return fmt.Errorf("failed to release wallet hold: %w", err)
		}
	}

	if _, err := tx.Exec(ctx,
		`UPDATE wallet_holds
		 SET status=$1, captured_amount=$2, released_amount=$3, request_log_id=$4, updated_at=NOW()
		 WHERE id=$5`,
		status, state.Plan.CaptureAmount, state.Plan.ReleaseAmount, requestLogID, state.Hold.ID,
	); err != nil {
		return fmt.Errorf("failed to update wallet hold: %w", err)
	}

	holdID := state.Hold.ID
	requestID := requestLogID
	_, err := insertRelayBillingEvent(ctx, tx, relayBillingEvent{
		UserID:       record.UserID,
		APIKeyID:     record.APIKeyID,
		RequestLogID: &requestID,
		WalletHoldID: &holdID,
		EventType:    eventType,
		Model:        record.Model,
		Amount:       state.Hold.Amount,
		Balance:      state.Plan.BalanceAfter,
		Reserved:     maxInt64(state.ReservedBefore-state.Hold.Amount, 0),
		Available:    state.Plan.BalanceAfter - maxInt64(state.ReservedBefore-state.Hold.Amount, 0),
		Status:       status,
		Note:         record.BillingNote,
	})
	if err != nil {
		return err
	}
	return nil
}

type walletTransactionWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (w *DBLogWriter) insertWalletTransaction(ctx context.Context, tx walletTransactionWriter, walletTx WalletTransactionRecord) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO wallet_transactions (user_id, request_log_id, type, amount, balance_before, balance_after, reference_type, reference_id, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		walletTx.UserID, walletTx.RequestLogID, walletTx.Type, walletTx.Amount,
		walletTx.BalanceBefore, walletTx.BalanceAfter,
		walletTx.ReferenceType, walletTx.ReferenceID, walletTx.Note,
	)
	if err != nil {
		return fmt.Errorf("failed to write wallet transaction: %w", err)
	}
	return nil
}

type relayBillingEvent struct {
	UserID       int64
	APIKeyID     int64
	RequestLogID *int64
	WalletHoldID *int64
	EventType    string
	Model        string
	Amount       int64
	Balance      int64
	Reserved     int64
	Available    int64
	Status       string
	Note         string
}

func insertRelayBillingEvent(ctx context.Context, tx holdCaptureTx, event relayBillingEvent) (int64, error) {
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

func nullableInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
