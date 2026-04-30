export default function UsagePage() {
  return (
    <div>
      <h1 className="text-2xl font-bold">用量统计</h1>
      <p className="mt-1 text-sm text-muted-foreground">按模型查看 Token 消耗和费用明细</p>

      <div className="mt-8 grid gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="text-sm text-muted-foreground">本月总费用</div>
          <div className="mt-1 text-2xl font-bold">¥0.00</div>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="text-sm text-muted-foreground">本月总请求</div>
          <div className="mt-1 text-2xl font-bold">0</div>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="text-sm text-muted-foreground">本月 Tokens</div>
          <div className="mt-1 text-2xl font-bold">0</div>
        </div>
      </div>

      <div className="mt-8">
        <h2 className="text-lg font-semibold">按模型分组</h2>
        <div className="mt-4 rounded-xl border border-border bg-card p-12 text-center">
          <p className="text-muted-foreground">暂无用量数据</p>
        </div>
      </div>
    </div>
  );
}
