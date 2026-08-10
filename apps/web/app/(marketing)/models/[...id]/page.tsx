import { notFound } from "next/navigation";
import { FALLBACK_MODELS, PROVIDER_LABELS, type Model } from "@/lib/fallback-models";

async function getModel(id: string): Promise<Model | null> {
  try {
    const apiUrl = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    const res = await fetch(`${apiUrl}/api/public/models/${id}`, { next: { revalidate: 300 } });
    if (!res.ok) throw new Error("api failed");
    const json = await res.json();
    return json.data || null;
  } catch {
    return FALLBACK_MODELS.find((m) => m.model_id === id) || null;
  }
}

export default async function ModelDetailPage({ params }: { params: Promise<{ id: string[] }> }) {
  const { id } = await params;
  // catch-all 路由：含斜杠的 model_id（如 deepseek/deepseek-v4-pro）会拆成多段，拼回完整 id。
  const modelId = id.join("/");
  const model = await getModel(modelId);

  if (!model) {
    notFound();
  }

  return (
    <div className="mx-auto max-w-4xl px-6 py-12">
      <a href="/models" className="text-sm text-muted-foreground hover:text-foreground mb-6 inline-block">
        ← 返回模型广场
      </a>

      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold">{model.display_name}</h1>
          <div className="mt-2 flex items-center gap-3">
            <span className="rounded-md bg-primary/10 px-2 py-0.5 text-sm text-primary">
              {PROVIDER_LABELS[model.provider] ?? model.provider}
            </span>
            <span className="text-sm text-muted-foreground">{model.category.toUpperCase()}</span>
            <span className="text-sm text-muted-foreground">
              {(model.context_length / 1000).toFixed(0)}K 上下文
            </span>
          </div>
        </div>
      </div>

      <p className="mt-6 text-muted-foreground leading-relaxed">{model.description}</p>

      {model.capabilities && model.capabilities.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold">能力</h2>
          <div className="mt-3 flex flex-wrap gap-2">
            {model.capabilities.map((cap) => (
              <span key={cap} className="rounded-lg bg-muted px-3 py-1 text-sm text-muted-foreground">
                {cap}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="mt-8">
        <h2 className="text-lg font-semibold">价格</h2>
        <div className="mt-3 grid grid-cols-2 gap-4">
          <div className="card-glow rounded-xl p-4">
            <div className="text-sm text-muted-foreground">输入价格</div>
            <div className="mt-1 text-xl font-bold">${(model.input_price / 1000).toFixed(2)} / {model.price_unit} tokens</div>
          </div>
          <div className="card-glow rounded-xl p-4">
            <div className="text-sm text-muted-foreground">输出价格</div>
            <div className="mt-1 text-xl font-bold">${(model.output_price / 1000).toFixed(2)} / {model.price_unit} tokens</div>
          </div>
        </div>
      </div>

      <div className="mt-8">
        <h2 className="text-lg font-semibold">调用示例</h2>
        <div className="mt-3 code-block rounded-xl overflow-hidden">
          <div className="border-b border-border/50 px-4 py-2">
            <span className="text-xs text-muted-foreground">Python (OpenAI SDK)</span>
          </div>
          <pre className="overflow-x-auto p-4 text-sm leading-relaxed">
            <code>{`from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="https://api.aitoken.dev/v1"
)

response = client.chat.completions.create(
    model="${model.model_id}",
    messages=[{"role": "user", "content": "你好"}],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content or "", end="")`}</code>
          </pre>
        </div>
      </div>
    </div>
  );
}
