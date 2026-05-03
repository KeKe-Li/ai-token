"use client";

import { useState, useRef, useEffect } from "react";

type Message = { role: "user" | "assistant" | "system"; content: string };

const MODELS = [
  "gpt-4o", "gpt-4o-mini",
  "claude-sonnet-4-6", "claude-haiku-4-5",
  "gemini-2.5-pro", "gemini-2.5-flash",
  "deepseek-chat", "deepseek-reasoner",
];

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function PlaygroundPage() {
  const [model, setModel] = useState("gpt-4o-mini");
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);
  const [systemPrompt, setSystemPrompt] = useState("");
  const [showSystem, setShowSystem] = useState(false);
  const [temperature, setTemperature] = useState(0.7);
  const [maxTokens, setMaxTokens] = useState(2048);
  const [totalTokens, setTotalTokens] = useState(0);
  const scrollRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: "smooth" });
  }, [messages]);

  async function handleSend() {
    if (!input.trim() || streaming) return;

    const apiKey = getApiKey();
    if (!apiKey) {
      alert("请先在 API Keys 页面创建密钥");
      return;
    }

    const userMsg: Message = { role: "user", content: input.trim() };
    const allMessages = [...messages, userMsg];
    setMessages(allMessages);
    setInput("");
    setStreaming(true);

    const requestMessages = systemPrompt
      ? [{ role: "system" as const, content: systemPrompt }, ...allMessages]
      : allMessages;

    const assistantMsg: Message = { role: "assistant", content: "" };
    setMessages([...allMessages, assistantMsg]);

    try {
      abortRef.current = new AbortController();
      const res = await fetch(`${API_BASE}/v1/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${apiKey}`,
        },
        body: JSON.stringify({
          model,
          messages: requestMessages.map((m) => ({ role: m.role, content: m.content })),
          stream: true,
          temperature,
          max_tokens: maxTokens,
        }),
        signal: abortRef.current.signal,
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: { message: `HTTP ${res.status}` } }));
        setMessages([...allMessages, { role: "assistant", content: `错误: ${err.error?.message || res.statusText}` }]);
        setStreaming(false);
        return;
      }

      const reader = res.body?.getReader();
      const decoder = new TextDecoder();
      let content = "";
      let tokens = 0;

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split("\n").filter((l) => l.startsWith("data: "));

        for (const line of lines) {
          const data = line.slice(6);
          if (data === "[DONE]") continue;
          try {
            const parsed = JSON.parse(data);
            const delta = parsed.choices?.[0]?.delta?.content;
            if (delta) {
              content += delta;
              tokens++;
              setMessages([...allMessages, { role: "assistant", content }]);
            }
          } catch { /* ignore parse errors */ }
        }
      }

      setTotalTokens((prev) => prev + tokens);
    } catch (err) {
      if (err instanceof Error && err.name === "AbortError") {
        // user cancelled
      } else {
        setMessages([...allMessages, { role: "assistant", content: `请求失败: ${err}` }]);
      }
    } finally {
      setStreaming(false);
      abortRef.current = null;
    }
  }

  function handleStop() {
    abortRef.current?.abort();
    setStreaming(false);
  }

  function handleClear() {
    setMessages([]);
    setTotalTokens(0);
  }

  function getApiKey(): string | null {
    return localStorage.getItem("playground_key") || promptForKey();
  }

  function promptForKey(): string | null {
    const key = window.prompt("请输入你的 API Key（sk-开头）：");
    if (key && key.startsWith("sk-")) {
      localStorage.setItem("playground_key", key);
      return key;
    }
    return null;
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  return (
    <div className="flex flex-col h-[calc(100vh-12rem)]">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-2xl font-bold">Playground</h1>
          <p className="mt-1 text-sm text-muted-foreground">在浏览器中直接测试 AI 模型</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-muted-foreground">{totalTokens} tokens</span>
          <button onClick={handleClear} className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-muted transition-colors">
            清空对话
          </button>
        </div>
      </div>

      {/* 配置栏 */}
      <div className="flex flex-wrap gap-3 mb-4 p-3 rounded-xl border border-border bg-card/50">
        <div>
          <label className="block text-xs text-muted-foreground mb-1">模型</label>
          <select value={model} onChange={(e) => setModel(e.target.value)}
            className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm">
            {MODELS.map((m) => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-xs text-muted-foreground mb-1">Temperature: {temperature}</label>
          <input type="range" min="0" max="2" step="0.1" value={temperature}
            onChange={(e) => setTemperature(parseFloat(e.target.value))}
            className="w-32 accent-primary" />
        </div>
        <div>
          <label className="block text-xs text-muted-foreground mb-1">Max Tokens</label>
          <input type="number" value={maxTokens} onChange={(e) => setMaxTokens(parseInt(e.target.value) || 2048)}
            className="w-24 rounded-lg border border-border bg-background px-2 py-1.5 text-sm" />
        </div>
        <div className="flex items-end">
          <button onClick={() => setShowSystem(!showSystem)}
            className={`rounded-lg border px-3 py-1.5 text-xs transition-colors ${showSystem ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:text-foreground"}`}>
            System Prompt
          </button>
        </div>
      </div>

      {showSystem && (
        <div className="mb-4">
          <textarea value={systemPrompt} onChange={(e) => setSystemPrompt(e.target.value)}
            placeholder="设置系统提示词..."
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm resize-none h-20 focus:border-primary focus:outline-none" />
        </div>
      )}

      {/* 消息区域 */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto rounded-xl border border-border bg-card/30 p-4 space-y-4 min-h-0">
        {messages.length === 0 && (
          <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
            选择模型，输入消息开始对话
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            <div className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed ${
              msg.role === "user"
                ? "bg-primary text-primary-foreground"
                : "bg-muted text-foreground"
            }`}>
              <pre className="whitespace-pre-wrap font-sans">{msg.content || (streaming && i === messages.length - 1 ? "●" : "")}</pre>
            </div>
          </div>
        ))}
      </div>

      {/* 输入区域 */}
      <div className="mt-4 flex gap-2">
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"
          rows={2}
          className="flex-1 rounded-xl border border-border bg-background px-4 py-3 text-sm resize-none focus:border-primary focus:outline-none"
        />
        {streaming ? (
          <button onClick={handleStop} className="self-end rounded-xl bg-destructive px-6 py-3 text-sm font-medium text-white">
            停止
          </button>
        ) : (
          <button onClick={handleSend} disabled={!input.trim()}
            className="self-end btn-glow rounded-xl px-6 py-3 text-sm font-medium text-white disabled:opacity-50">
            发送
          </button>
        )}
      </div>
    </div>
  );
}
