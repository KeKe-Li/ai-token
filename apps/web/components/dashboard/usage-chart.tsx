"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type DayUsage = { date: string; requests: number; tokens: number; cost: number };

export function UsageChart() {
  const [data, setData] = useState<DayUsage[]>([]);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: DayUsage[] }>("/api/user/usage/daily?days=7", token)
      .then((res) => setData(res.data || []))
      .catch(() => generateMockData());
  }, []);

  function generateMockData() {
    const days: DayUsage[] = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      days.push({ date: d.toISOString().slice(5, 10), requests: 0, tokens: 0, cost: 0 });
    }
    setData(days);
  }

  if (data.length === 0) return null;

  const maxTokens = Math.max(...data.map((d) => d.tokens), 1);

  return (
    <div className="card-glow rounded-xl p-6">
      <h3 className="text-sm font-medium text-muted-foreground mb-4">近 7 天用量趋势</h3>
      <div className="flex items-end gap-1 h-32">
        {data.map((d, i) => {
          const height = Math.max((d.tokens / maxTokens) * 100, 2);
          return (
            <div key={i} className="flex-1 flex flex-col items-center gap-1">
              <div className="w-full relative group">
                <div
                  className="w-full rounded-t bg-gradient-to-t from-primary/60 to-primary transition-all hover:from-primary/80 hover:to-primary"
                  style={{ height: `${height}%`, minHeight: "2px" }}
                />
                <div className="absolute -top-8 left-1/2 -translate-x-1/2 hidden group-hover:block rounded bg-card border border-border px-2 py-1 text-xs whitespace-nowrap shadow-lg z-10">
                  {d.tokens.toLocaleString()} tokens
                </div>
              </div>
              <span className="text-[10px] text-muted-foreground">{d.date}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
