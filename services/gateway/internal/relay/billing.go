package relay

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	BillingStatusPriced        = "priced"
	BillingStatusUnpriced      = "unpriced"
	BillingStatusZeroCost      = "zero_cost"
	BillingStatusCharged       = "charged"
	BillingStatusChargeFailed  = "charge_failed"
	BillingStatusNoUser        = "no_user"
	BillingStatusPreAuthOK     = "preauth_ok"
	BillingStatusPreAuthFailed = "preauth_failed"
)

const (
	WalletTransactionAPICharge       = "api_charge"
	WalletTransactionAPIChargeFailed = "api_charge_failed"
	WalletReferenceAPIRequest        = "api_request"
)

type ModelPricing struct {
	InputPrice  int64
	OutputPrice int64
	PriceUnit   int
}

type WalletTransactionRecord struct {
	UserID        int64
	RequestLogID  int64
	Type          string
	Amount        int64
	BalanceBefore int64
	BalanceAfter  int64
	ReferenceType string
	ReferenceID   string
	Note          string
}

type PricingProvider interface {
	GetPricing(ctx context.Context, model string) (*ModelPricing, error)
}

func CalculateCost(pricing *ModelPricing, inputTokens, outputTokens int) int64 {
	if pricing == nil || pricing.PriceUnit == 0 {
		return 0
	}

	inputCost := int64(inputTokens) * pricing.InputPrice / int64(pricing.PriceUnit)
	outputCost := int64(outputTokens) * pricing.OutputPrice / int64(pricing.PriceUnit)

	return inputCost + outputCost
}

type PreflightBalanceDecision struct {
	Allowed        bool
	Status         string
	RequiredAmount int64
	Balance        int64
	Note           string
}

func EstimatePreAuthorizationCost(pricing *ModelPricing, maxTokens *int, defaultMaxTokens int) int64 {
	if pricing == nil || pricing.PriceUnit <= 0 {
		return 0
	}
	tokens := defaultMaxTokens
	if maxTokens != nil && *maxTokens > 0 {
		tokens = *maxTokens
	}
	if tokens <= 0 {
		return 0
	}
	return int64(tokens) * pricing.OutputPrice / int64(pricing.PriceUnit)
}

func CheckPreflightBalance(balance int64, requiredAmount int64) PreflightBalanceDecision {
	if requiredAmount <= 0 {
		return PreflightBalanceDecision{Allowed: true, Status: BillingStatusPreAuthOK, RequiredAmount: requiredAmount, Balance: balance, Note: "无需预授权"}
	}
	if balance < requiredAmount {
		return PreflightBalanceDecision{Allowed: false, Status: BillingStatusPreAuthFailed, RequiredAmount: requiredAmount, Balance: balance, Note: "余额低于预授权金额"}
	}
	return PreflightBalanceDecision{Allowed: true, Status: BillingStatusPreAuthOK, RequiredAmount: requiredAmount, Balance: balance, Note: "预授权检查通过"}
}

func WalletTransactionTypeForBillingStatus(status string) (string, bool) {
	switch status {
	case BillingStatusCharged:
		return WalletTransactionAPICharge, true
	case BillingStatusChargeFailed:
		return WalletTransactionAPIChargeFailed, true
	default:
		return "", false
	}
}

func BuildWalletTransaction(record *UsageRecord, requestLogID, balanceBefore, balanceAfter int64) (WalletTransactionRecord, bool) {
	if record == nil || record.UserID <= 0 || requestLogID <= 0 {
		return WalletTransactionRecord{}, false
	}

	txType, ok := WalletTransactionTypeForBillingStatus(record.BillingStatus)
	if !ok {
		return WalletTransactionRecord{}, false
	}

	amount := int64(0)
	if record.BillingStatus == BillingStatusCharged {
		amount = -record.Cost
	}

	return WalletTransactionRecord{
		UserID:        record.UserID,
		RequestLogID:  requestLogID,
		Type:          txType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		ReferenceType: WalletReferenceAPIRequest,
		ReferenceID:   fmt.Sprintf("%d", requestLogID),
		Note:          record.BillingNote,
	}, true
}

func ResolvePricing(ctx context.Context, provider PricingProvider, model string) (*ModelPricing, string, string) {
	if provider == nil {
		return GetDefaultPricing(model), BillingStatusPriced, "使用环境默认价格"
	}

	pricing, err := provider.GetPricing(ctx, model)
	if err != nil || pricing == nil {
		if err != nil {
			return nil, BillingStatusUnpriced, err.Error()
		}
		return nil, BillingStatusUnpriced, "模型未配置价格"
	}

	return pricing, BillingStatusPriced, "使用模型表价格"
}

func ParsePriceUnit(unit string) (int, error) {
	normalized := strings.TrimSpace(strings.ToUpper(unit))
	if normalized == "" {
		return 1000000, nil
	}
	switch normalized {
	case "1K", "K":
		return 1000, nil
	case "1M", "M":
		return 1000000, nil
	}
	value, err := strconv.Atoi(normalized)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid price unit %q", unit)
	}
	return value, nil
}

func ClassifyChargeResult(cost int64, userID int64, rowsAffected int64, execErr error) (string, string, error) {
	if cost <= 0 {
		return BillingStatusZeroCost, "本次请求成本为 0", nil
	}
	if userID <= 0 {
		return BillingStatusNoUser, "缺少用户 ID，未执行扣费", nil
	}
	if execErr != nil {
		return BillingStatusChargeFailed, execErr.Error(), execErr
	}
	if rowsAffected == 0 {
		err := fmt.Errorf("余额不足或用户不存在，未能扣减 %d", cost)
		return BillingStatusChargeFailed, err.Error(), err
	}
	return BillingStatusCharged, "扣费成功", nil
}

func GetDefaultPricing(model string) *ModelPricing {
	prices := map[string]*ModelPricing{
		"gpt-4o":            {InputPrice: 2500, OutputPrice: 10000, PriceUnit: 1000000},
		"gpt-4o-mini":       {InputPrice: 150, OutputPrice: 600, PriceUnit: 1000000},
		"claude-sonnet-4-6": {InputPrice: 3000, OutputPrice: 15000, PriceUnit: 1000000},
		"claude-haiku-4-5":  {InputPrice: 800, OutputPrice: 4000, PriceUnit: 1000000},
		"gemini-2.5-pro":    {InputPrice: 1250, OutputPrice: 10000, PriceUnit: 1000000},
		"gemini-2.5-flash":  {InputPrice: 150, OutputPrice: 600, PriceUnit: 1000000},
		"deepseek-chat":     {InputPrice: 270, OutputPrice: 1100, PriceUnit: 1000000},
		"deepseek-reasoner": {InputPrice: 550, OutputPrice: 2190, PriceUnit: 1000000},
	}

	if p, ok := prices[model]; ok {
		return p
	}
	return &ModelPricing{InputPrice: 1000, OutputPrice: 3000, PriceUnit: 1000000}
}
