export type Model = {
  model_id: string;
  display_name: string;
  provider: string;
  category: string;
  context_length: number;
  input_price: number;
  output_price: number;
  price_unit: string;
  capabilities: string[];
  description: string;
};

export type PricingModel = Pick<
  Model,
  "model_id" | "display_name" | "provider" | "context_length" | "input_price" | "output_price" | "price_unit"
>;

// 供应商展示名映射：数据层统一使用小写标识，展示层通过此表还原规范大小写。
export const PROVIDER_LABELS: Record<string, string> = {
  openai: "OpenAI",
  anthropic: "Anthropic",
  google: "Google",
  deepseek: "DeepSeek",
};

export const FALLBACK_MODELS: Model[] = [
  // OpenAI GPT-5 系列
  { model_id: "gpt-5.5", display_name: "GPT-5.5", provider: "openai", category: "llm", context_length: 1000000, input_price: 2500, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "OpenAI 最新旗舰模型" },
  { model_id: "gpt-5.4", display_name: "GPT-5.4", provider: "openai", category: "llm", context_length: 1000000, input_price: 2500, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "OpenAI 旗舰模型" },
  { model_id: "gpt-5.4-codex", display_name: "GPT-5.4 Codex", provider: "openai", category: "llm", context_length: 1000000, input_price: 2500, output_price: 10000, price_unit: "1M", capabilities: ["function_call", "streaming", "thinking"], description: "GPT-5.4 Codex 编程专用模型" },
  { model_id: "gpt-5.3-codex", display_name: "GPT-5.3 Codex", provider: "openai", category: "llm", context_length: 256000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["function_call", "streaming", "thinking"], description: "OpenAI Codex 编程模型" },
  { model_id: "gpt-5.2", display_name: "GPT-5.2", provider: "openai", category: "llm", context_length: 256000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "OpenAI 高级模型" },
  { model_id: "gpt-5.1", display_name: "GPT-5.1", provider: "openai", category: "llm", context_length: 256000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "OpenAI 高级模型" },
  { model_id: "gpt-5", display_name: "GPT-5", provider: "openai", category: "llm", context_length: 256000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "OpenAI GPT-5" },
  { model_id: "gpt-5-codex", display_name: "GPT-5 Codex", provider: "openai", category: "llm", context_length: 256000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["function_call", "streaming", "thinking"], description: "GPT-5 Codex 编程专用" },
  { model_id: "gpt-5-mini", display_name: "GPT-5 Mini", provider: "openai", category: "llm", context_length: 256000, input_price: 300, output_price: 1200, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "GPT-5 轻量版" },
  { model_id: "gpt-5-nano", display_name: "GPT-5 Nano", provider: "openai", category: "llm", context_length: 128000, input_price: 100, output_price: 400, price_unit: "1M", capabilities: ["function_call", "streaming"], description: "GPT-5 超轻量版" },
  // OpenAI GPT-4 系列
  { model_id: "gpt-4o", display_name: "GPT-4o", provider: "openai", category: "llm", context_length: 128000, input_price: 2500, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "json_mode"], description: "GPT-4o 多模态模型" },
  { model_id: "gpt-4o-mini", display_name: "GPT-4o Mini", provider: "openai", category: "llm", context_length: 128000, input_price: 150, output_price: 600, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "json_mode"], description: "轻量级 GPT-4o" },
  { model_id: "gpt-4.1", display_name: "GPT-4.1", provider: "openai", category: "llm", context_length: 1000000, input_price: 2000, output_price: 8000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "GPT-4.1 长上下文" },
  { model_id: "gpt-4.1-mini", display_name: "GPT-4.1 Mini", provider: "openai", category: "llm", context_length: 1000000, input_price: 400, output_price: 1600, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "GPT-4.1 轻量版" },
  { model_id: "gpt-4.1-nano", display_name: "GPT-4.1 Nano", provider: "openai", category: "llm", context_length: 1000000, input_price: 100, output_price: 400, price_unit: "1M", capabilities: ["function_call", "streaming"], description: "GPT-4.1 超轻量版" },
  { model_id: "gpt-4", display_name: "GPT-4", provider: "openai", category: "llm", context_length: 128000, input_price: 30000, output_price: 60000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "经典 GPT-4" },
  { model_id: "gpt-4-turbo", display_name: "GPT-4 Turbo", provider: "openai", category: "llm", context_length: 128000, input_price: 10000, output_price: 30000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "json_mode"], description: "GPT-4 Turbo" },
  { model_id: "gpt-3.5-turbo", display_name: "GPT-3.5 Turbo", provider: "openai", category: "llm", context_length: 16385, input_price: 500, output_price: 1500, price_unit: "1M", capabilities: ["function_call", "streaming"], description: "经典 GPT-3.5" },
  // OpenAI 推理系列
  { model_id: "o3", display_name: "o3", provider: "openai", category: "llm", context_length: 200000, input_price: 10000, output_price: 40000, price_unit: "1M", capabilities: ["streaming", "thinking"], description: "OpenAI o3 推理模型" },
  { model_id: "o3-mini", display_name: "o3 Mini", provider: "openai", category: "llm", context_length: 200000, input_price: 1100, output_price: 4400, price_unit: "1M", capabilities: ["streaming", "thinking"], description: "o3 轻量推理" },
  { model_id: "o4-mini", display_name: "o4 Mini", provider: "openai", category: "llm", context_length: 200000, input_price: 1100, output_price: 4400, price_unit: "1M", capabilities: ["streaming", "thinking"], description: "o4 轻量推理" },
  // Anthropic Claude Opus 系列
  { model_id: "claude-opus-4-7", display_name: "Claude Opus 4.7", provider: "anthropic", category: "llm", context_length: 200000, input_price: 15000, output_price: 75000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Anthropic 最强旗舰模型" },
  { model_id: "claude-opus-4-6", display_name: "Claude Opus 4.6", provider: "anthropic", category: "llm", context_length: 200000, input_price: 15000, output_price: 75000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Opus 4.6" },
  { model_id: "claude-opus-4-5-20251101", display_name: "Claude Opus 4.5", provider: "anthropic", category: "llm", context_length: 200000, input_price: 15000, output_price: 75000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Opus 4.5" },
  { model_id: "claude-opus-4-20250514", display_name: "Claude Opus 4", provider: "anthropic", category: "llm", context_length: 200000, input_price: 15000, output_price: 75000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Opus 4" },
  { model_id: "claude-opus-4-1-20250805", display_name: "Claude Opus 4.1", provider: "anthropic", category: "llm", context_length: 200000, input_price: 15000, output_price: 75000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Opus 4.1" },
  // Anthropic Claude Sonnet 系列
  { model_id: "claude-sonnet-4-6", display_name: "Claude Sonnet 4.6", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Anthropic 最新编程模型" },
  { model_id: "claude-sonnet-4-5-20250929", display_name: "Claude Sonnet 4.5", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Sonnet 4.5" },
  { model_id: "claude-sonnet-4-20250514", display_name: "Claude Sonnet 4", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude Sonnet 4" },
  { model_id: "claude-3-7-sonnet-20250219", display_name: "Claude 3.7 Sonnet", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Claude 3.7 Sonnet" },
  { model_id: "claude-3-5-sonnet-20241022", display_name: "Claude 3.5 Sonnet", provider: "anthropic", category: "llm", context_length: 200000, input_price: 3000, output_price: 15000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Claude 3.5 Sonnet" },
  // Anthropic Claude Haiku 系列
  { model_id: "claude-haiku-4-5", display_name: "Claude Haiku 4.5", provider: "anthropic", category: "llm", context_length: 200000, input_price: 800, output_price: 4000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Anthropic 轻量快速模型" },
  { model_id: "claude-haiku-4-5-20251001", display_name: "Claude Haiku 4.5 (Oct)", provider: "anthropic", category: "llm", context_length: 200000, input_price: 800, output_price: 4000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Claude Haiku 4.5 稳定版" },
  { model_id: "claude-3-5-haiku-20241022", display_name: "Claude 3.5 Haiku", provider: "anthropic", category: "llm", context_length: 200000, input_price: 250, output_price: 1250, price_unit: "1M", capabilities: ["vision", "streaming"], description: "Claude 3.5 Haiku" },
  // DeepSeek
  { model_id: "deepseek-chat", display_name: "DeepSeek V3", provider: "deepseek", category: "llm", context_length: 64000, input_price: 270, output_price: 1100, price_unit: "1M", capabilities: ["function_call", "streaming", "json_mode"], description: "DeepSeek 通用对话模型" },
  { model_id: "deepseek-reasoner", display_name: "DeepSeek R1", provider: "deepseek", category: "llm", context_length: 64000, input_price: 550, output_price: 2190, price_unit: "1M", capabilities: ["streaming", "thinking"], description: "DeepSeek 深度推理模型" },
  { model_id: "deepseek/deepseek-v4-flash", display_name: "DeepSeek V4 Flash", provider: "deepseek", category: "llm", context_length: 128000, input_price: 200, output_price: 800, price_unit: "1M", capabilities: ["function_call", "streaming", "json_mode"], description: "DeepSeek V4 Flash 快速模型" },
  { model_id: "deepseek/deepseek-v4-pro", display_name: "DeepSeek V4 Pro", provider: "deepseek", category: "llm", context_length: 128000, input_price: 1000, output_price: 4000, price_unit: "1M", capabilities: ["function_call", "streaming", "thinking", "json_mode"], description: "DeepSeek V4 Pro 旗舰模型" },
  // Google
  { model_id: "gemini-2.5-pro", display_name: "Gemini 2.5 Pro", provider: "google", category: "llm", context_length: 1000000, input_price: 1250, output_price: 10000, price_unit: "1M", capabilities: ["vision", "function_call", "streaming", "thinking"], description: "Google 旗舰模型,100万上下文" },
  { model_id: "gemini-2.5-flash", display_name: "Gemini 2.5 Flash", provider: "google", category: "llm", context_length: 1000000, input_price: 150, output_price: 600, price_unit: "1M", capabilities: ["vision", "function_call", "streaming"], description: "Google 快速模型" },
];
