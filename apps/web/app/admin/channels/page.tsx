"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api";

type Channel = { id: number; name: string; provider: string; base_url: string; models: string[]; status: number; priority: number; weight: number; created_at: string };

const PROVIDERS = [
  { value: "openai", label: "OpenAI", url: "https://api.openai.com" },
  { value: "anthropic", label: "Anthropic", url: "https://api.anthropic.com" },
  { value: "google", label: "Google", url: "https://generativelanguage.googleapis.com" },
  { value: "deepseek", label: "DeepSeek", url: "https://api.deepseek.com" },
];

export default function ChannelsPage() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({ name: "", provider: "openai", base_url: "https://api.openai.com", api_key: "", models: "", priority: 0, weight: 1 });

  useEffect(() => { load(); }, []);

  async function load() {
    const token = localStorage.getItem("token");
    if (!token) return;
    try { const res = await api.get<{ data: Channel[] }>("/api/admin/channels", token); setChannels(res.data || []); } catch {}
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    const token = localStorage.getItem("token");
    try {
      await api.post("/api/admin/channels", { ...form, models: form.models.split(",").map((m) => m.trim()).filter(Boolean), priority: Number(form.priority), weight: Number(form.weight), rate_limit: 0 }, token || undefined);
      setShowForm(false);
      setForm({ name: "", provider: "openai", base_url: "https://api.openai.com", api_key: "", models: "", priority: 0, weight: 1 });
      load();
    } catch (err) { alert(err instanceof Error ? err.message : "创建失败"); }
    finally { setLoading(false); }
  }

  async function handleDelete(id: number) {
    if (!confirm("确定删除此渠道？")) return;
    const token = localStorage.getItem("token");
    try { await api.delete(`/api/admin/channels/${id}`, token || undefined); load(); } catch {}
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">渠道管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">配置供应商 API 渠道</p>
        </div>
        <button onClick={() => setShowForm(true)} className="btn-glow rounded-lg px-4 py-2 text-sm font-medium text-white">+ 添加渠道</button>
      </div>

      {showForm && (
        <div className="mt-6 card-glow rounded-xl p-6">
          <h3 className="text-lg font-semibold">新建渠道</h3>
          <form onSubmit={handleCreate} className="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium mb-1.5">名称 *</label>
              <input type="text" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="OpenAI 主力" className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">供应商 *</label>
              <select value={form.provider} onChange={(e) => { const p = PROVIDERS.find((x) => x.value === e.target.value); setForm({ ...form, provider: e.target.value, base_url: p?.url || "" }); }} className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
                {PROVIDERS.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">Base URL *</label>
              <input type="url" required value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">API Key *</label>
              <input type="password" required value={form.api_key} onChange={(e) => setForm({ ...form, api_key: e.target.value })} placeholder="sk-..." className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium mb-1.5">支持模型（逗号分隔）*</label>
              <input type="text" required value={form.models} onChange={(e) => setForm({ ...form, models: e.target.value })} placeholder="gpt-4o, gpt-4o-mini" className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">优先级</label>
              <input type="number" value={form.priority} onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })} className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">权重</label>
              <input type="number" value={form.weight} onChange={(e) => setForm({ ...form, weight: Number(e.target.value) })} className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div className="md:col-span-2 flex gap-3">
              <button type="submit" disabled={loading} className="btn-glow rounded-lg px-6 py-2.5 text-sm font-medium text-white disabled:opacity-50">{loading ? "创建中..." : "保存"}</button>
              <button type="button" onClick={() => setShowForm(false)} className="rounded-lg border border-border px-6 py-2.5 text-sm hover:bg-muted">取消</button>
            </div>
          </form>
        </div>
      )}

      <div className="mt-8">
        {channels.length === 0 ? (
          <div className="card-glow rounded-xl p-16 text-center">
            <div className="text-4xl mb-4">🔌</div>
            <p className="text-foreground font-medium">暂无渠道配置</p>
            <p className="mt-1 text-sm text-muted-foreground">添加供应商渠道后，API 请求将通过这些渠道转发</p>
          </div>
        ) : (
          <div className="space-y-3">
            {channels.map((ch) => (
              <div key={ch.id} className="card-glow rounded-xl p-5">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-3">
                      <h3 className="font-medium">{ch.name}</h3>
                      <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs text-primary">{ch.provider}</span>
                      <span className={`rounded-full px-2 py-0.5 text-xs ${ch.status === 1 ? "text-emerald-400 bg-emerald-500/10" : "text-red-400 bg-red-500/10"}`}>
                        {ch.status === 1 ? "启用" : "禁用"}
                      </span>
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">{ch.base_url}</div>
                    <div className="mt-2 flex flex-wrap gap-1">
                      {ch.models.map((m) => <span key={m} className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground font-mono">{m}</span>)}
                    </div>
                    <div className="mt-2 text-xs text-muted-foreground">优先级: {ch.priority} · 权重: {ch.weight}</div>
                  </div>
                  <button onClick={() => handleDelete(ch.id)} className="shrink-0 rounded-lg border border-destructive/20 px-3 py-1.5 text-xs text-destructive hover:bg-destructive/10">删除</button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
