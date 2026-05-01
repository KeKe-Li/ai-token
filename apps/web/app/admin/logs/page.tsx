export default function AdminLogsPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold">全局日志</h2>
      <p className="mt-1 text-sm text-muted-foreground">查看所有用户的 API 调用记录</p>

      <div className="mt-6 flex gap-4 flex-wrap">
        <select className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option value="">全部模型</option>
          <option value="gpt-4o">GPT-4o</option>
          <option value="claude-sonnet-4-6">Claude Sonnet 4.6</option>
          <option value="gemini-2.5-pro">Gemini 2.5 Pro</option>
          <option value="deepseek-chat">DeepSeek V3</option>
        </select>
        <select className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option value="">全部状态</option>
          <option value="200">成功 (2xx)</option>
          <option value="400">客户端错误 (4xx)</option>
          <option value="500">服务端错误 (5xx)</option>
        </select>
        <select className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground">
          <option value="">全部用户</option>
        </select>
      </div>

      <div className="mt-6 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="px-4 py-3 text-left font-medium">时间</th>
              <th className="px-4 py-3 text-left font-medium">用户</th>
              <th className="px-4 py-3 text-left font-medium">模型</th>
              <th className="px-4 py-3 text-left font-medium">渠道</th>
              <th className="px-4 py-3 text-right font-medium">Tokens</th>
              <th className="px-4 py-3 text-right font-medium">费用</th>
              <th className="px-4 py-3 text-right font-medium">延迟</th>
              <th className="px-4 py-3 text-center font-medium">状态</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={8} className="px-4 py-16 text-center text-muted-foreground">
                暂无调用记录
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}
