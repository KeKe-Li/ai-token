package relay

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBillingHoldCaptureReleaseIntegration(t *testing.T) {
	databaseURL := os.Getenv("AI_TOKEN_INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("设置 AI_TOKEN_INTEGRATION_DATABASE_URL 后运行真实 PostgreSQL 账务集成验证")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db := setupBillingIntegrationDB(t, ctx, databaseURL)
	userStore := model.NewUserStore(db)
	logStore := model.NewLogStore(db)
	eventStore := model.NewBillingEventStore(db)

	userID := createBillingIntegrationUser(t, ctx, db, 5000)
	const apiKeyID int64 = 123

	holdResult, err := userStore.CreateWalletHold(ctx, userID, apiKeyID, "gpt-integration", 1000)
	if err != nil {
		t.Fatalf("创建预授权 hold 失败: %v", err)
	}
	if !holdResult.Decision.Allowed || holdResult.HoldID <= 0 {
		t.Fatalf("预授权 hold 应创建成功: %+v", holdResult)
	}
	assertUserBalance(t, ctx, db, userID, 5000, 1000, 0, 0)

	writer := NewDBLogWriter(db, fakePricingProvider{
		pricing: &ModelPricing{InputPrice: 0, OutputPrice: 100, PriceUnit: 1000},
	})
	record := &UsageRecord{
		UserID:         userID,
		APIKeyID:       apiKeyID,
		WalletHoldID:   holdResult.HoldID,
		Model:          "gpt-integration",
		Method:         "POST",
		Path:           "/v1/chat/completions",
		StatusCode:     200,
		InputTokens:    100,
		OutputTokens:   2500,
		ReservedAmount: 1000,
		LatencyMs:      88,
		IP:             "127.0.0.1",
	}
	if err := writer.WriteLog(ctx, record); err != nil {
		t.Fatalf("写入 capture 账务日志失败: %v", err)
	}
	if record.Cost != 250 || record.BillingStatus != BillingStatusCharged {
		t.Fatalf("集成扣费结果不正确: cost=%d status=%s note=%s", record.Cost, record.BillingStatus, record.BillingNote)
	}
	assertUserBalance(t, ctx, db, userID, 4750, 0, 250, 1)

	var requestLogID int64
	var requestWalletHoldID int64
	if err := db.QueryRow(ctx,
		`SELECT id, wallet_hold_id FROM request_logs WHERE user_id=$1 AND model=$2`,
		userID, "gpt-integration",
	).Scan(&requestLogID, &requestWalletHoldID); err != nil {
		t.Fatalf("读取请求日志失败: %v", err)
	}
	if requestWalletHoldID != holdResult.HoldID {
		t.Fatalf("请求日志应关联 hold，got=%d want=%d", requestWalletHoldID, holdResult.HoldID)
	}

	drillDownLogs, err := logStore.ListAllFiltered(ctx, model.RequestLogListFilter{
		RequestLogID: &requestLogID,
		Limit:        10,
		Offset:       0,
	})
	if err != nil {
		t.Fatalf("request_log_id drill-down 查询失败: %v", err)
	}
	if len(drillDownLogs) != 1 || drillDownLogs[0].ID != requestLogID || drillDownLogs[0].WalletHoldID == nil || *drillDownLogs[0].WalletHoldID != holdResult.HoldID {
		t.Fatalf("request_log_id drill-down 应精确返回目标日志并保留 hold 关联: %+v", drillDownLogs)
	}

	assertCapturedHold(t, ctx, db, holdResult.HoldID, requestLogID, 250, 750)
	assertWalletTransaction(t, ctx, db, userID, requestLogID, -250, 5000, 4750)
	assertBillingEventTypes(t, ctx, eventStore, holdResult.HoldID, model.BillingEventPreAuthHeld, model.BillingEventHoldCaptured)

	manualHold, err := userStore.CreateWalletHold(ctx, userID, apiKeyID, "gpt-integration", 300)
	if err != nil {
		t.Fatalf("创建手动释放 hold 失败: %v", err)
	}
	if err := userStore.ReleaseWalletHold(ctx, manualHold.HoldID, "集成测试手动释放"); err != nil {
		t.Fatalf("手动释放 hold 失败: %v", err)
	}
	assertUserBalance(t, ctx, db, userID, 4750, 0, 250, 1)
	assertReleasedHold(t, ctx, db, manualHold.HoldID, 300)
	assertBillingEventTypes(t, ctx, eventStore, manualHold.HoldID, model.BillingEventPreAuthHeld, model.BillingEventHoldReleased)

	expiredHold, err := userStore.CreateWalletHold(ctx, userID, apiKeyID, "gpt-integration", 200)
	if err != nil {
		t.Fatalf("创建过期释放 hold 失败: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE wallet_holds SET expires_at = NOW() - INTERVAL '1 minute' WHERE id=$1`, expiredHold.HoldID); err != nil {
		t.Fatalf("设置 hold 过期失败: %v", err)
	}
	released, err := userStore.ReleaseExpiredWalletHolds(ctx, 100)
	if err != nil {
		t.Fatalf("释放过期 hold 失败: %v", err)
	}
	if released != 1 {
		t.Fatalf("应只释放 1 个过期 hold，实际=%d", released)
	}
	assertUserBalance(t, ctx, db, userID, 4750, 0, 250, 1)
	assertReleasedHold(t, ctx, db, expiredHold.HoldID, 200)
	assertBillingEventTypes(t, ctx, eventStore, expiredHold.HoldID, model.BillingEventPreAuthHeld, model.BillingEventHoldReleased)
}

func setupBillingIntegrationDB(t *testing.T, ctx context.Context, databaseURL string) *pgxpool.Pool {
	t.Helper()

	adminCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("解析集成测试数据库 URL 失败: %v", err)
	}
	adminDB, err := pgxpool.NewWithConfig(ctx, adminCfg)
	if err != nil {
		t.Fatalf("连接集成测试数据库失败: %v", err)
	}
	defer adminDB.Close()

	schema := fmt.Sprintf("ai_token_it_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := adminDB.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("创建集成测试 schema 失败: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("解析集成测试数据库 URL 失败: %v", err)
	}
	cfg.MaxConns = 2
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("连接集成测试 schema 失败: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cleanupDB, err := pgxpool.NewWithConfig(cleanupCtx, adminCfg)
		if err != nil {
			t.Logf("连接清理数据库失败: %v", err)
			return
		}
		defer cleanupDB.Close()
		if _, err := cleanupDB.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE"); err != nil {
			t.Logf("清理集成测试 schema 失败: %v", err)
		}
	})

	applyGatewayMigrations(t, ctx, db)
	return db
}

func applyGatewayMigrations(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil {
		t.Fatalf("读取迁移文件失败: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("未找到迁移文件，无法执行真实集成验证")
	}
	sort.Strings(matches)
	for _, path := range matches {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("读取迁移文件 %s 失败: %v", path, err)
		}
		if _, err := db.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("执行迁移文件 %s 失败: %v", path, err)
		}
	}
}

func createBillingIntegrationUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, balance int64) int64 {
	t.Helper()

	var userID int64
	email := fmt.Sprintf("billing-it-%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, balance)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		strings.TrimSuffix(email, "@example.com"), email, "integration-test-password-hash", balance,
	).Scan(&userID); err != nil {
		t.Fatalf("创建集成测试用户失败: %v", err)
	}
	return userID
}

func assertUserBalance(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID, balance, reserved, used, requests int64) {
	t.Helper()

	var gotBalance int64
	var gotReserved int64
	var gotUsed int64
	var gotRequests int64
	if err := db.QueryRow(ctx,
		`SELECT balance, reserved_balance, used_amount, request_count FROM users WHERE id=$1`,
		userID,
	).Scan(&gotBalance, &gotReserved, &gotUsed, &gotRequests); err != nil {
		t.Fatalf("读取用户余额失败: %v", err)
	}
	if gotBalance != balance || gotReserved != reserved || gotUsed != used || gotRequests != requests {
		t.Fatalf("用户余额快照不正确: balance=%d reserved=%d used=%d requests=%d，期望 balance=%d reserved=%d used=%d requests=%d",
			gotBalance, gotReserved, gotUsed, gotRequests, balance, reserved, used, requests)
	}
}

func assertCapturedHold(t *testing.T, ctx context.Context, db *pgxpool.Pool, holdID, requestLogID, captured, released int64) {
	t.Helper()

	var status string
	var gotRequestLogID int64
	var gotCaptured int64
	var gotReleased int64
	if err := db.QueryRow(ctx,
		`SELECT status, request_log_id, captured_amount, released_amount FROM wallet_holds WHERE id=$1`,
		holdID,
	).Scan(&status, &gotRequestLogID, &gotCaptured, &gotReleased); err != nil {
		t.Fatalf("读取 captured hold 失败: %v", err)
	}
	if status != model.WalletHoldStatusCaptured || gotRequestLogID != requestLogID || gotCaptured != captured || gotReleased != released {
		t.Fatalf("captured hold 状态不正确: status=%s request_log_id=%d captured=%d released=%d",
			status, gotRequestLogID, gotCaptured, gotReleased)
	}
}

func assertReleasedHold(t *testing.T, ctx context.Context, db *pgxpool.Pool, holdID, released int64) {
	t.Helper()

	var status string
	var gotReleased int64
	if err := db.QueryRow(ctx,
		`SELECT status, released_amount FROM wallet_holds WHERE id=$1`,
		holdID,
	).Scan(&status, &gotReleased); err != nil {
		t.Fatalf("读取 released hold 失败: %v", err)
	}
	if status != model.WalletHoldStatusReleased || gotReleased != released {
		t.Fatalf("released hold 状态不正确: status=%s released=%d", status, gotReleased)
	}
}

func assertWalletTransaction(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID, requestLogID, amount, before, after int64) {
	t.Helper()

	var txType string
	var gotAmount int64
	var gotBefore int64
	var gotAfter int64
	if err := db.QueryRow(ctx,
		`SELECT type, amount, balance_before, balance_after
		 FROM wallet_transactions
		 WHERE user_id=$1 AND request_log_id=$2`,
		userID, requestLogID,
	).Scan(&txType, &gotAmount, &gotBefore, &gotAfter); err != nil {
		t.Fatalf("读取钱包流水失败: %v", err)
	}
	if txType != WalletTransactionAPICharge || gotAmount != amount || gotBefore != before || gotAfter != after {
		t.Fatalf("钱包流水不正确: type=%s amount=%d before=%d after=%d", txType, gotAmount, gotBefore, gotAfter)
	}
}

func assertBillingEventTypes(t *testing.T, ctx context.Context, store *model.BillingEventStore, holdID int64, expected ...string) {
	t.Helper()

	events, err := store.ListByWalletHold(ctx, holdID, 20, 0)
	if err != nil {
		t.Fatalf("读取账务事件失败: %v", err)
	}
	seen := make(map[string]bool, len(events))
	for _, event := range events {
		seen[event.EventType] = true
	}
	for _, eventType := range expected {
		if !seen[eventType] {
			t.Fatalf("hold #%d 缺少账务事件 %s，events=%+v", holdID, eventType, events)
		}
	}
}
