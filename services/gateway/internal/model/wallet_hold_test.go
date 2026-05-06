package model

import "testing"

func TestCheckReserveBalanceUsesAvailableBalance(t *testing.T) {
	decision := CheckReserveBalance(1000, 700, 400)
	if decision.Allowed {
		t.Fatal("余额扣除已有预留后不足时，不应允许继续创建 hold")
	}
	if decision.Available != 300 {
		t.Fatalf("可用余额应为 balance-reserved_balance，实际=%d", decision.Available)
	}
}

func TestPlanHoldCaptureReleasesUnusedReservation(t *testing.T) {
	plan := PlanHoldCapture(1000, 250, 5000)
	if !plan.Allowed {
		t.Fatalf("余额足够时应允许 capture: %+v", plan)
	}
	if plan.CaptureAmount != 250 || plan.ReleaseAmount != 750 {
		t.Fatalf("应 capture 实际成本并 release 剩余 hold，实际=%+v", plan)
	}
	if plan.BalanceAfter != 4750 {
		t.Fatalf("扣费后余额错误，实际=%d", plan.BalanceAfter)
	}
}

func TestPlanHoldCaptureRejectsCostAboveBalance(t *testing.T) {
	plan := PlanHoldCapture(1000, 6000, 5000)
	if plan.Allowed {
		t.Fatalf("实际成本超过余额时不应允许 capture: %+v", plan)
	}
	if plan.ReleaseAmount != 1000 {
		t.Fatalf("capture 失败时应释放 hold，避免资金长期冻结，实际=%+v", plan)
	}
}

func TestBuildExpiredHoldReleaseEventCapturesAvailableBalance(t *testing.T) {
	event := BuildExpiredHoldReleaseEvent(WalletHold{ID: 11, UserID: 7, APIKeyID: 3, Model: "gpt-4o", Amount: 900}, 5000, 1200)
	if event.EventType != BillingEventHoldReleased {
		t.Fatalf("过期 hold 应生成释放事件，实际=%s", event.EventType)
	}
	if event.WalletHoldID == nil || *event.WalletHoldID != 11 {
		t.Fatalf("释放事件应关联 hold id，实际=%v", event.WalletHoldID)
	}
	if event.Reserved != 300 || event.Available != 4700 {
		t.Fatalf("释放事件应记录释放后的冻结与可用余额，实际 reserved=%d available=%d", event.Reserved, event.Available)
	}
}

func TestNormalizeWalletHoldStatusFilter(t *testing.T) {
	if status, ok := NormalizeWalletHoldStatusFilter("held"); !ok || status != WalletHoldStatusHeld {
		t.Fatalf("held 应是合法状态过滤，status=%s ok=%v", status, ok)
	}
	if status, ok := NormalizeWalletHoldStatusFilter(""); !ok || status != "" {
		t.Fatalf("空状态应表示不过滤，status=%s ok=%v", status, ok)
	}
	if _, ok := NormalizeWalletHoldStatusFilter("unknown"); ok {
		t.Fatal("未知 hold 状态不应被接受，避免拼接不可信过滤条件")
	}
}

func TestBuildManualHoldReleaseNote(t *testing.T) {
	if got := BuildManualHoldReleaseNote("用户反馈冻结异常"); got != "管理员手动释放异常 hold: 用户反馈冻结异常" {
		t.Fatalf("手动释放说明不正确: %s", got)
	}
	if got := BuildManualHoldReleaseNote("   "); got != "管理员手动释放异常 hold" {
		t.Fatalf("空原因应使用默认说明: %s", got)
	}
}
