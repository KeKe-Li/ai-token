export default function AdminPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold">系统概览</h2>
      <p className="mt-1 text-sm text-muted-foreground">全局统计和系统状态</p>

      <div className="mt-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard label="总用户数" value="1" />
        <StatsCard label="活跃渠道" value="0" />
        <StatsCard label="今日请求" value="0" />
        <StatsCard label="今日收入" value="¥0.00" />
      </div>
    </div>
  );
}

function StatsCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1 text-2xl font-bold">{value}</div>
    </div>
  );
}
