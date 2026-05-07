"use client";

import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";

type BillingEvent = {
  id: number;
  user_id: number;
  api_key_id: number;
  request_log_id?: number;
  wallet_hold_id?: number;
  event_type: string;
  model?: string;
  amount: number;
  balance: number;
  reserved_balance: number;
  available_balance: number;
  status: string;
  note?: string;
  created_at: string;
};

export default function AdminBillingEventsPage() {
  const [events, setEvents] = useState<BillingEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [userID, setUserID] = useState("");
  const [walletHoldID, setWalletHoldID] = useState<string | null>(null);

  const load = () => {
    const token = localStorage.getItem("token");
    if (!token) return;
    setLoading(true);
    const params = new URLSearchParams({ limit: "100" });
    if (userID.trim()) params.set("user_id", userID.trim());
    if (typeof window !== "undefined") {
      const holdID = new URLSearchParams(window.location.search).get("wallet_hold_id");
      if (holdID) {
        params.set("wallet_hold_id", holdID);
        setWalletHoldID(holdID);
      }
    }
    api.get<{ data: BillingEvent[] }>(`/api/admin/billing-events?${params.toString()}`, token)
      .then((res) => setEvents(res.data || []))
      .catch(() => setEvents([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const summary = useMemo(() => {
    return events.reduce(
      (acc, event) => {
        if (event.event_type.includes("failed") || event.status === "failed") acc.failed += 1;
        if (event.event_type === "preauth_held") acc.held += 1;
        if (event.event_type === "hold_captured") acc.captured += 1;
        if (event.event_type === "hold_released") acc.released += 1;
        return acc;
      },
      { failed: 0, held: 0, captured: 0, released: 0 },
    );
  }, [events]);

  return (
    <div>
      <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
        <div>
          <h2 className="text-2xl font-bold">账务事件</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            审计预授权、capture、release、缺价拒绝和失败事件，辅助定位账务状态。
          </p>
        </div>
        <div className="flex items-center gap-2">
          <input
            value={userID}
            onChange={(event) => setUserID(event.target.value)}
            placeholder="按用户 ID 过滤"
            className="w-36 rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
          />
          <button onClick={load} className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white">查询</button>
        </div>
      </div>

      {walletHoldID && (
        <div className="mt-4 flex flex-wrap items-center gap-3 rounded-xl border border-primary/30 bg-primary/5 px-4 py-3 text-sm">
          <span className="text-muted-foreground">当前过滤：</span>
          <span className="font-mono text-primary">Hold #{walletHoldID}</span>
          <a className="text-primary hover:underline" href="/admin/billing-events">清除过滤</a>
        </div>
      )}

      <div className="mt-6 grid gap-4 md:grid-cols-4">
        <MetricCard label="预授权" value={`${summary.held}`} />
        <MetricCard label="Capture" value={`${summary.captured}`} tone="green" />
        <MetricCard label="Release" value={`${summary.released}`} tone="yellow" />
        <MetricCard label="失败" value={`${summary.failed}`} tone="red" />
      </div>

      <div className="mt-6 overflow-x-auto">
        {loading ? (
          <div className="py-16 text-center text-muted-foreground">加载中...</div>
        ) : events.length === 0 ? (
          <div className="card-glow rounded-xl p-16 text-center">
            <div className="text-4xl mb-4">🧮</div>
            <p className="font-medium text-foreground">暂无账务事件</p>
            <p className="mt-1 text-sm text-muted-foreground">有预授权、扣费结算或失败拒绝后会显示在这里。</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-3 text-left font-medium">时间</th>
                <th className="px-3 py-3 text-left font-medium">用户</th>
                <th className="px-3 py-3 text-left font-medium">事件</th>
                <th className="px-3 py-3 text-left font-medium">模型</th>
                <th className="px-3 py-3 text-right font-medium">金额</th>
                <th className="px-3 py-3 text-right font-medium">余额</th>
                <th className="px-3 py-3 text-right font-medium">冻结</th>
                <th className="px-3 py-3 text-right font-medium">可用</th>
                <th className="px-3 py-3 text-left font-medium">关联</th>
                <th className="px-3 py-3 text-left font-medium">说明</th>
              </tr>
            </thead>
            <tbody>
              {events.map((event) => (
                <tr key={event.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                  <td className="px-3 py-3 text-xs text-muted-foreground whitespace-nowrap">{formatTime(event.created_at)}</td>
                  <td className="px-3 py-3 text-xs">#{event.user_id}</td>
                  <td className="px-3 py-3"><EventBadge eventType={event.event_type} status={event.status} /></td>
                  <td className="px-3 py-3 font-mono text-xs text-muted-foreground">{event.model || "-"}</td>
                  <td className="px-3 py-3 text-right font-mono">{formatMoney(event.amount)}</td>
                  <td className="px-3 py-3 text-right text-muted-foreground">{formatMoney(event.balance)}</td>
                  <td className="px-3 py-3 text-right text-yellow-400">{formatMoney(event.reserved_balance)}</td>
                  <td className="px-3 py-3 text-right text-emerald-400">{formatMoney(event.available_balance)}</td>
                  <td className="px-3 py-3 text-xs text-muted-foreground">
                    {event.request_log_id ? (
                      <a className="text-primary hover:underline" href={`/admin/logs?request_log_id=${event.request_log_id}`}>请求 #{event.request_log_id}</a>
                    ) : event.wallet_hold_id ? (
                      <a className="text-primary hover:underline" href={`/admin/billing-events?wallet_hold_id=${event.wallet_hold_id}`}>Hold #{event.wallet_hold_id}</a>
                    ) : (
                      "-"
                    )}
                  </td>
                  <td className="px-3 py-3 text-xs text-muted-foreground max-w-xs truncate" title={event.note || ""}>{event.note || "-"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

function MetricCard({ label, value, tone }: { label: string; value: string; tone?: "green" | "yellow" | "red" }) {
  const color = tone === "green" ? "text-emerald-400" : tone === "yellow" ? "text-yellow-400" : tone === "red" ? "text-red-400" : "text-foreground";
  return (
    <div className="card-glow rounded-xl p-4">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={`mt-2 text-xl font-semibold ${color}`}>{value}</div>
    </div>
  );
}

function EventBadge({ eventType, status }: { eventType: string; status: string }) {
  const failed = eventType.includes("failed") || status === "failed";
  const captured = eventType === "hold_captured";
  const released = eventType === "hold_released";
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs ${
      failed ? "bg-red-500/10 text-red-400" :
      captured ? "bg-emerald-500/10 text-emerald-400" :
      released ? "bg-yellow-500/10 text-yellow-400" :
      "bg-primary/10 text-primary"
    }`}>
      {eventLabel(eventType)}
    </span>
  );
}

function eventLabel(eventType: string) {
  switch (eventType) {
    case "preauth_held": return "预授权";
    case "preauth_failed": return "预授权失败";
    case "model_unpriced": return "模型缺价";
    case "preauth_error": return "预授权异常";
    case "hold_captured": return "Capture";
    case "hold_released": return "Release";
    case "capture_failed": return "Capture 失败";
    default: return eventType;
  }
}

function formatMoney(value: number) {
  return `¥${(value / 1000).toFixed(4)}`;
}

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}
