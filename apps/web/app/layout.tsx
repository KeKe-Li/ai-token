import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AI Token - AI 模型聚合中转平台",
  description: "一个 API，接入所有主流 AI 模型。统一接口、智能路由、自动故障切换。",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" className="dark">
      <body className="min-h-screen antialiased">
        <Header />
        <main className="flex-1">{children}</main>
        <Footer />
      </body>
    </html>
  );
}

function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background/80 backdrop-blur-sm">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <a href="/" className="text-xl font-bold text-primary">
            AI Token
          </a>
          <nav className="hidden md:flex items-center gap-6 text-sm text-muted-foreground">
            <a href="/models" className="hover:text-foreground transition-colors">
              模型广场
            </a>
            <a href="/pricing" className="hover:text-foreground transition-colors">
              价格
            </a>
            <a href="/docs" className="hover:text-foreground transition-colors">
              文档
            </a>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <a
            href="/login"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            登录
          </a>
          <a
            href="/register"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            免费开始
          </a>
        </div>
      </div>
    </header>
  );
}

function Footer() {
  return (
    <footer className="border-t border-border bg-background py-12">
      <div className="mx-auto max-w-7xl px-6">
        <div className="flex flex-col md:flex-row justify-between gap-8">
          <div>
            <h3 className="text-lg font-bold text-primary">AI Token</h3>
            <p className="mt-2 text-sm text-muted-foreground max-w-xs">
              面向开发者的 AI 模型聚合中转平台。一个 API，接入所有主流模型。
            </p>
          </div>
          <div className="grid grid-cols-2 gap-8 text-sm">
            <div>
              <h4 className="font-medium text-foreground">产品</h4>
              <ul className="mt-3 space-y-2 text-muted-foreground">
                <li><a href="/models" className="hover:text-foreground">模型广场</a></li>
                <li><a href="/pricing" className="hover:text-foreground">价格</a></li>
                <li><a href="/docs" className="hover:text-foreground">API 文档</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-medium text-foreground">支持</h4>
              <ul className="mt-3 space-y-2 text-muted-foreground">
                <li><a href="/docs" className="hover:text-foreground">开发文档</a></li>
                <li><a href="https://github.com/KeKe-Li/ai-token" className="hover:text-foreground">GitHub</a></li>
              </ul>
            </div>
          </div>
        </div>
        <div className="mt-8 border-t border-border pt-8 text-center text-sm text-muted-foreground">
          &copy; 2026 AI Token. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
