"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api";
import { FALLBACK_MODELS, PROVIDER_LABELS, type Model } from "@/lib/fallback-models";

const PROVIDERS = ["all", "openai", "anthropic", "google", "deepseek"] as const;

const PROVIDER_COLORS: Record<string, string> = {
  openai: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  anthropic: "bg-orange-500/10 text-orange-400 border-orange-500/20",
  google: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  deepseek: "bg-purple-500/10 text-purple-400 border-purple-500/20",
};

export default function ModelsPage() {
  const [models, setModels] = useState<Model[]>(FALLBACK_MODELS);
  const [provider, setProvider] = useState<string>("all");
  const [search, setSearch] = useState("");

  useEffect(() => {
    api.get<{ data: Model[] }>("/api/public/models")
      .then((res) => { if (res.data && res.data.length > 0) setModels(res.data); })
      .catch(() => {});
  }, []);

  const filtered = models.filter((m) => {
    if (provider !== "all" && m.provider !== provider) return false;
    if (search && !m.display_name.toLowerCase().includes(search.toLowerCase()) && !m.model_id.includes(search.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">模型广场</h1>
        <p className="mt-2 text-muted-foreground">浏览所有可用的 AI 模型，查看能力、价格和调用方式</p>
      </div>

      <div className="mb-8 flex flex-col gap-4 md:flex-row md:items-center">
        <input
          type="text" placeholder="搜索模型..." value={search} onChange={(e) => setSearch(e.target.value)}
          className="rounded-lg border border-border bg-background px-4 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none"
        />
        <div className="flex gap-2">
          {PROVIDERS.map((p) => (
            <button key={p} onClick={() => setProvider(p)}
              className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors ${
                provider === p ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:text-foreground"
              }`}>
              {p === "all" ? "全部" : p.charAt(0).toUpperCase() + p.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filtered.map((model) => (
          <a key={model.model_id} href={`/models/${model.model_id}`}
            className="card-glow group rounded-xl p-5 transition-all hover:border-primary/50">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-semibold text-foreground group-hover:text-primary transition-colors">{model.display_name}</h3>
                <span className={`mt-1 inline-block rounded-md border px-2 py-0.5 text-xs ${PROVIDER_COLORS[model.provider] || "bg-muted text-muted-foreground"}`}>{PROVIDER_LABELS[model.provider] ?? model.provider}</span>
              </div>
              <span className="text-xs text-muted-foreground">{(model.context_length / 1000).toFixed(0)}K ctx</span>
            </div>
            <p className="mt-3 text-sm text-muted-foreground line-clamp-2">{model.description}</p>
            <div className="mt-4 flex flex-wrap gap-1.5">
              {model.capabilities?.slice(0, 4).map((cap) => (
                <span key={cap} className="rounded-md bg-muted px-2 py-0.5 text-xs text-muted-foreground">{cap}</span>
              ))}
            </div>
            <div className="mt-4 flex items-center justify-between border-t border-border pt-3 text-xs text-muted-foreground">
              <span>输入: ${(model.input_price / 1000).toFixed(2)}/{model.price_unit}</span>
              <span>输出: ${(model.output_price / 1000).toFixed(2)}/{model.price_unit}</span>
            </div>
          </a>
        ))}
      </div>

      {filtered.length === 0 && (
        <div className="py-16 text-center text-muted-foreground">没有找到匹配的模型</div>
      )}
    </div>
  );
}
