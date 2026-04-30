"use client";

import { useState } from "react";

type ApiKey = {
  id: number;
  name: string;
  key_prefix: string;
  status: number;
  created_at: string;
};

export default function KeysPage() {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setShowCreate(false);
    setCreatedKey("sk-demo-" + Math.random().toString(36).slice(2, 34));
    setNewKeyName("");
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">API Keys</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理你的 API 密钥</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          创建密钥
        </button>
      </div>

      {createdKey && (
        <div className="mt-6 rounded-lg border border-primary/20 bg-primary/5 p-4">
          <p className="text-sm font-medium text-foreground">密钥已创建，请立即复制保存：</p>
          <div className="mt-2 flex items-center gap-2">
            <code className="flex-1 rounded bg-background px-3 py-2 text-sm font-mono text-foreground border border-border">
              {createdKey}
            </code>
            <button
              onClick={() => { navigator.clipboard.writeText(createdKey); }}
              className="rounded-lg border border-border px-3 py-2 text-sm hover:bg-muted transition-colors"
            >
              复制
            </button>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            此密钥仅显示一次，关闭后无法再次查看完整密钥
          </p>
          <button
            onClick={() => setCreatedKey(null)}
            className="mt-2 text-xs text-primary hover:underline"
          >
            我已保存，关闭提示
          </button>
        </div>
      )}

      {showCreate && (
        <div className="mt-6 rounded-lg border border-border bg-card p-4">
          <form onSubmit={handleCreate} className="flex items-end gap-4">
            <div className="flex-1">
              <label className="block text-sm font-medium mb-1.5">密钥名称</label>
              <input
                type="text"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                required
                placeholder="例如: 生产环境"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <button
              type="submit"
              className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              确认创建
            </button>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted"
            >
              取消
            </button>
          </form>
        </div>
      )}

      <div className="mt-8">
        {keys.length === 0 ? (
          <div className="rounded-xl border border-border bg-card p-12 text-center">
            <p className="text-muted-foreground">还没有创建任何 API 密钥</p>
            <button
              onClick={() => setShowCreate(true)}
              className="mt-4 text-sm text-primary hover:underline"
            >
              创建第一个密钥
            </button>
          </div>
        ) : (
          <div className="rounded-xl border border-border overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium">名称</th>
                  <th className="px-4 py-3 text-left font-medium">密钥</th>
                  <th className="px-4 py-3 text-left font-medium">状态</th>
                  <th className="px-4 py-3 text-left font-medium">创建时间</th>
                  <th className="px-4 py-3 text-right font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {keys.map((key) => (
                  <tr key={key.id} className="border-t border-border">
                    <td className="px-4 py-3">{key.name}</td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{key.key_prefix}...</td>
                    <td className="px-4 py-3">
                      <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-400">
                        正常
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{key.created_at}</td>
                    <td className="px-4 py-3 text-right">
                      <button className="text-xs text-destructive hover:underline">删除</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
