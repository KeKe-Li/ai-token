"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type UserProfile = { id: number; username: string; email: string; role: number; balance: number; created_at: string };

export default function SettingsPage() {
  const [user, setUser] = useState<UserProfile | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) return;
    api.get<UserProfile>("/api/user/profile", token).then(setUser).catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold">账户设置</h1>
      <p className="mt-1 text-sm text-muted-foreground">管理你的账户信息</p>

      <div className="mt-8 space-y-6">
        <div className="card-glow rounded-xl p-6">
          <h2 className="text-lg font-semibold">基本信息</h2>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium mb-1.5">用户名</label>
              <input type="text" disabled value={user?.username ?? ""} className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">邮箱</label>
              <input type="email" disabled value={user?.email ?? ""} className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">角色</label>
              <input type="text" disabled value={user && user.role >= 10 ? "管理员" : "普通用户"} className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">注册时间</label>
              <input type="text" disabled value={user?.created_at ? new Date(user.created_at).toLocaleDateString("zh-CN") : ""} className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm text-muted-foreground" />
            </div>
          </div>
        </div>

        <div className="card-glow rounded-xl p-6">
          <h2 className="text-lg font-semibold">修改密码</h2>
          <form className="mt-4 max-w-md space-y-4" onSubmit={(e) => { e.preventDefault(); alert("密码修改功能即将上线"); }}>
            <div>
              <label className="block text-sm font-medium mb-1.5">当前密码</label>
              <input type="password" className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">新密码</label>
              <input type="password" className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5">确认新密码</label>
              <input type="password" className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none" />
            </div>
            <button type="submit" className="btn-glow rounded-lg px-4 py-2 text-sm font-medium text-white">更新密码</button>
          </form>
        </div>
      </div>
    </div>
  );
}
