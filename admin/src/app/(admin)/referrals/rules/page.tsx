"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { ArrowLeft, Plus, Trash2, Pencil } from "lucide-react";
import {
  listCommissionRules,
  upsertCommissionRule,
  deleteCommissionRule,
  getLeaderboard,
  type CommissionRule,
  type LeaderboardRow,
} from "@/services/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";

const fmt = (n: number) => new Intl.NumberFormat("vi-VN").format(n);

// stage1_days/stage2_days không còn ý nghĩa (mô hình payment-count) nhưng backend
// vẫn validate min=1 → gửi giá trị cố định, không hiển thị.
const FIXED_DAYS = { stage1_days: 90, stage2_days: 180 };

type RuleForm = {
  stage1_pct: number;
  stage2_pct: number;
  stage3_pct: number;
  stage3_cap_months: number;
  base_type: "net" | "gross";
  minimum_payout: number;
};

const DEFAULT_FORM: RuleForm = {
  stage1_pct: 0.5,
  stage2_pct: 0.3,
  stage3_pct: 0.2,
  stage3_cap_months: 24,
  base_type: "net",
  minimum_payout: 100000,
};

function ruleToForm(r: CommissionRule): RuleForm {
  return {
    stage1_pct: r.stage1_pct,
    stage2_pct: r.stage2_pct,
    stage3_pct: r.stage3_pct,
    stage3_cap_months: r.stage3_cap_months,
    base_type: r.base_type,
    minimum_payout: r.minimum_payout,
  };
}

export default function CommissionRulesPage() {
  const { user } = useAuth();
  const canManage = user?.role === "owner" || user?.role === "admin";
  const [rules, setRules] = useState<CommissionRule[]>([]);
  const [partners, setPartners] = useState<LeaderboardRow[]>([]);
  const [loading, setLoading] = useState(true);

  // Default-rule editing
  const [editingDefault, setEditingDefault] = useState(false);
  const [form, setForm] = useState<RuleForm>(DEFAULT_FORM);

  // Per-partner add/edit
  const [pOpen, setPOpen] = useState(false);
  const [pCodeId, setPCodeId] = useState<string>(""); // referral_code_id đang chỉnh
  const [pForm, setPForm] = useState<RuleForm>(DEFAULT_FORM);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [r, lb] = await Promise.all([listCommissionRules(), getLeaderboard()]);
      setRules((r.data ?? []).filter((x) => !x.active_to));
      setPartners(lb.data ?? []);
    } catch {
      toast.error("Không tải được dữ liệu quy tắc");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const defaultRule = rules.find((r) => !r.referral_code_id);
  const overrides = rules.filter((r) => r.referral_code_id);
  const partnerByCode = new Map(partners.filter((p) => p.referral_code_id).map((p) => [p.referral_code_id as string, p]));
  const availablePartners = partners.filter(
    (p) => p.referral_code_id && !overrides.some((o) => o.referral_code_id === p.referral_code_id),
  );

  function startEditDefault() {
    if (defaultRule) setForm(ruleToForm(defaultRule));
    else setForm(DEFAULT_FORM);
    setEditingDefault(true);
  }

  async function saveDefault() {
    try {
      await upsertCommissionRule({ referral_code_id: null, ...FIXED_DAYS, ...form });
      toast.success("Đã lưu quy tắc mặc định");
      setEditingDefault(false);
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Lưu thất bại");
    }
  }

  function openAddPartner() {
    setPCodeId("");
    setPForm(defaultRule ? ruleToForm(defaultRule) : DEFAULT_FORM);
    setPOpen(true);
  }

  function openEditPartner(rule: CommissionRule) {
    setPCodeId(rule.referral_code_id as string);
    setPForm(ruleToForm(rule));
    setPOpen(true);
  }

  async function savePartner() {
    if (!pCodeId) {
      toast.error("Hãy chọn đối tác");
      return;
    }
    try {
      await upsertCommissionRule({ referral_code_id: pCodeId, ...FIXED_DAYS, ...pForm });
      toast.success("Đã lưu quy tắc riêng cho đối tác");
      setPOpen(false);
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Lưu thất bại");
    }
  }

  async function removeOverride(rule: CommissionRule) {
    const p = partnerByCode.get(rule.referral_code_id as string);
    if (!confirm(`Gỡ quy tắc riêng của ${p?.name || p?.phone || "đối tác này"}? Đối tác sẽ quay về quy tắc mặc định.`)) return;
    try {
      await deleteCommissionRule(rule.referral_code_id as string);
      toast.success("Đã gỡ về quy tắc mặc định");
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gỡ thất bại");
    }
  }

  return (
    <div className="space-y-5 max-w-3xl">
      <Link href="/referrals" className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-primary">
        <ArrowLeft className="h-4 w-4" /> Quay lại
      </Link>
      <h1 className="text-2xl font-bold">Cài đặt quy tắc hoa hồng</h1>

      {loading ? (
        <div className="text-center text-gray-500 py-8">Đang tải…</div>
      ) : (
        <>
          {/* ── Quy tắc mặc định ── */}
          <Card className="p-4 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-semibold">Quy tắc mặc định</h2>
                <p className="text-xs text-gray-500">Áp dụng cho mọi đối tác chưa có quy tắc riêng.</p>
              </div>
              {canManage && !editingDefault && (
                <Button size="sm" onClick={startEditDefault}>Chỉnh sửa</Button>
              )}
            </div>

            {editingDefault ? (
              <RuleEditor form={form} setForm={setForm} onSave={saveDefault} onCancel={() => setEditingDefault(false)} />
            ) : defaultRule ? (
              <RuleView rule={defaultRule} />
            ) : (
              <div className="text-center text-gray-500 py-4">Chưa có quy tắc mặc định.</div>
            )}
          </Card>

          {/* ── Quy tắc riêng từng đối tác ── */}
          <Card className="p-4 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-semibold">Quy tắc riêng từng đối tác</h2>
                <p className="text-xs text-gray-500">Đối tác đặc biệt (VIP) có thể hưởng % hoặc thời hạn khác mặc định.</p>
              </div>
              {canManage && !pOpen && (
                <Button size="sm" variant="outline" onClick={openAddPartner}>
                  <Plus className="h-4 w-4 mr-1" /> Thêm
                </Button>
              )}
            </div>

            {pOpen && (
              <div className="border rounded-lg p-3 bg-muted/30 space-y-3">
                <div>
                  <label className="text-sm font-medium block mb-1">Đối tác</label>
                  {pCodeId && partnerByCode.get(pCodeId) ? (
                    <div className="text-sm py-1.5">
                      <strong>{partnerByCode.get(pCodeId)?.name || "(chưa đặt tên)"}</strong>{" "}
                      <span className="text-gray-500">{partnerByCode.get(pCodeId)?.phone} · mã {partnerByCode.get(pCodeId)?.code}</span>
                    </div>
                  ) : (
                    <select
                      value={pCodeId}
                      onChange={(e) => setPCodeId(e.target.value)}
                      className="w-full border rounded px-2 py-1.5 text-sm"
                    >
                      <option value="">— Chọn đối tác —</option>
                      {availablePartners.map((p) => (
                        <option key={p.referral_code_id} value={p.referral_code_id as string}>
                          {(p.name || "(chưa đặt tên)") + " · " + p.phone + " · mã " + p.code}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
                <RuleEditor form={pForm} setForm={setPForm} onSave={savePartner} onCancel={() => setPOpen(false)} />
              </div>
            )}

            {overrides.length === 0 && !pOpen ? (
              <div className="text-center text-gray-400 py-4 text-sm">Chưa có quy tắc riêng nào.</div>
            ) : (
              <div className="space-y-2">
                {overrides.map((o) => {
                  const p = partnerByCode.get(o.referral_code_id as string);
                  return (
                    <div key={o.id} className="flex items-center justify-between border rounded-lg p-3 text-sm">
                      <div>
                        <div className="font-medium">{p?.name || "(không rõ)"} <span className="text-gray-500 font-normal">{p?.phone}</span></div>
                        <div className="text-gray-600 text-xs mt-0.5">
                          L1 {(o.stage1_pct * 100).toFixed(0)}% · L2 {(o.stage2_pct * 100).toFixed(0)}% · L3 {(o.stage3_pct * 100).toFixed(0)}%
                          {" · "}
                          {o.stage3_cap_months > 0 ? `lần 3 trong ${o.stage3_cap_months} tháng` : "lần 3 vĩnh viễn"}
                          {" · "}ngưỡng {fmt(o.minimum_payout)}đ
                        </div>
                      </div>
                      {canManage && (
                        <div className="flex gap-1">
                          <Button size="sm" variant="ghost" className="h-8 w-8 p-0" onClick={() => openEditPartner(o)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button size="sm" variant="ghost" className="h-8 w-8 p-0 text-destructive hover:text-destructive" onClick={() => removeOverride(o)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
}

// ── Form chỉnh sửa 1 quy tắc (dùng chung cho mặc định + per-partner) ──
function RuleEditor({
  form, setForm, onSave, onCancel,
}: {
  form: RuleForm;
  setForm: (f: RuleForm) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-3">
        <PctField label="Lần 1" value={form.stage1_pct} onChange={(v) => setForm({ ...form, stage1_pct: v })} />
        <PctField label="Lần 2" value={form.stage2_pct} onChange={(v) => setForm({ ...form, stage2_pct: v })} />
        <PctField label="Lần 3 trở đi" value={form.stage3_pct} onChange={(v) => setForm({ ...form, stage3_pct: v })} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="text-sm font-medium block mb-1">Giới hạn hoa hồng lần 3 (tháng)</label>
          <Input
            type="number"
            min={0}
            value={form.stage3_cap_months}
            onChange={(e) => setForm({ ...form, stage3_cap_months: Number(e.target.value) })}
          />
          <div className="text-[11px] text-gray-500 mt-1">
            Tính từ ngày TV được giới thiệu đăng ký. Quá hạn → ngừng trả hoa hồng. <strong>0 = vĩnh viễn.</strong>
          </div>
        </div>
        <div>
          <label className="text-sm font-medium block mb-1">Ngưỡng rút tối thiểu (VND)</label>
          <Input
            type="number"
            value={form.minimum_payout}
            onChange={(e) => setForm({ ...form, minimum_payout: Number(e.target.value) })}
          />
        </div>
      </div>

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

      <div className="flex gap-2">
        <Button onClick={onSave}>Lưu</Button>
        <Button variant="outline" onClick={onCancel}>Hủy</Button>
      </div>
    </div>
  );
}

function PctField({ label, value, onChange }: { label: string; value: number; onChange: (v: number) => void }) {
  return (
    <div className="border rounded p-3 space-y-2">
      <div className="text-xs font-semibold text-gray-600">{label}</div>
      <Input type="number" step="0.01" min={0} max={1} value={value} onChange={(e) => onChange(Number(e.target.value))} />
      <div className="text-[10px] text-gray-400">Vd: 0.5 = 50%</div>
    </div>
  );
}

function RuleView({ rule }: { rule: CommissionRule }) {
  return (
    <div className="grid grid-cols-3 gap-4 text-sm">
      <Display label="Lần 1" pct={rule.stage1_pct} />
      <Display label="Lần 2" pct={rule.stage2_pct} />
      <Display label="Lần 3 trở đi" pct={rule.stage3_pct} />
      <div className="col-span-3 border-t pt-3 mt-1 text-gray-600">
        Giới hạn lần 3: <strong>{rule.stage3_cap_months > 0 ? `${rule.stage3_cap_months} tháng kể từ khi giới thiệu` : "vĩnh viễn"}</strong>
        {" · "}
        Cơ sở tính: <strong>{rule.base_type === "net" ? "Ròng" : "Gộp"}</strong>
        {" · "}
        Ngưỡng rút: <strong>{fmt(rule.minimum_payout)} đ</strong>
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
