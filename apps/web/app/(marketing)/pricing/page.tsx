export default function PricingPage() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="mb-12 text-center">
        <h1 className="text-3xl font-bold">透明定价</h1>
        <p className="mt-2 text-muted-foreground">
          按用量计费，无月费无最低消费，用多少付多少
        </p>
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
            <PriceRow model="GPT-4o" provider="OpenAI" ctx="128K" input="$2.50/1M" output="$10.00/1M" />
            <PriceRow model="GPT-4o Mini" provider="OpenAI" ctx="128K" input="$0.15/1M" output="$0.60/1M" />
            <PriceRow model="Claude Sonnet 4.6" provider="Anthropic" ctx="200K" input="$3.00/1M" output="$15.00/1M" />
            <PriceRow model="Claude Haiku 4.5" provider="Anthropic" ctx="200K" input="$0.80/1M" output="$4.00/1M" />
            <PriceRow model="Gemini 2.5 Pro" provider="Google" ctx="1M" input="$1.25/1M" output="$10.00/1M" />
            <PriceRow model="Gemini 2.5 Flash" provider="Google" ctx="1M" input="$0.15/1M" output="$0.60/1M" />
            <PriceRow model="DeepSeek V3" provider="DeepSeek" ctx="64K" input="$0.27/1M" output="$1.10/1M" />
            <PriceRow model="DeepSeek R1" provider="DeepSeek" ctx="64K" input="$0.55/1M" output="$2.19/1M" />
          </tbody>
        </table>
      </div>

      <div className="mt-12 text-center text-sm text-muted-foreground">
        <p>价格按 token 用量计费，1M = 100 万 tokens。注册即送免费额度。</p>
      </div>
    </div>
  );
}

function PriceRow({
  model,
  provider,
  ctx,
  input,
  output,
}: {
  model: string;
  provider: string;
  ctx: string;
  input: string;
  output: string;
}) {
  return (
    <tr className="border-b border-border hover:bg-muted/50 transition-colors">
      <td className="px-4 py-3 font-medium text-foreground">{model}</td>
      <td className="px-4 py-3 text-muted-foreground">{provider}</td>
      <td className="px-4 py-3 text-muted-foreground">{ctx}</td>
      <td className="px-4 py-3 text-right text-muted-foreground">{input}</td>
      <td className="px-4 py-3 text-right text-muted-foreground">{output}</td>
    </tr>
  );
}
