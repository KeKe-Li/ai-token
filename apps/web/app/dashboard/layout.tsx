"use client";

import { usePathname } from "next/navigation";
import { useAuthGuard } from "@/lib/auth-guard";

const NAV_ITEMS = [
  { href: "/dashboard", label: "概览" },
  { href: "/dashboard/playground", label: "Playground" },
  { href: "/dashboard/keys", label: "API Keys" },
  { href: "/dashboard/logs", label: "调用日志" },
  { href: "/dashboard/usage", label: "用量统计" },
  { href: "/dashboard/topup", label: "充值" },
  { href: "/dashboard/settings", label: "账户设置" },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { checked } = useAuthGuard(true);
  const pathname = usePathname();

  if (!checked) return null;

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="flex flex-col md:flex-row gap-8">
        <aside className="md:w-56 shrink-0">
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
