export function Footer() {
  return (
    <footer className="border-t border-border bg-background py-12">
      <div className="mx-auto max-w-7xl px-6">
        <div className="flex flex-col md:flex-row justify-between gap-8">
          <div>
            <h3 className="text-lg font-bold gradient-text">AI Token</h3>
            <p className="mt-2 text-sm text-muted-foreground max-w-xs leading-relaxed">
              面向开发者的 AI 模型聚合中转平台。一个 API，接入所有主流模型。
            </p>
          </div>
          <div className="grid grid-cols-2 gap-12 text-sm">
            <div>
              <h4 className="font-medium text-foreground">产品</h4>
              <ul className="mt-3 space-y-2 text-muted-foreground">
                <li><a href="/models" className="hover:text-foreground transition-colors">模型广场</a></li>
                <li><a href="/pricing" className="hover:text-foreground transition-colors">价格</a></li>
                <li><a href="/docs" className="hover:text-foreground transition-colors">API 文档</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-medium text-foreground">支持</h4>
              <ul className="mt-3 space-y-2 text-muted-foreground">
                <li><a href="/docs" className="hover:text-foreground transition-colors">开发文档</a></li>
                <li><a href="https://github.com/KeKe-Li/ai-token" className="hover:text-foreground transition-colors">GitHub</a></li>
              </ul>
            </div>
          </div>
        </div>
        <div className="mt-10 border-t border-border pt-8 text-center text-sm text-muted-foreground">
          &copy; 2026 AI Token. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
