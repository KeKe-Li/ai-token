-- 补充新模型: GPT-5.4 Codex, DeepSeek V4 系列
INSERT INTO models (model_id, display_name, provider, category, context_length, input_price, output_price, price_unit, capabilities, description) VALUES
('gpt-5.4-codex', 'GPT-5.4 Codex', 'openai', 'llm', 1000000, 2500, 10000, '1M', ARRAY['function_call','streaming','thinking'], 'GPT-5.4 Codex 编程专用模型'),
('deepseek/deepseek-v4-flash', 'DeepSeek V4 Flash', 'deepseek', 'llm', 128000, 200, 800, '1M', ARRAY['function_call','streaming','json_mode'], 'DeepSeek V4 Flash 快速模型'),
('deepseek/deepseek-v4-pro', 'DeepSeek V4 Pro', 'deepseek', 'llm', 128000, 1000, 4000, '1M', ARRAY['function_call','streaming','thinking','json_mode'], 'DeepSeek V4 Pro 旗舰模型')
ON CONFLICT (model_id) DO NOTHING;
