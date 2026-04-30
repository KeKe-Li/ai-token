import { notFound } from "next/navigation";

const MODELS: Record<string, {
  display_name: string;
  provider: string;
  category: string;
  context_length: number;
  input_price: number;
  output_price: number;
  price_unit: string;
  capabilities: string[];
  description: string;
}> = {
  "gpt-4o": { display_name: "GPT-4o", provider: "openai", category: "llm", context_length: 128000, input_price: 2500, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "json_mode"], description: "OpenAI 最新旗舰多模态模型，支持文本、图像输入，具备强大的推理和编程能力。" },
  "gpt-4o-mini": { display_name: "GPT-4o Mini", provider: "openai", category: "llm", context_length: 128000, input_price: 150, output_price: 600, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "json_mode"], description: "GPT-4o 的轻量版本，性价比极高，适合大部分日常任务。" },
  "claude-sonnet-4-6": { display_name: "Claude Sonnet 4.6", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Anthropic 最新编程优化模型，200K 上下文，支持扩展思考。" },
  "claude-haiku-4-5": { display_name: "Claude Haiku 4.5", provider: "anthropic", category: "llm", context_length: 200000, input_price: 800, output_price: 4000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Anthropic 快速轻量模型，适合高频低延迟场景。" },
  "gemini-2.5-pro": { display_name: "Gemini 2.5 Pro", provider: "google", category: "llm", context_length: 1000000, input_price: 1250, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Google 最新旗舰模型，100 万 token 超长上下文，支持深度思考。" },
  "gemini-2.5-flash": { display_name: "Gemini 2.5 Flash", provider: "google", category: "llm", context_length: 1000000, input_price: 150, output_price: 600, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Google 快速模型，100 万上下文，极高性价比。" },
  "deepseek-chat": { display_name: "DeepSeek V3", provider: "deepseek", category: "llm", context_length: 64000, input_price: 270, output_price: 1100, price_unit: "1M", capabilities: ["function_call", "streaming", "json_mode"], description: "DeepSeek 通用对话模型，中文能力优秀，性价比极高。" },
  "deepseek-reasoner": { display_name: "DeepSeek R1", provider: "deepseek", category: "llm", context_length: 64000, input_price: 550, output_price: 2190, price_unit: "1M", capabilities: ["streaming", "thinking"], description: "DeepSeek 深度推理模型，适合数学、逻辑和复杂问题求解。" },
};

export default async function ModelDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const model = MODELS[id];

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
              {model.provider}
            </span>
            <span className="text-sm text-muted-foreground">{model.category.toUpperCase()}</span>
            <span className="text-sm text-muted-foreground">
              {(model.context_length / 1000)}K 上下文
            </span>
          </div>
        </div>
      </div>

      <p className="mt-6 text-muted-foreground leading-relaxed">{model.description}</p>

      {/* 能力标签 */}
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

      {/* 价格 */}
      <div className="mt-8">
        <h2 className="text-lg font-semibold">价格</h2>
        <div className="mt-3 grid grid-cols-2 gap-4">
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="text-sm text-muted-foreground">输入价格</div>
            <div className="mt-1 text-xl font-bold">${(model.input_price / 1000).toFixed(2)} / {model.price_unit} tokens</div>
          </div>
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="text-sm text-muted-foreground">输出价格</div>
            <div className="mt-1 text-xl font-bold">${(model.output_price / 1000).toFixed(2)} / {model.price_unit} tokens</div>
          </div>
        </div>
      </div>

      {/* 调用示例 */}
      <div className="mt-8">
        <h2 className="text-lg font-semibold">调用示例</h2>
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-background">
          <div className="border-b border-border px-4 py-2">
            <span className="text-xs text-muted-foreground">Python (OpenAI SDK)</span>
          </div>
          <pre className="overflow-x-auto p-4 text-sm leading-relaxed">
            <code>{`from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="https://api.aitoken.dev/v1"
)

response = client.chat.completions.create(
    model="${id}",
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
