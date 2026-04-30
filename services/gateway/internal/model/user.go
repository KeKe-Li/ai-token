package model

import (
	"context"
	"fmt"
	"time"

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
		 RETURNING id, username, email, role, status, balance, used_amount, request_count, group_name, created_at, updated_at`,
		username, email, string(hash),
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status,
		&user.Balance, &user.UsedAmount, &user.RequestCount, &user.GroupName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.Balance, &user.UsedAmount, &user.RequestCount, &user.GroupName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.Balance, &user.UsedAmount, &user.RequestCount, &user.GroupName,
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

func (u *User) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (s *UserStore) ListAll(ctx context.Context) ([]User, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, username, email, password_hash, role, status, balance, used_amount, request_count, group_name, created_at, updated_at
		 FROM users ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.Status,
			&u.Balance, &u.UsedAmount, &u.RequestCount, &u.GroupName,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserStore) AdminUpdate(ctx context.Context, id int64, status *int, balance *int64, role *int) error {
	if status != nil {
		if _, err := s.db.Exec(ctx, `UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2`, *status, id); err != nil {
			return err
		}
	}
	if balance != nil {
		if _, err := s.db.Exec(ctx, `UPDATE users SET balance=$1, updated_at=NOW() WHERE id=$2`, *balance, id); err != nil {
			return err
		}
	}
	if role != nil {
		if _, err := s.db.Exec(ctx, `UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, *role, id); err != nil {
			return err
		}
	}
	return nil
}
