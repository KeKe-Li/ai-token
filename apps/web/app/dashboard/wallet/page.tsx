"use client";

import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";

type WalletTransaction = {
  id: number;
  user_id: number;
  request_log_id?: number;
  type: string;
  amount: number;
  balance_before?: number;
  balance_after?: number;
  reference_type?: string;
  reference_id?: string;
  note?: string;
  created_at: string;
};

export default function WalletPage() {
  const [transactions, setTransactions] = useState<WalletTransaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: WalletTransaction[] }>("/api/user/wallet/transactions?limit=100", token)
      .then((res) => setTransactions(res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const summary = useMemo(() => {
    return transactions.reduce(
      (acc, tx) => {
        if (tx.amount > 0) acc.income += tx.amount;
        if (tx.amount < 0) acc.expense += Math.abs(tx.amount);
        return acc;
      },
      { income: 0, expense: 0 },
    );
  }, [transactions]);

  return (
    <div>
      <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 className="text-2xl font-bold">钱包流水</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            查看余额变动、API 扣费、管理员调账和失败扣费审计。
          </p>
        </div>
        <a href="/dashboard/topup" className="text-sm text-primary hover:underline">去充值</a>
      </div>

      <div className="mt-6 grid gap-4 md:grid-cols-3">
        <MetricCard label="本页入账" value={formatMoney(summary.income)} tone="green" />
        <MetricCard label="本页出账" value={formatMoney(summary.expense)} tone="red" />
        <MetricCard label="流水条数" value={`${transactions.length}`} />
      </div>

      <div className="mt-6 overflow-x-auto">
        {loading ? (
          <div className="py-16 text-center text-muted-foreground">加载中...</div>
        ) : transactions.length === 0 ? (
          <div className="card-glow rounded-xl p-16 text-center">
            <div className="text-4xl mb-4">💳</div>
            <p className="font-medium text-foreground">暂无钱包流水</p>
            <p className="mt-1 text-sm text-muted-foreground">产生 API 消费或管理员调账后会显示在这里。</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-3 text-left font-medium">时间</th>
                <th className="px-3 py-3 text-left font-medium">类型</th>
                <th className="px-3 py-3 text-right font-medium">金额</th>
                <th className="px-3 py-3 text-right font-medium">调整前</th>
                <th className="px-3 py-3 text-right font-medium">调整后</th>
                <th className="px-3 py-3 text-left font-medium">关联</th>
                <th className="px-3 py-3 text-left font-medium">说明</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => (
                <tr key={tx.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                  <td className="px-3 py-3 text-xs text-muted-foreground whitespace-nowrap">
                    {formatTime(tx.created_at)}
                  </td>
                  <td className="px-3 py-3"><TypeBadge type={tx.type} /></td>
                  <td className={`px-3 py-3 text-right font-mono ${tx.amount < 0 ? "text-red-400" : tx.amount > 0 ? "text-emerald-400" : "text-muted-foreground"}`}>
                    {formatSignedMoney(tx.amount)}
                  </td>
                  <td className="px-3 py-3 text-right text-muted-foreground">{formatOptionalMoney(tx.balance_before)}</td>
                  <td className="px-3 py-3 text-right text-muted-foreground">{formatOptionalMoney(tx.balance_after)}</td>
                  <td className="px-3 py-3 text-xs text-muted-foreground">
                    {tx.request_log_id ? `请求 #${tx.request_log_id}` : tx.reference_type ? `${tx.reference_type}:${tx.reference_id || "-"}` : "-"}
                  </td>
                  <td className="px-3 py-3 text-xs text-muted-foreground max-w-xs truncate" title={tx.note || ""}>{tx.note || "-"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

function MetricCard({ label, value, tone }: { label: string; value: string; tone?: "green" | "red" }) {
  const color = tone === "green" ? "text-emerald-400" : tone === "red" ? "text-red-400" : "text-foreground";
  return (
    <div className="card-glow rounded-xl p-4">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={`mt-2 text-xl font-semibold ${color}`}>{value}</div>
    </div>
  );
}

function TypeBadge({ type }: { type: string }) {
  const label = typeLabel(type);
  const danger = type.includes("failed");
  const debit = type === "api_charge";
  const credit = type === "admin_adjustment";
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs ${
      danger ? "bg-yellow-500/10 text-yellow-400" :
      debit ? "bg-red-500/10 text-red-400" :
      credit ? "bg-primary/10 text-primary" :
      "bg-muted text-muted-foreground"
    }`}>
      {label}
    </span>
  );
}

function typeLabel(type: string) {
  switch (type) {
    case "api_charge": return "API 消费";
    case "api_charge_failed": return "扣费失败";
    case "admin_adjustment": return "管理员调账";
    default: return type;
  }
}

function formatMoney(value: number) {
  return `¥${(value / 1000).toFixed(4)}`;
}

function formatSignedMoney(value: number) {
  const prefix = value > 0 ? "+" : value < 0 ? "-" : "";
  return `${prefix}${formatMoney(Math.abs(value))}`;
}

function formatOptionalMoney(value?: number) {
  if (value === undefined || value === null) return "-";
  return formatMoney(value);
}

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}
