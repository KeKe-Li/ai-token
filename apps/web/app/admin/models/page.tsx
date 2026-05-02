"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type AIModel = { id: number; model_id: string; display_name: string; provider: string; category: string; context_length: number; input_price: number; output_price: number; price_unit: string; status: number };

export default function AdminModelsPage() {
  const [models, setModels] = useState<AIModel[]>([]);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: AIModel[] }>("/api/admin/models", token).then((res) => setModels(res.data || [])).catch(() => {});
  }, []);

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">模型管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">配置模型信息和定价</p>
        </div>
      </div>

      <div className="mt-8 overflow-x-auto">
        {models.length === 0 ? (
          <div className="card-glow rounded-xl p-12 text-center">
            <p className="text-muted-foreground">暂无模型数据（需先运行数据库迁移）</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-4 py-3 text-left font-medium">模型 ID</th>
                <th className="px-4 py-3 text-left font-medium">显示名</th>
                <th className="px-4 py-3 text-left font-medium">供应商</th>
                <th className="px-4 py-3 text-left font-medium">类别</th>
                <th className="px-4 py-3 text-right font-medium">输入价格</th>
                <th className="px-4 py-3 text-right font-medium">输出价格</th>
                <th className="px-4 py-3 text-center font-medium">状态</th>
              </tr>
            </thead>
            <tbody>
              {models.map((m) => (
                <tr key={m.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs">{m.model_id}</td>
                  <td className="px-4 py-3">{m.display_name}</td>
                  <td className="px-4 py-3">
                    <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs text-primary">{m.provider}</span>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">{m.category}</td>
                  <td className="px-4 py-3 text-right text-muted-foreground">${(m.input_price / 1000).toFixed(2)}/{m.price_unit}</td>
                  <td className="px-4 py-3 text-right text-muted-foreground">${(m.output_price / 1000).toFixed(2)}/{m.price_unit}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`rounded-full px-2 py-0.5 text-xs ${m.status === 1 ? "text-emerald-400 bg-emerald-500/10" : "text-red-400 bg-red-500/10"}`}>
                      {m.status === 1 ? "可用" : "下线"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
