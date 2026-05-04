"use client";

import { useState } from "react";
import { useAuthGuard } from "@/lib/auth-guard";

const PLANS = [
  { amount: 10, bonus: 0, label: "¥10" },
  { amount: 50, bonus: 5, label: "¥50 +¥5" },
  { amount: 100, bonus: 15, label: "¥100 +¥15" },
  { amount: 500, bonus: 100, label: "¥500 +¥100" },
  { amount: 1000, bonus: 250, label: "¥1000 +¥250" },
];

export default function TopupPage() {
  const { checked } = useAuthGuard(true);
  const [selected, setSelected] = useState(100);
  const [customAmount, setCustomAmount] = useState("");
  const [useCustom, setUseCustom] = useState(false);

  if (!checked) return null;

  const finalAmount = useCustom ? parseInt(customAmount) || 0 : selected;
  const plan = PLANS.find((p) => p.amount === finalAmount);
  const bonus = plan?.bonus || 0;

  return (
    <div>
      <h1 className="text-2xl font-bold">充值</h1>
      <p className="mt-1 text-sm text-muted-foreground">为账户充值余额，用于 API 调用</p>

      <div className="mt-8">
        <h2 className="text-lg font-semibold">选择金额</h2>
        <div className="mt-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
          {PLANS.map((plan) => (
            <button
              key={plan.amount}
              onClick={() => { setSelected(plan.amount); setUseCustom(false); }}
              className={`rounded-xl border p-4 text-center transition-all ${
                !useCustom && selected === plan.amount
                  ? "border-primary bg-primary/10"
                  : "border-border hover:border-primary/50"
              }`}
            >
              <div className="text-lg font-bold">¥{plan.amount}</div>
              {plan.bonus > 0 && (
                <div className="mt-1 text-xs text-emerald-400">赠送 ¥{plan.bonus}</div>
              )}
            </button>
          ))}
        </div>

        <div className="mt-4">
          <button
            onClick={() => setUseCustom(true)}
            className={`text-sm ${useCustom ? "text-primary" : "text-muted-foreground hover:text-foreground"}`}
          >
            自定义金额
          </button>
          {useCustom && (
            <div className="mt-2 flex items-center gap-2">
              <span className="text-lg">¥</span>
              <input
                type="number"
                value={customAmount}
                onChange={(e) => setCustomAmount(e.target.value)}
                placeholder="输入金额"
                min={1}
                className="w-40 rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
              />
            </div>
          )}
        </div>
      </div>

      <div className="mt-8 card-glow rounded-xl p-6">
        <h2 className="text-lg font-semibold">支付方式</h2>
        <div className="mt-4 grid grid-cols-2 md:grid-cols-3 gap-3">
          <PaymentOption label="支付宝" icon="💳" disabled />
          <PaymentOption label="微信支付" icon="💚" disabled />
          <PaymentOption label="Stripe" icon="💎" disabled />
        </div>
        <p className="mt-4 text-xs text-muted-foreground">
          支付功能即将上线。当前可联系管理员手动充值。
        </p>
      </div>

      <div className="mt-6 card-glow rounded-xl p-6 flex items-center justify-between">
        <div>
          <div className="text-sm text-muted-foreground">充值金额</div>
          <div className="text-2xl font-bold">¥{finalAmount}{bonus > 0 && <span className="text-sm text-emerald-400 ml-2">+¥{bonus} 赠送</span>}</div>
        </div>
        <button
          disabled
          className="btn-glow rounded-xl px-8 py-3 text-sm font-medium text-white opacity-50 cursor-not-allowed"
        >
          立即支付
        </button>
      </div>
    </div>
  );
}

function PaymentOption({ label, icon, disabled }: { label: string; icon: string; disabled?: boolean }) {
  return (
    <div className={`rounded-xl border border-border p-4 text-center ${disabled ? "opacity-50" : "hover:border-primary/50 cursor-pointer"}`}>
      <div className="text-2xl">{icon}</div>
      <div className="mt-1 text-sm">{label}</div>
      {disabled && <div className="mt-1 text-[10px] text-muted-foreground">即将上线</div>}
    </div>
  );
}
