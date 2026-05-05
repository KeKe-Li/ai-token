package relay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakePricingProvider struct {
	pricing *ModelPricing
	err     error
}

type fakeWalletWriter struct {
	sql  string
	args []any
}

func (f *fakeWalletWriter) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	f.sql = sql
	f.args = arguments
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (f fakePricingProvider) GetPricing(ctx context.Context, model string) (*ModelPricing, error) {
	return f.pricing, f.err
}

func TestResolvePricingUsesProviderPrice(t *testing.T) {
	pricing, status, note := ResolvePricing(context.Background(), fakePricingProvider{
		pricing: &ModelPricing{InputPrice: 7, OutputPrice: 11, PriceUnit: 1000},
	}, "custom-model")

	if status != BillingStatusPriced {
		t.Fatalf("期望使用模型表价格时状态为 %s，实际=%s note=%s", BillingStatusPriced, status, note)
	}
	if got := CalculateCost(pricing, 1000, 2000); got != 29 {
		t.Fatalf("期望使用 provider 返回的价格计费，实际成本=%d", got)
	}
}

func TestResolvePricingMarksMissingProviderPriceAsUnpriced(t *testing.T) {
	pricing, status, _ := ResolvePricing(context.Background(), fakePricingProvider{err: errors.New("not found")}, "missing-model")
	if pricing != nil {
		t.Fatalf("模型表缺失价格时不应回退硬编码默认价格，实际=%+v", pricing)
	}
	if status != BillingStatusUnpriced {
		t.Fatalf("模型表缺失价格应标记为 %s，实际=%s", BillingStatusUnpriced, status)
	}
}

func TestParsePriceUnit(t *testing.T) {
	cases := map[string]int{"1K": 1000, "1M": 1000000, "1000": 1000, "": 1000000}
	for input, want := range cases {
		got, err := ParsePriceUnit(input)
		if err != nil {
			t.Fatalf("ParsePriceUnit(%q) 不应失败: %v", input, err)
		}
		if got != want {
			t.Fatalf("ParsePriceUnit(%q)=%d，期望=%d", input, got, want)
		}
	}
}

func TestClassifyChargeResultDetectsInsufficientBalance(t *testing.T) {
	status, note, err := ClassifyChargeResult(42, 7, 0, nil)
	if err == nil {
		t.Fatal("扣费影响 0 行时应返回错误，避免余额扣减失败被静默吞掉")
	}
	if status != BillingStatusChargeFailed {
		t.Fatalf("期望状态=%s，实际=%s note=%s", BillingStatusChargeFailed, status, note)
	}
}

func TestEstimatePreAuthorizationCostUsesRequestedMaxTokens(t *testing.T) {
	pricing := &ModelPricing{InputPrice: 100, OutputPrice: 400, PriceUnit: 1000}
	maxTokens := 2500
	if got := EstimatePreAuthorizationCost(pricing, &maxTokens, 1024); got != 1000 {
		t.Fatalf("预授权应按 max_tokens 的最大输出成本估算，实际=%d", got)
	}
}

func TestEstimatePreAuthorizationCostFallsBackToDefaultMaxTokens(t *testing.T) {
	pricing := &ModelPricing{InputPrice: 100, OutputPrice: 400, PriceUnit: 1000}
	if got := EstimatePreAuthorizationCost(pricing, nil, 1000); got != 400 {
		t.Fatalf("未传 max_tokens 时应按默认输出上限估算，实际=%d", got)
	}
}

func TestPreflightBalanceRejectsInsufficientReservedAmount(t *testing.T) {
	decision := CheckPreflightBalance(399, 400)
	if decision.Allowed {
		t.Fatal("余额小于预授权金额时应拒绝请求")
	}
	if decision.Status != BillingStatusPreAuthFailed {
		t.Fatalf("期望预授权失败状态=%s，实际=%s", BillingStatusPreAuthFailed, decision.Status)
	}
}

func TestWalletTransactionTypeForBillingStatus(t *testing.T) {
	txType, ok := WalletTransactionTypeForBillingStatus(BillingStatusCharged)
	if !ok || txType != WalletTransactionAPICharge {
		t.Fatalf("charged 应生成 API 消费流水，type=%s ok=%v", txType, ok)
	}
	txType, ok = WalletTransactionTypeForBillingStatus(BillingStatusChargeFailed)
	if !ok || txType != WalletTransactionAPIChargeFailed {
		t.Fatalf("charge_failed 应生成扣费失败流水，type=%s ok=%v", txType, ok)
	}
}

func TestBuildWalletTransactionForChargedRequest(t *testing.T) {
	record := &UsageRecord{
		UserID:        7,
		Cost:          250,
		BillingStatus: BillingStatusCharged,
		BillingNote:   "扣费成功",
	}

	tx, ok := BuildWalletTransaction(record, 99, 1000, 750)
	if !ok {
		t.Fatal("扣费成功应生成钱包流水")
	}
	if tx.Type != WalletTransactionAPICharge {
		t.Fatalf("流水类型错误，实际=%s", tx.Type)
	}
	if tx.Amount != -250 {
		t.Fatalf("API 消费流水金额应为负数，实际=%d", tx.Amount)
	}
	if tx.RequestLogID != 99 || tx.BalanceBefore != 1000 || tx.BalanceAfter != 750 {
		t.Fatalf("流水关联或余额快照错误: %+v", tx)
	}
}

func TestBuildWalletTransactionForChargeFailedRequest(t *testing.T) {
	record := &UsageRecord{
		UserID:        7,
		Cost:          250,
		BillingStatus: BillingStatusChargeFailed,
		BillingNote:   "余额不足",
	}

	tx, ok := BuildWalletTransaction(record, 99, 100, 100)
	if !ok {
		t.Fatal("扣费失败应生成 0 金额钱包流水，保留失败审计")
	}
	if tx.Type != WalletTransactionAPIChargeFailed {
		t.Fatalf("流水类型错误，实际=%s", tx.Type)
	}
	if tx.Amount != 0 {
		t.Fatalf("扣费失败流水金额应为 0，实际=%d", tx.Amount)
	}
	if tx.BalanceBefore != tx.BalanceAfter {
		t.Fatalf("扣费失败不应改变余额快照: %+v", tx)
	}
}

func TestInsertWalletTransactionWritesAuditColumns(t *testing.T) {
	writer := &fakeWalletWriter{}
	walletTx := WalletTransactionRecord{
		UserID:        7,
		RequestLogID:  99,
		Type:          WalletTransactionAPICharge,
		Amount:        -250,
		BalanceBefore: 1000,
		BalanceAfter:  750,
		ReferenceType: WalletReferenceAPIRequest,
		ReferenceID:   "99",
		Note:          "扣费成功",
	}

	if err := (&DBLogWriter{}).insertWalletTransaction(context.Background(), writer, walletTx); err != nil {
		t.Fatalf("写入钱包流水不应失败: %v", err)
	}
	if !strings.Contains(writer.sql, "wallet_transactions") || !strings.Contains(writer.sql, "balance_before") || !strings.Contains(writer.sql, "balance_after") {
		t.Fatalf("钱包流水 SQL 必须包含流水表与余额快照字段，sql=%s", writer.sql)
	}
	if len(writer.args) != 9 {
		t.Fatalf("钱包流水参数数量错误，实际=%d", len(writer.args))
	}
	if writer.args[3] != int64(-250) || writer.args[4] != int64(1000) || writer.args[5] != int64(750) {
		t.Fatalf("钱包流水金额或余额快照参数错误: %#v", writer.args)
	}
}
