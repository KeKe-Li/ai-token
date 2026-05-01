"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api";

type ApiKey = {
  id: number;
  name: string;
  key_prefix: string;
  status: number;
  models: string[] | null;
  expires_at: string | null;
  created_at: string;
};

const EXPIRY_OPTIONS = [
  { label: "永不过期", value: "" },
  { label: "1 天", value: "1" },
  { label: "7 天", value: "7" },
  { label: "30 天", value: "30" },
  { label: "90 天", value: "90" },
  { label: "180 天", value: "180" },
  { label: "365 天", value: "365" },
  { label: "自定义", value: "custom" },
];

const MODEL_OPTIONS = [
  "gpt-4o", "gpt-4o-mini",
  "claude-sonnet-4-6", "claude-haiku-4-5",
  "gemini-2.5-pro", "gemini-2.5-flash",
  "deepseek-chat", "deepseek-reasoner",
];

export default function KeysPage() {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [expiry, setExpiry] = useState("");
  const [customDate, setCustomDate] = useState("");
  const [selectedModels, setSelectedModels] = useState<string[]>([]);
  const [allModels, setAllModels] = useState(true);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => { loadKeys(); }, []);

  async function loadKeys() {
    const token = localStorage.getItem("token");
    if (!token) return;
    try {
      const res = await api.get<{ data: ApiKey[] }>("/api/user/keys", token);
      setKeys(res.data || []);
    } catch { /* ignore */ }
  }

  function getExpiresAt(): string | null {
    if (!expiry) return null;
    if (expiry === "custom") return customDate ? new Date(customDate).toISOString() : null;
    const days = parseInt(expiry);
    const date = new Date();
    date.setDate(date.getDate() + days);
    return date.toISOString();
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    const token = localStorage.getItem("token");
    try {
      const res = await api.post<{ key: string; data: ApiKey }>("/api/user/keys", {
        name: newKeyName,
        models: allModels ? null : selectedModels,
        expires_at: getExpiresAt(),
      }, token || undefined);
      setCreatedKey(res.key);
      setShowCreate(false);
      setNewKeyName("");
      setExpiry("");
      setSelectedModels([]);
      setAllModels(true);
      loadKeys();
    } catch (err) {
      alert(err instanceof Error ? err.message : "创建失败");
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("确定要删除此密钥吗？删除后无法恢复。")) return;
    const token = localStorage.getItem("token");
    try {
      await api.delete(`/api/user/keys/${id}`, token || undefined);
      loadKeys();
    } catch { /* ignore */ }
  }

  function handleCopy(text: string) {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString("zh-CN", {
      year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit",
    });
  }

  function getExpiryStatus(expiresAt: string | null) {
    if (!expiresAt) return { label: "永不过期", color: "text-emerald-400 bg-emerald-500/10" };
    const now = new Date();
    const exp = new Date(expiresAt);
    if (exp < now) return { label: "已过期", color: "text-red-400 bg-red-500/10" };
    const days = Math.ceil((exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    if (days <= 7) return { label: `${days}天后过期`, color: "text-yellow-400 bg-yellow-500/10" };
    return { label: formatDate(expiresAt) + " 过期", color: "text-muted-foreground bg-muted" };
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">API Keys</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理你的 API 密钥，用于调用中转接口</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="btn-glow rounded-lg px-4 py-2 text-sm font-medium text-white"
        >
          + 创建密钥
        </button>
      </div>

      {createdKey && (
        <div className="mt-6 rounded-xl border border-primary/30 bg-primary/5 p-5">
          <div className="flex items-center gap-2 text-sm font-medium text-foreground">
            <span className="w-5 h-5 rounded-full bg-emerald-500/20 text-emerald-400 text-xs flex items-center justify-center">✓</span>
            密钥创建成功，请立即复制保存
          </div>
          <div className="mt-3 flex items-center gap-2">
            <code className="flex-1 rounded-lg bg-background px-4 py-2.5 text-sm font-mono text-foreground border border-border overflow-x-auto">
              {createdKey}
            </code>
            <button
              onClick={() => handleCopy(createdKey)}
              className={`shrink-0 rounded-lg px-4 py-2.5 text-sm font-medium transition-all ${
                copied ? "bg-emerald-500/20 text-emerald-400" : "bg-muted text-foreground hover:bg-muted/80"
              }`}
            >
              {copied ? "已复制 ✓" : "复制"}
            </button>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">此密钥仅显示一次，关闭后无法再次查看。</p>
          <button onClick={() => setCreatedKey(null)} className="mt-3 text-xs text-primary hover:underline">
            我已保存，关闭提示
          </button>
        </div>
      )}

      {showCreate && (
        <div className="mt-6 rounded-xl border border-border bg-card p-6">
          <h3 className="text-lg font-semibold">创建新密钥</h3>
          <form onSubmit={handleCreate} className="mt-4 space-y-5">
            <div>
              <label className="block text-sm font-medium mb-1.5">密钥名称 <span className="text-destructive">*</span></label>
              <input
                type="text" value={newKeyName} onChange={(e) => setNewKeyName(e.target.value)}
                required placeholder="例如：生产环境、测试用"
                className="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1.5">过期时间</label>
              <div className="flex flex-wrap gap-2">
                {EXPIRY_OPTIONS.map((opt) => (
                  <button key={opt.value} type="button" onClick={() => setExpiry(opt.value)}
                    className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition-all ${
                      expiry === opt.value ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:text-foreground"
                    }`}>
                    {opt.label}
                  </button>
                ))}
              </div>
              {expiry === "custom" && (
                <input type="datetime-local" value={customDate} onChange={(e) => setCustomDate(e.target.value)}
                  className="mt-2 rounded-lg border border-border bg-background px-4 py-2 text-sm focus:border-primary focus:outline-none" />
              )}
            </div>

            <div>
              <label className="block text-sm font-medium mb-1.5">模型权限</label>
              <div className="flex items-center gap-3 mb-3">
                <button type="button" onClick={() => { setAllModels(true); setSelectedModels([]); }}
                  className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition-all ${
                    allModels ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground"
                  }`}>全部模型</button>
                <button type="button" onClick={() => setAllModels(false)}
                  className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition-all ${
                    !allModels ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground"
                  }`}>指定模型</button>
              </div>
              {!allModels && (
                <div className="flex flex-wrap gap-2">
                  {MODEL_OPTIONS.map((m) => (
                    <button key={m} type="button"
                      onClick={() => setSelectedModels((prev) => prev.includes(m) ? prev.filter((x) => x !== m) : [...prev, m])}
                      className={`rounded-lg border px-3 py-1.5 text-xs font-mono transition-all ${
                        selectedModels.includes(m) ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground"
                      }`}>{m}</button>
                  ))}
                </div>
              )}
            </div>

            <div className="flex gap-3 pt-2">
              <button type="submit" disabled={loading}
                className="btn-glow rounded-lg px-6 py-2.5 text-sm font-medium text-white disabled:opacity-50">
                {loading ? "创建中..." : "确认创建"}
              </button>
              <button type="button" onClick={() => setShowCreate(false)}
                className="rounded-lg border border-border px-6 py-2.5 text-sm hover:bg-muted transition-colors">取消</button>
            </div>
          </form>
        </div>
      )}

      <div className="mt-8">
        {keys.length === 0 ? (
          <div className="rounded-xl border border-border bg-card p-16 text-center">
            <div className="text-4xl mb-4">🔑</div>
            <p className="text-foreground font-medium">还没有创建任何 API 密钥</p>
            <p className="mt-1 text-sm text-muted-foreground">创建密钥后即可调用 AI 模型</p>
            <button onClick={() => setShowCreate(true)}
              className="mt-4 btn-glow rounded-lg px-4 py-2 text-sm font-medium text-white">
              创建第一个密钥
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {keys.map((key) => {
              const expiryInfo = getExpiryStatus(key.expires_at);
              return (
                <div key={key.id} className="card-glow rounded-xl p-5">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <h3 className="font-medium text-foreground">{key.name}</h3>
                        <span className={`rounded-full px-2 py-0.5 text-xs ${
                          key.status === 1 ? "text-emerald-400 bg-emerald-500/10" : "text-red-400 bg-red-500/10"
                        }`}>{key.status === 1 ? "正常" : "已禁用"}</span>
                      </div>
                      <code className="mt-1.5 inline-block text-sm font-mono text-muted-foreground">
                        {key.key_prefix}••••••••••••
                      </code>
                      <div className="mt-2 flex items-center gap-4 text-xs text-muted-foreground flex-wrap">
                        <span>创建于 {formatDate(key.created_at)}</span>
                        <span className={`rounded-full px-2 py-0.5 ${expiryInfo.color}`}>{expiryInfo.label}</span>
                        {key.models && key.models.length > 0 && <span>{key.models.length} 个模型</span>}
                      </div>
                    </div>
                    <button onClick={() => handleDelete(key.id)}
                      className="shrink-0 rounded-lg border border-destructive/20 px-3 py-1.5 text-xs text-destructive hover:bg-destructive/10 transition-colors">
                      删除
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
