package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         int       `json:"role"`
	Status       int       `json:"status"`
	Balance      int64     `json:"balance"`
	Reserved     int64     `json:"reserved_balance"`
	UsedAmount   int64     `json:"used_amount"`
	RequestCount int64     `json:"request_count"`
	GroupName    string    `json:"group_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user User
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, email, role, status, balance, reserved_balance, used_amount, request_count, group_name, created_at, updated_at`,
		username, email, string(hash),
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status,
		&user.Balance, &user.Reserved, &user.UsedAmount, &user.RequestCount, &user.GroupName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, reserved_balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.Balance, &user.Reserved, &user.UsedAmount, &user.RequestCount, &user.GroupName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, reserved_balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.Balance, &user.Reserved, &user.UsedAmount, &user.RequestCount, &user.GroupName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (s *UserStore) DeductBalance(ctx context.Context, userID int64, amount int64) error {
	result, err := s.db.Exec(ctx,
		`UPDATE users SET balance = balance - $1, used_amount = used_amount + $1, request_count = request_count + 1, updated_at = NOW()
		 WHERE id = $2 AND balance >= $1`,
		amount, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient balance")
	}
	return nil
}

func (s *UserStore) CheckBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	err := s.db.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to check balance: %w", err)
	}
	return balance, nil
}

type WalletHoldResult struct {
	HoldID         int64
	Decision       ReserveBalanceDecision
	BillingEventID int64
}

func (s *UserStore) CreateWalletHold(ctx context.Context, userID, apiKeyID int64, model string, amount int64) (*WalletHoldResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin wallet hold tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var balance int64
	var reservedBalance int64
	if err := tx.QueryRow(ctx, `SELECT balance, reserved_balance FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&balance, &reservedBalance); err != nil {
		return nil, fmt.Errorf("failed to lock user balance for hold: %w", err)
	}

	decision := CheckReserveBalance(balance, reservedBalance, amount)
	result := &WalletHoldResult{Decision: decision}
	if !decision.Allowed {
		eventID, err := insertBillingEvent(ctx, tx, BillingEvent{
			UserID:    userID,
			APIKeyID:  apiKeyID,
			EventType: BillingEventPreAuthFailed,
			Model:     model,
			Amount:    amount,
			Balance:   balance,
			Reserved:  reservedBalance,
			Available: decision.Available,
			Status:    WalletHoldStatusFailed,
			Note:      decision.Note,
		})
		if err != nil {
			return nil, err
		}
		result.BillingEventID = eventID
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit failed wallet hold tx: %w", err)
		}
		return result, nil
	}

	if amount <= 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit zero wallet hold tx: %w", err)
		}
		return result, nil
	}

	if _, err := tx.Exec(ctx,
		`UPDATE users SET reserved_balance = reserved_balance + $1, updated_at = NOW() WHERE id=$2`,
		amount, userID,
	); err != nil {
		return nil, fmt.Errorf("failed to update reserved balance: %w", err)
	}

	var holdID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO wallet_holds (user_id, api_key_id, model, amount, status, reason, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW() + INTERVAL '30 minutes')
		 RETURNING id`,
		userID, apiKeyID, model, amount, WalletHoldStatusHeld, WalletHoldReasonAPIRequest,
	).Scan(&holdID); err != nil {
		return nil, fmt.Errorf("failed to insert wallet hold: %w", err)
	}

	holdIDPtr := holdID
	eventID, err := insertBillingEvent(ctx, tx, BillingEvent{
		UserID:       userID,
		APIKeyID:     apiKeyID,
		WalletHoldID: &holdIDPtr,
		EventType:    BillingEventPreAuthHeld,
		Model:        model,
		Amount:       amount,
		Balance:      balance,
		Reserved:     reservedBalance + amount,
		Available:    decision.Available - amount,
		Status:       WalletHoldStatusHeld,
		Note:         decision.Note,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit wallet hold tx: %w", err)
	}
	result.HoldID = holdID
	result.BillingEventID = eventID
	return result, nil
}

func (s *UserStore) ReleaseWalletHold(ctx context.Context, holdID int64, note string) error {
	if holdID <= 0 {
		return nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin wallet hold release tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var hold WalletHold
	if err := tx.QueryRow(ctx,
		`SELECT id, user_id, api_key_id, model, amount, status FROM wallet_holds WHERE id=$1 FOR UPDATE`,
		holdID,
	).Scan(&hold.ID, &hold.UserID, &hold.APIKeyID, &hold.Model, &hold.Amount, &hold.Status); err != nil {
		return fmt.Errorf("failed to lock wallet hold: %w", err)
	}
	if hold.Status != WalletHoldStatusHeld {
		return tx.Commit(ctx)
	}

	var balance int64
	var reservedBalance int64
	if err := tx.QueryRow(ctx, `SELECT balance, reserved_balance FROM users WHERE id=$1 FOR UPDATE`, hold.UserID).Scan(&balance, &reservedBalance); err != nil {
		return fmt.Errorf("failed to lock user for hold release: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE users SET reserved_balance = GREATEST(reserved_balance - $1, 0), updated_at = NOW() WHERE id=$2`,
		hold.Amount, hold.UserID,
	); err != nil {
		return fmt.Errorf("failed to release reserved balance: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE wallet_holds SET status=$1, released_amount=$2, updated_at=NOW() WHERE id=$3`,
		WalletHoldStatusReleased, hold.Amount, hold.ID,
	); err != nil {
		return fmt.Errorf("failed to update wallet hold release status: %w", err)
	}

	holdIDPtr := hold.ID
	if _, err := insertBillingEvent(ctx, tx, BillingEvent{
		UserID:       hold.UserID,
		APIKeyID:     hold.APIKeyID,
		WalletHoldID: &holdIDPtr,
		EventType:    BillingEventHoldReleased,
		Model:        hold.Model,
		Amount:       hold.Amount,
		Balance:      balance,
		Reserved:     maxInt64(reservedBalance-hold.Amount, 0),
		Available:    balance - maxInt64(reservedBalance-hold.Amount, 0),
		Status:       WalletHoldStatusReleased,
		Note:         note,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit wallet hold release tx: %w", err)
	}
	return nil
}

func (s *UserStore) CreateBillingEvent(ctx context.Context, event BillingEvent) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin billing event tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := insertBillingEvent(ctx, tx, event); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit billing event tx: %w", err)
	}
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (s *UserStore) ListAll(ctx context.Context) ([]User, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, reserved_balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.Status,
			&u.Balance, &u.Reserved, &u.UsedAmount, &u.RequestCount, &u.GroupName,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserStore) AdminUpdate(ctx context.Context, id int64, status *int, balance *int64, role *int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin admin user update tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var balanceBefore int64
	var walletTx *WalletTransaction
	if balance != nil {
		if err := tx.QueryRow(ctx, `SELECT balance FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&balanceBefore); err != nil {
			return fmt.Errorf("failed to lock user balance: %w", err)
		}
		if balanceBefore != *balance {
			adjustment := BuildAdminBalanceAdjustmentTransaction(id, balanceBefore, *balance, "管理员调整用户余额")
			walletTx = &adjustment
		}
	}

	if status != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2`, *status, id); err != nil {
			return err
		}
	}
	if balance != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET balance=$1, updated_at=NOW() WHERE id=$2`, *balance, id); err != nil {
			return err
		}
	}
	if role != nil {
		if _, err := tx.Exec(ctx, `UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, *role, id); err != nil {
			return err
		}
	}

	if walletTx != nil {
		if err := insertWalletTransaction(ctx, tx, *walletTx); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit admin user update tx: %w", err)
	}
	return nil
}

func insertWalletTransaction(ctx context.Context, tx pgx.Tx, walletTx WalletTransaction) error {
	var requestLogID any
	if walletTx.RequestLogID != nil {
		requestLogID = *walletTx.RequestLogID
	}

	var balanceBefore any
	if walletTx.BalanceBefore != nil {
		balanceBefore = *walletTx.BalanceBefore
	}

	var balanceAfter any
	if walletTx.BalanceAfter != nil {
		balanceAfter = *walletTx.BalanceAfter
	}

	_, err := tx.Exec(ctx,
		`INSERT INTO wallet_transactions (user_id, request_log_id, type, amount, balance_before, balance_after, reference_type, reference_id, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		walletTx.UserID, requestLogID, walletTx.Type, walletTx.Amount,
		balanceBefore, balanceAfter, walletTx.ReferenceType, walletTx.ReferenceID, walletTx.Note,
	)
	if err != nil {
		return fmt.Errorf("failed to insert wallet transaction: %w", err)
	}
	return nil
}
