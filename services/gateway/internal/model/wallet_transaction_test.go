package model

import "testing"

func TestBuildAdminBalanceAdjustmentTransactionCapturesDelta(t *testing.T) {
	tx := BuildAdminBalanceAdjustmentTransaction(7, 1000, 1600, "管理员手动充值")

	if tx.Type != WalletTransactionAdminAdjustment {
		t.Fatalf("管理员调账应生成 %s 流水，实际=%s", WalletTransactionAdminAdjustment, tx.Type)
	}
	if tx.Amount != 600 {
		t.Fatalf("调账流水金额应记录余额差值，实际=%d", tx.Amount)
	}
	if tx.BalanceBefore == nil || *tx.BalanceBefore != 1000 {
		t.Fatalf("调账流水应记录调整前余额，实际=%v", tx.BalanceBefore)
	}
	if tx.BalanceAfter == nil || *tx.BalanceAfter != 1600 {
		t.Fatalf("调账流水应记录调整后余额，实际=%v", tx.BalanceAfter)
	}
	if tx.ReferenceType != WalletReferenceAdminUserUpdate || tx.ReferenceID != "7" {
		t.Fatalf("调账流水应关联管理员用户更新事件，实际=%s/%s", tx.ReferenceType, tx.ReferenceID)
	}
}

func TestBuildAdminBalanceAdjustmentTransactionAllowsNegativeDelta(t *testing.T) {
	tx := BuildAdminBalanceAdjustmentTransaction(7, 1600, 1000, "管理员扣减余额")
	if tx.Amount != -600 {
		t.Fatalf("扣减余额时流水金额应为负数，实际=%d", tx.Amount)
	}
}
