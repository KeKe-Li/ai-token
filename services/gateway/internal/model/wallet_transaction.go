package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	WalletTransactionAdminAdjustment = "admin_adjustment"
	WalletReferenceAdminUserUpdate   = "admin_user_update"
)

// WalletTransaction 表示钱包余额变动流水。
// amount 采用与 users.balance 相同单位：0.001 元；API 消费为负数。
type WalletTransaction struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	RequestLogID  *int64    `json:"request_log_id,omitempty"`
	Type          string    `json:"type"`
	Amount        int64     `json:"amount"`
	BalanceBefore *int64    `json:"balance_before,omitempty"`
	BalanceAfter  *int64    `json:"balance_after,omitempty"`
	ReferenceType string    `json:"reference_type,omitempty"`
	ReferenceID   string    `json:"reference_id,omitempty"`
	Note          string    `json:"note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func BuildAdminBalanceAdjustmentTransaction(userID, balanceBefore, balanceAfter int64, note string) WalletTransaction {
	before := balanceBefore
	after := balanceAfter
	return WalletTransaction{
		UserID:        userID,
		Type:          WalletTransactionAdminAdjustment,
		Amount:        balanceAfter - balanceBefore,
		BalanceBefore: &before,
		BalanceAfter:  &after,
		ReferenceType: WalletReferenceAdminUserUpdate,
		ReferenceID:   strconv.FormatInt(userID, 10),
		Note:          note,
	}
}

type WalletTransactionStore struct {
	db *pgxpool.Pool
}

func NewWalletTransactionStore(db *pgxpool.Pool) *WalletTransactionStore {
	return &WalletTransactionStore{db: db}
}

func (s *WalletTransactionStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]WalletTransaction, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, request_log_id, type, amount, balance_before, balance_after, reference_type, reference_id, note, created_at
		 FROM wallet_transactions
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallet transactions: %w", err)
	}
	defer rows.Close()

	return scanWalletTransactions(rows)
}

func (s *WalletTransactionStore) ListAll(ctx context.Context, userID *int64, limit, offset int) ([]WalletTransaction, error) {
	if userID != nil {
		return s.ListByUser(ctx, *userID, limit, offset)
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, request_log_id, type, amount, balance_before, balance_after, reference_type, reference_id, note, created_at
		 FROM wallet_transactions
		 ORDER BY created_at DESC, id DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallet transactions: %w", err)
	}
	defer rows.Close()

	return scanWalletTransactions(rows)
}

type walletRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanWalletTransactions(rows walletRows) ([]WalletTransaction, error) {
	var transactions []WalletTransaction
	for rows.Next() {
		var tx WalletTransaction
		var requestLogID pgtype.Int8
		var balanceBefore pgtype.Int8
		var balanceAfter pgtype.Int8
		var referenceType pgtype.Text
		var referenceID pgtype.Text
		var note pgtype.Text

		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&requestLogID,
			&tx.Type,
			&tx.Amount,
			&balanceBefore,
			&balanceAfter,
			&referenceType,
			&referenceID,
			&note,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan wallet transaction: %w", err)
		}

		if requestLogID.Valid {
			value := requestLogID.Int64
			tx.RequestLogID = &value
		}
		if balanceBefore.Valid {
			value := balanceBefore.Int64
			tx.BalanceBefore = &value
		}
		if balanceAfter.Valid {
			value := balanceAfter.Int64
			tx.BalanceAfter = &value
		}
		if referenceType.Valid {
			tx.ReferenceType = referenceType.String
		}
		if referenceID.Valid {
			tx.ReferenceID = referenceID.String
		}
		if note.Valid {
			tx.Note = note.String
		}

		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate wallet transactions: %w", err)
	}
	return transactions, nil
}
