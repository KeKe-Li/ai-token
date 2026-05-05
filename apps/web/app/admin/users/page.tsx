"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type User = { id: number; username: string; email: string; role: number; status: number; balance: number; reserved_balance?: number; used_amount: number; request_count: number; created_at: string };

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<{ data: User[] }>("/api/admin/users", token).then((res) => setUsers(res.data || [])).catch(() => {});
  }, []);

  return (
    <div>
      <h2 className="text-2xl font-bold">用户管理</h2>
      <p className="mt-1 text-sm text-muted-foreground">管理平台用户和权限</p>

      <div className="mt-8 overflow-x-auto">
        {users.length === 0 ? (
          <div className="card-glow rounded-xl p-12 text-center">
            <p className="text-muted-foreground">暂无用户数据</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-4 py-3 text-left font-medium">用户名</th>
                <th className="px-4 py-3 text-left font-medium">邮箱</th>
                <th className="px-4 py-3 text-left font-medium">角色</th>
                <th className="px-4 py-3 text-right font-medium">余额</th>
                <th className="px-4 py-3 text-right font-medium">冻结</th>
                <th className="px-4 py-3 text-right font-medium">消费</th>
                <th className="px-4 py-3 text-right font-medium">请求数</th>
                <th className="px-4 py-3 text-center font-medium">状态</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-b border-border hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3 font-medium">{u.username}</td>
                  <td className="px-4 py-3 text-muted-foreground">{u.email}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-md px-2 py-0.5 text-xs ${u.role >= 10 ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"}`}>
                      {u.role >= 10 ? "管理员" : "用户"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right">¥{(u.balance / 1000).toFixed(2)}</td>
                  <td className="px-4 py-3 text-right text-yellow-400">¥{((u.reserved_balance ?? 0) / 1000).toFixed(2)}</td>
                  <td className="px-4 py-3 text-right text-muted-foreground">¥{(u.used_amount / 1000).toFixed(2)}</td>
                  <td className="px-4 py-3 text-right text-muted-foreground">{u.request_count}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={`rounded-full px-2 py-0.5 text-xs ${u.status === 1 ? "text-emerald-400 bg-emerald-500/10" : "text-red-400 bg-red-500/10"}`}>
                      {u.status === 1 ? "正常" : "禁用"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
