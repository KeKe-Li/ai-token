export default function LogsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold">调用日志</h1>
      <p className="mt-1 text-sm text-muted-foreground">查看所有 API 调用记录</p>

      <div className="mt-6 flex gap-4">
        <select className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option value="">全部模型</option>
          <option value="gpt-4o">GPT-4o</option>
          <option value="claude-sonnet-4-6">Claude Sonnet 4.6</option>
          <option value="gemini-2.5-pro">Gemini 2.5 Pro</option>
          <option value="deepseek-chat">DeepSeek V3</option>
        </select>
        <select className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option value="">全部状态</option>
          <option value="200">成功</option>
          <option value="400">客户端错误</option>
          <option value="500">服务端错误</option>
        </select>
      </div>

      <div className="mt-6 rounded-xl border border-border bg-card p-12 text-center">
        <p className="text-muted-foreground">暂无调用记录</p>
        <p className="mt-2 text-sm text-muted-foreground">
          使用 API Key 发起请求后，调用记录会显示在这里
        </p>
      </div>
    </div>
  );
}
