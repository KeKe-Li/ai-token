"use client";

export default function SettingsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold">账户设置</h1>
      <p className="mt-1 text-sm text-muted-foreground">管理你的账户信息</p>

      <div className="mt-8 space-y-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">基本信息</h2>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium mb-1.5">用户名</label>
              <input
                type="text"
                disabled
                value="admin"
                className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">邮箱</label>
              <input
                type="email"
                disabled
                value="admin@aitoken.dev"
                className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground"
              />
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">修改密码</h2>
          <form className="mt-4 max-w-md space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1.5">当前密码</label>
              <input
                type="password"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">新密码</label>
              <input
                type="password"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">确认新密码</label>
              <input
                type="password"
                className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
            <button
              type="submit"
              className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              更新密码
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
