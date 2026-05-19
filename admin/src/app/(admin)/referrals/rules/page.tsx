"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { listCommissionRules, upsertCommissionRule, type CommissionRule } from "@/services/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";

const fmt = (n: number) => new Intl.NumberFormat("vi-VN").format(n);

export default function CommissionRulesPage() {
  const { user } = useAuth();
  const canManage = user?.role === "owner" || user?.role === "admin";
  const [rules, setRules] = useState<CommissionRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({
    stage1_days: 90,
    stage1_pct: 0.5,
    stage2_days: 180,
    stage2_pct: 0.3,
    stage3_pct: 0.2,
    base_type: "net" as "net" | "gross",
    minimum_payout: 100000,
  });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const r = await listCommissionRules();
      const active = (r.data ?? []).filter((x) => !x.active_to);
      setRules(active);
    } catch {
      toast.error("Không tải được rules");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const defaultRule = rules.find((r) => !r.referral_code_id);

  function startEditDefault() {
    if (defaultRule) {
      setForm({
        stage1_days: defaultRule.stage1_days,
        stage1_pct: defaultRule.stage1_pct,
        stage2_days: defaultRule.stage2_days,
        stage2_pct: defaultRule.stage2_pct,
        stage3_pct: defaultRule.stage3_pct,
        base_type: defaultRule.base_type,
        minimum_payout: defaultRule.minimum_payout,
      });
    }
    setEditingId("default");
  }

  async function save() {
    try {
      await upsertCommissionRule({
        referral_code_id: null,
        ...form,
      });
      toast.success("Đã lưu rule mặc định");
      setEditingId(null);
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Lưu thất bại");
    }
  }

  return (
    <div className="space-y-4 max-w-3xl">
      <Link
        href="/referrals"
        className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" /> Quay lại
      </Link>
      <h1 className="text-2xl font-bold">Cài đặt quy tắc hoa hồng</h1>
      <p className="text-sm text-gray-600">
        Quy tắc mặc định áp dụng cho mọi đối tác chưa có thỏa thuận riêng. Per-partner override có thể tạo sau (chưa hỗ trợ UI).
      </p>

      {loading ? (
        <div className="text-center text-gray-500 py-8">Đang tải…</div>
      ) : (
        <Card className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold">Quy tắc mặc định</h2>
            {canManage && editingId !== "default" && (
              <Button size="sm" onClick={startEditDefault}>
                Chỉnh sửa
              </Button>
            )}
          </div>

          {editingId === "default" ? (
            <div className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <Stage label="Lần thanh toán đầu tiên" days={form.stage1_days} pct={form.stage1_pct}
                  onDays={(v) => setForm({ ...form, stage1_days: v })}
                  onPct={(v) => setForm({ ...form, stage1_pct: v })}
                />
                <Stage label="Lần thanh toán thứ 2" days={form.stage2_days} pct={form.stage2_pct}
                  onDays={(v) => setForm({ ...form, stage2_days: v })}
                  onPct={(v) => setForm({ ...form, stage2_pct: v })}
                />
                <Stage label="Lần thứ 3 trở đi (vĩnh viễn)" days={null} pct={form.stage3_pct}
                  onDays={() => {}}
                  onPct={(v) => setForm({ ...form, stage3_pct: v })}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium block mb-1">Cơ sở tính</label>
                  <select
                    value={form.base_type}
                    onChange={(e) => setForm({ ...form, base_type: e.target.value as "net" | "gross" })}
                    className="w-full border rounded px-2 py-1.5 text-sm"
                  >
                    <option value="net">Doanh thu ròng (sau phí nền tảng)</option>
                    <option value="gross">Doanh thu gộp</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium block mb-1">Ngưỡng payout tối thiểu (VND)</label>
                  <Input
                    type="number"
                    value={form.minimum_payout}
                    onChange={(e) => setForm({ ...form, minimum_payout: Number(e.target.value) })}
                  />
                </div>
              </div>

              <div className="flex gap-2">
                <Button onClick={save}>Lưu</Button>
                <Button variant="outline" onClick={() => setEditingId(null)}>Hủy</Button>
              </div>
            </div>
          ) : defaultRule ? (
            <div className="grid grid-cols-3 gap-4 text-sm">
              <Display label="Lần thanh toán đầu tiên" pct={defaultRule.stage1_pct} />
              <Display label="Lần thanh toán thứ 2" pct={defaultRule.stage2_pct} />
              <Display label="Lần thứ 3 trở đi (vĩnh viễn)" pct={defaultRule.stage3_pct} />
              <div className="col-span-3 border-t pt-3 mt-1 text-gray-600">
                Cơ sở tính: <strong>{defaultRule.base_type === "net" ? "Ròng" : "Gộp"}</strong>
                {" · "}
                Ngưỡng payout: <strong>{fmt(defaultRule.minimum_payout)} đ</strong>
              </div>
            </div>
          ) : (
            <div className="text-center text-gray-500 py-4">Chưa có rule mặc định.</div>
          )}
        </Card>
      )}
    </div>
  );
}

// `days` field giữ trong payload backend (column stage1_days/stage2_days vẫn
// có trong DB) nhưng KHÔNG hiển thị/edit ở UI — hết ý nghĩa từ khi đổi sang
// mô hình payment-count. Engine bỏ qua giá trị này, dùng paymentSequence.
function Stage({
  label, days, pct, onDays, onPct,
}: {
  label: string;
  days: number | null;
  pct: number;
  onDays: (v: number) => void;
  onPct: (v: number) => void;
}) {
  void days; void onDays; // unused — payment-count model
  return (
    <div className="border rounded p-3 space-y-2">
      <div className="text-xs font-semibold text-gray-600">{label}</div>
      <div>
        <label className="text-xs block mb-1">% hoa hồng</label>
        <Input
          type="number"
          step="0.01"
          min={0}
          max={1}
          value={pct}
          onChange={(e) => onPct(Number(e.target.value))}
        />
        <div className="text-[10px] text-gray-400 mt-1">Vd: 0.45 = 45%</div>
      </div>
    </div>
  );
}

function Display({ label, pct }: { label: string; pct: number }) {
  return (
    <div className="border rounded p-3">
      <div className="text-xs font-semibold text-gray-600">{label}</div>
      <div className="mt-1">
        <span className="text-lg font-bold">{(pct * 100).toFixed(0)}%</span>
      </div>
    </div>
  );
}
