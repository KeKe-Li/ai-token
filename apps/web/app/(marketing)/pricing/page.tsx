"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type AIModel = {
  model_id: string;
  display_name: string;
  provider: string;
  context_length: number;
  input_price: number;
  output_price: number;
  price_unit: string;
};

const FALLBACK: AIModel[] = [
  { model_id: "gpt-4o", display_name: "GPT-4o", provider: "OpenAI", context_length: 128000, input_price: 2500, output_price: 10000, price_unit: "1M" },
  { model_id: "gpt-4o-mini", display_name: "GPT-4o Mini", provider: "OpenAI", context_length: 128000, input_price: 150, output_price: 600, price_unit: "1M" },
  { model_id: "claude-sonnet-4-6", display_name: "Claude Sonnet 4.6", provider: "Anthropic", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M" },
  { model_id: "claude-haiku-4-5", display_name: "Claude Haiku 4.5", provider: "Anthropic", context_length: 200000, input_price: 800, output_price: 4000, price_unit: "1M" },
  { model_id: "gemini-2.5-pro", display_name: "Gemini 2.5 Pro", provider: "Google", context_length: 1000000, input_price: 1250, output_price: 10000, price_unit: "1M" },
  { model_id: "gemini-2.5-flash", display_name: "Gemini 2.5 Flash", provider: "Google", context_length: 1000000, input_price: 150, output_price: 600, price_unit: "1M" },
  { model_id: "deepseek-chat", display_name: "DeepSeek V3", provider: "DeepSeek", context_length: 64000, input_price: 270, output_price: 1100, price_unit: "1M" },
  { model_id: "deepseek-reasoner", display_name: "DeepSeek R1", provider: "DeepSeek", context_length: 64000, input_price: 550, output_price: 2190, price_unit: "1M" },
];

export default function PricingPage() {
  const [models, setModels] = useState<AIModel[]>(FALLBACK);

  useEffect(() => {
    const token = localStorage.getItem("token");
    api.get<{ data: AIModel[] }>("/api/admin/models", token || undefined)
      .then((res) => { if (res.data?.length) setModels(res.data); })
      .catch(() => {});
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="mb-12 text-center">
        <h1 className="text-3xl font-bold">透明定价</h1>
        <p className="mt-2 text-muted-foreground">按用量计费，无月费无最低消费，用多少付多少</p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="px-4 py-3 text-left font-medium text-foreground">模型</th>
              <th className="px-4 py-3 text-left font-medium text-foreground">供应商</th>
              <th className="px-4 py-3 text-left font-medium text-foreground">上下文</th>
              <th className="px-4 py-3 text-right font-medium text-foreground">输入价格</th>
              <th className="px-4 py-3 text-right font-medium text-foreground">输出价格</th>
            </tr>
          </thead>
          <tbody>
            {models.map((m) => (
              <tr key={m.model_id} className="border-b border-border hover:bg-muted/30 transition-colors">
                <td className="px-4 py-3 font-medium text-foreground">{m.display_name}</td>
                <td className="px-4 py-3">
                  <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs text-primary">{m.provider}</span>
                </td>
                <td className="px-4 py-3 text-muted-foreground">{(m.context_length / 1000).toFixed(0)}K</td>
                <td className="px-4 py-3 text-right text-muted-foreground">${(m.input_price / 1000).toFixed(2)}/{m.price_unit}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">${(m.output_price / 1000).toFixed(2)}/{m.price_unit}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-12 text-center text-sm text-muted-foreground">
        <p>价格按 token 用量计费，1M = 100 万 tokens。注册即送免费额度。</p>
        <a href="/register" className="mt-4 inline-block btn-glow rounded-lg px-6 py-2.5 text-sm font-medium text-white">
          立即免费开始
        </a>
      </div>
    </div>
  );
}
