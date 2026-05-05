package relay

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBPricingProvider struct {
	db *pgxpool.Pool
}

func NewDBPricingProvider(db *pgxpool.Pool) *DBPricingProvider {
	return &DBPricingProvider{db: db}
}

func (p *DBPricingProvider) GetPricing(ctx context.Context, model string) (*ModelPricing, error) {
	if p == nil || p.db == nil {
		return nil, fmt.Errorf("pricing provider is not configured")
	}

	var inputPrice int64
	var outputPrice int64
	var priceUnit string
	err := p.db.QueryRow(ctx,
		`SELECT input_price, output_price, price_unit
		 FROM models
		 WHERE model_id = $1 AND status = 1`,
		model,
	).Scan(&inputPrice, &outputPrice, &priceUnit)
	if err != nil {
		return nil, fmt.Errorf("model pricing not found for %s: %w", model, err)
	}

	unit, err := ParsePriceUnit(priceUnit)
	if err != nil {
		return nil, err
	}

	return &ModelPricing{
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		PriceUnit:   unit,
	}, nil
}
