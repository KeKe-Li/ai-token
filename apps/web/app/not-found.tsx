export default function NotFound() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center px-6">
      <div className="text-center">
        <div className="text-6xl font-bold gradient-text">404</div>
        <p className="mt-4 text-lg text-muted-foreground">页面未找到</p>
        <a href="/" className="mt-6 inline-block btn-glow rounded-lg px-6 py-2.5 text-sm font-medium text-white">
          返回首页
        </a>
      </div>
    </div>
  );
}
