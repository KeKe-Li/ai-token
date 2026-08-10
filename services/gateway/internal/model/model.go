package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrModelNotFound 表示按 model_id 查询时没有命中激活的模型。
// 上层据此返回 404，而非把数据库故障也当成「不存在」。
var ErrModelNotFound = errors.New("model not found")

// modelColumns 是 models 表全部字段的单一事实来源，供各查询拼接 SQL。
const modelColumns = "id, model_id, display_name, provider, category, context_length, input_price, output_price, price_unit, capabilities, status, icon_url, description, created_at, updated_at"

// rowScanner 抽象 pgx.Row 与 pgx.Rows 的公共 Scan 行为，便于共享扫描逻辑。
type rowScanner interface {
	Scan(dest ...any) error
}

// scanModel 按 modelColumns 的顺序把一行结果扫描为 AIModel。
func scanModel(row rowScanner) (AIModel, error) {
	var m AIModel
	err := row.Scan(&m.ID, &m.ModelID, &m.DisplayName, &m.Provider, &m.Category,
		&m.ContextLength, &m.InputPrice, &m.OutputPrice, &m.PriceUnit,
		&m.Capabilities, &m.Status, &m.IconURL, &m.Description,
		&m.CreatedAt, &m.UpdatedAt)
	return m, err
}

type AIModel struct {
	ID            int64     `json:"id"`
	ModelID       string    `json:"model_id"`
	DisplayName   string    `json:"display_name"`
	Provider      string    `json:"provider"`
	Category      string    `json:"category"`
	ContextLength int       `json:"context_length"`
	InputPrice    int64     `json:"input_price"`
	OutputPrice   int64     `json:"output_price"`
	PriceUnit     string    `json:"price_unit"`
	Capabilities  []string  `json:"capabilities"`
	Status        int       `json:"status"`
	IconURL       *string   `json:"icon_url,omitempty"`
	Description   *string   `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ModelStore struct {
	db *pgxpool.Pool
}

func NewModelStore(db *pgxpool.Pool) *ModelStore {
	return &ModelStore{db: db}
}

func (s *ModelStore) List(ctx context.Context) ([]AIModel, error) {
	rows, err := s.db.Query(ctx,
		`SELECT `+modelColumns+` FROM models ORDER BY provider, model_id`)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []AIModel
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, m)
	}
	return models, nil
}

func (s *ModelStore) Create(ctx context.Context, m *AIModel) error {
	err := s.db.QueryRow(ctx,
		`INSERT INTO models (model_id, display_name, provider, category, context_length, input_price, output_price, price_unit, capabilities, status, icon_url, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, created_at, updated_at`,
		m.ModelID, m.DisplayName, m.Provider, m.Category, m.ContextLength,
		m.InputPrice, m.OutputPrice, m.PriceUnit, m.Capabilities, m.Status,
		m.IconURL, m.Description,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

func (s *ModelStore) Update(ctx context.Context, m *AIModel) error {
	_, err := s.db.Exec(ctx,
		`UPDATE models SET display_name=$1, provider=$2, category=$3, context_length=$4, input_price=$5, output_price=$6, price_unit=$7, capabilities=$8, status=$9, icon_url=$10, description=$11, updated_at=NOW()
		 WHERE id=$12`,
		m.DisplayName, m.Provider, m.Category, m.ContextLength,
		m.InputPrice, m.OutputPrice, m.PriceUnit, m.Capabilities, m.Status,
		m.IconURL, m.Description, m.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}
	return nil
}

func (s *ModelStore) ListActive(ctx context.Context) ([]AIModel, error) {
	rows, err := s.db.Query(ctx,
		`SELECT `+modelColumns+` FROM models WHERE status = 1 ORDER BY provider, model_id`)
	if err != nil {
		return nil, fmt.Errorf("failed to list active models: %w", err)
	}
	defer rows.Close()

	var models []AIModel
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, m)
	}
	return models, nil
}

func (s *ModelStore) GetByModelID(ctx context.Context, modelID string) (*AIModel, error) {
	m, err := scanModel(s.db.QueryRow(ctx,
		`SELECT `+modelColumns+` FROM models WHERE model_id = $1 AND status = 1`, modelID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &m, nil
}

func (s *ModelStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `DELETE FROM models WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("model not found")
	}
	return nil
}
