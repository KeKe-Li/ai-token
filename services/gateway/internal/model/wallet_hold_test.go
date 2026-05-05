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
