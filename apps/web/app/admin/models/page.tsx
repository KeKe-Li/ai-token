export default function AdminModelsPage() {
  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">模型管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">配置模型信息和定价</p>
        </div>
        <button className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
          添加模型
        </button>
      </div>

      <div className="mt-8 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="px-4 py-3 text-left font-medium">模型 ID</th>
              <th className="px-4 py-3 text-left font-medium">供应商</th>
              <th className="px-4 py-3 text-left font-medium">类别</th>
              <th className="px-4 py-3 text-right font-medium">输入价格</th>
              <th className="px-4 py-3 text-right font-medium">输出价格</th>
              <th className="px-4 py-3 text-center font-medium">状态</th>
            </tr>
          </thead>
          <tbody>
            <ModelRow id="gpt-4o" provider="openai" category="llm" input="$2.50/1M" output="$10.00/1M" />
            <ModelRow id="gpt-4o-mini" provider="openai" category="llm" input="$0.15/1M" output="$0.60/1M" />
            <ModelRow id="claude-sonnet-4-6" provider="anthropic" category="llm" input="$3.00/1M" output="$15.00/1M" />
            <ModelRow id="claude-haiku-4-5" provider="anthropic" category="llm" input="$0.80/1M" output="$4.00/1M" />
            <ModelRow id="gemini-2.5-pro" provider="google" category="llm" input="$1.25/1M" output="$10.00/1M" />
            <ModelRow id="gemini-2.5-flash" provider="google" category="llm" input="$0.15/1M" output="$0.60/1M" />
            <ModelRow id="deepseek-chat" provider="deepseek" category="llm" input="$0.27/1M" output="$1.10/1M" />
            <ModelRow id="deepseek-reasoner" provider="deepseek" category="llm" input="$0.55/1M" output="$2.19/1M" />
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ModelRow({ id, provider, category, input, output }: {
  id: string; provider: string; category: string; input: string; output: string;
}) {
  return (
    <tr className="border-b border-border hover:bg-muted/50 transition-colors">
      <td className="px-4 py-3 font-mono text-xs">{id}</td>
      <td className="px-4 py-3 text-muted-foreground">{provider}</td>
      <td className="px-4 py-3 text-muted-foreground">{category}</td>
      <td className="px-4 py-3 text-right text-muted-foreground">{input}</td>
      <td className="px-4 py-3 text-right text-muted-foreground">{output}</td>
      <td className="px-4 py-3 text-center">
        <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-400">可用</span>
      </td>
    </tr>
  );
}
