export default function AdminUsersPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold">用户管理</h2>
      <p className="mt-1 text-sm text-muted-foreground">管理平台用户和权限</p>

      <div className="mt-8 overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="px-4 py-3 text-left font-medium">用户名</th>
              <th className="px-4 py-3 text-left font-medium">邮箱</th>
              <th className="px-4 py-3 text-left font-medium">角色</th>
              <th className="px-4 py-3 text-right font-medium">余额</th>
              <th className="px-4 py-3 text-right font-medium">请求数</th>
              <th className="px-4 py-3 text-center font-medium">状态</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border hover:bg-muted/50 transition-colors">
              <td className="px-4 py-3 font-medium">admin</td>
              <td className="px-4 py-3 text-muted-foreground">admin@aitoken.dev</td>
              <td className="px-4 py-3">
                <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs text-primary">管理员</span>
              </td>
              <td className="px-4 py-3 text-right">¥999.999</td>
              <td className="px-4 py-3 text-right text-muted-foreground">0</td>
              <td className="px-4 py-3 text-center">
                <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-400">正常</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}
