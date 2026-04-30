"use client";

import { useState } from "react";

export default function ChannelsPage() {
  const [showForm, setShowForm] = useState(false);

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">渠道管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">配置供应商 API 渠道</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          添加渠道
        </button>
      </div>

      {showForm && (
        <div className="mt-6 rounded-xl border border-border bg-card p-6">
          <h3 className="text-lg font-semibold">新建渠道</h3>
          <form className="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium mb-1.5">渠道名称</label>
              <input
                type="text"
                placeholder="例如: OpenAI 主力"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">供应商</label>
              <select className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="google">Google</option>
                <option value="deepseek">DeepSeek</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">Base URL</label>
              <input
                type="url"
                placeholder="https://api.openai.com"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">API Key</label>
              <input
                type="password"
                placeholder="sk-..."
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">优先级</label>
              <input
                type="number"
                defaultValue={0}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">权重</label>
              <input
                type="number"
                defaultValue={1}
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div className="md:col-span-2 flex gap-3">
              <button
                type="submit"
                className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                保存
              </button>
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted"
              >
                取消
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="mt-8 rounded-xl border border-border bg-card p-12 text-center">
        <p className="text-muted-foreground">暂无渠道配置</p>
        <p className="mt-2 text-sm text-muted-foreground">
          添加供应商渠道后，API 请求将通过这些渠道转发
        </p>
      </div>
    </div>
  );
}
