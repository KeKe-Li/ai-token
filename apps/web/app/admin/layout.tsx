"use client";

import { usePathname } from "next/navigation";
import { useAuthGuard } from "@/lib/auth-guard";

const NAV_ITEMS = [
  { href: "/admin", label: "概览" },
  { href: "/admin/channels", label: "渠道管理" },
  { href: "/admin/models", label: "模型管理" },
  { href: "/admin/users", label: "用户管理" },
  { href: "/admin/logs", label: "全局日志" },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { checked } = useAuthGuard(true);
  const pathname = usePathname();

  if (!checked) return null;

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="mb-6 flex items-center gap-2">
        <span className="rounded-md bg-destructive/10 px-2 py-0.5 text-xs font-medium text-destructive">
          管理员
        </span>
        <h1 className="text-lg font-semibold">后台管理</h1>
      </div>
      <div className="flex flex-col md:flex-row gap-8">
        <aside className="md:w-48 shrink-0">
          <nav className="flex md:flex-col gap-1">
            {NAV_ITEMS.map((item) => (
              <a
                key={item.href}
                href={item.href}
                className={`rounded-lg px-3 py-2 text-sm transition-colors ${
                  pathname === item.href
                    ? "bg-primary/10 text-primary font-medium"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted"
                }`}
              >
                {item.label}
              </a>
            ))}
          </nav>
        </aside>
        <div className="flex-1 min-w-0">{children}</div>
      </div>
    </div>
  );
}
