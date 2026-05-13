"use client";

import { useEffect, useState, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  getPayableForReferrer,
  createPayout,
  listPayouts,
  markPayoutSent,
  type PayableRecord,
  type PayoutRow,
} from "@/services/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";

const fmt = (n: number) => new Intl.NumberFormat("vi-VN").format(n) + " đ";

export default function PayoutsPage() {
  const search = useSearchParams();
  const referrerId = search.get("referrer") || "";
  const { user } = useAuth();
  const canManage = user?.role === "owner" || user?.role === "admin";

  const [payable, setPayable] = useState<PayableRecord[]>([]);
  const [totalPayable, setTotalPayable] = useState(0);
  const [payouts, setPayouts] = useState<PayoutRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [method, setMethod] = useState("bank");
  const [bankAcc, setBankAcc] = useState("");
  const [bankName, setBankName] = useState("");
  const [holderName, setHolderName] = useState("");
  const [note, setNote] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [pay, list] = await Promise.all([
        referrerId ? getPayableForReferrer(referrerId) : Promise.resolve({ data: [], total_amount: 0, count: 0 }),
        listPayouts(),
      ]);
      setPayable(pay.data ?? []);
      setTotalPayable(pay.total_amount ?? 0);
      setSelected(new Set((pay.data ?? []).map((r) => r.id)));
      setPayouts(list.data ?? []);
    } catch {
      toast.error("Không tải được dữ liệu");
    } finally {
      setLoading(false);
    }
  }, [referrerId]);

  useEffect(() => {
    load();
  }, [load]);

  const selectedTotal = payable
    .filter((r) => selected.has(r.id))
    .reduce((sum, r) => sum + r.commission_amount, 0);

  async function submit() {
    if (!referrerId || selected.size === 0) return;
    const body: Parameters<typeof createPayout>[0] = {
      referrer_user_id: referrerId,
      record_ids: Array.from(selected),
      method,
      note: note || undefined,
    };
    if (method === "bank") {
      if (!bankAcc || !bankName || !holderName) {
        toast.error("Vui lòng nhập thông tin ngân hàng");
        return;
      }
      body.bank_info = { account_no: bankAcc, bank_name: bankName, holder_name: holderName };
    }
    try {
      await createPayout(body);
      toast.success("Đã tạo payout. Trạng thái: chờ chuyển khoản.");
      setSelected(new Set());
      setNote("");
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Tạo payout thất bại");
    }
  }

  async function markSent(id: string) {
    try {
      await markPayoutSent(id);
      toast.success("Đã đánh dấu đã chuyển");
      load();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Cập nhật thất bại");
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <Link
          href="/referrals"
          className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-primary"
        >
          <ArrowLeft className="h-4 w-4" /> Quay lại
        </Link>
      </div>
      <h1 className="text-2xl font-bold">Quản lý payout</h1>

      {canManage && referrerId ? (
        <Card className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="font-semibold">Tạo payout cho đối tác</h2>
              <p className="text-xs text-gray-500">Referrer ID: {referrerId}</p>
            </div>
            <div className="text-right">
              <div className="text-xs text-gray-500">Tổng có thể trả</div>
              <div className="text-xl font-bold text-green-700">{fmt(totalPayable)}</div>
            </div>
          </div>

          {loading ? (
            <div className="text-center text-gray-500 py-4">Đang tải…</div>
          ) : payable.length === 0 ? (
            <div className="text-center text-gray-500 py-4">Không có hoa hồng nào có thể trả.</div>
          ) : (
            <div className="border rounded overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left">
                  <tr>
                    <th className="px-3 py-2 w-8">
                      <input
                        type="checkbox"
                        checked={selected.size === payable.length}
                        onChange={(e) =>
                          setSelected(e.target.checked ? new Set(payable.map((r) => r.id)) : new Set())
                        }
                      />
                    </th>
                    <th className="px-3 py-2 font-semibold">Ngày tạo</th>
                    <th className="px-3 py-2 font-semibold">Giai đoạn</th>
                    <th className="px-3 py-2 font-semibold text-right">Số tiền</th>
                  </tr>
                </thead>
                <tbody>
                  {payable.map((r) => (
                    <tr key={r.id} className="border-t hover:bg-gray-50">
                      <td className="px-3 py-2">
                        <input
                          type="checkbox"
                          checked={selected.has(r.id)}
                          onChange={(e) => {
                            const next = new Set(selected);
                            if (e.target.checked) next.add(r.id);
                            else next.delete(r.id);
                            setSelected(next);
                          }}
                        />
                      </td>
                      <td className="px-3 py-2">{new Date(r.created_at).toLocaleDateString("vi-VN")}</td>
                      <td className="px-3 py-2">
                        GD {r.stage} ({(r.rate * 100).toFixed(0)}%)
                      </td>
                      <td className="px-3 py-2 text-right font-medium">{fmt(r.commission_amount)}</td>
                    </tr>
                  ))}
                </tbody>
                <tfoot className="bg-gray-50 border-t">
                  <tr>
                    <td colSpan={3} className="px-3 py-2 font-semibold">Tổng đã chọn ({selected.size}):</td>
                    <td className="px-3 py-2 text-right font-bold text-green-700">{fmt(selectedTotal)}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          )}

          {selected.size > 0 && (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium block mb-1">Phương thức</label>
                <select value={method} onChange={(e) => setMethod(e.target.value)} className="w-full border rounded px-2 py-1.5 text-sm">
                  <option value="bank">Chuyển khoản ngân hàng</option>
                  <option value="momo">MoMo</option>
                  <option value="cash">Tiền mặt</option>
                  <option value="other">Khác</option>
                </select>
              </div>
              <div>
                <label className="text-sm font-medium block mb-1">Ghi chú</label>
                <Input value={note} onChange={(e) => setNote(e.target.value)} />
              </div>
              {method === "bank" && (
                <>
                  <Input placeholder="Số tài khoản" value={bankAcc} onChange={(e) => setBankAcc(e.target.value)} />
                  <Input placeholder="Tên ngân hàng" value={bankName} onChange={(e) => setBankName(e.target.value)} />
                  <Input placeholder="Chủ tài khoản" value={holderName} onChange={(e) => setHolderName(e.target.value)} />
                </>
              )}
              <div className="col-span-2 flex justify-end">
                <Button onClick={submit}>Tạo payout {fmt(selectedTotal)}</Button>
              </div>
            </div>
          )}
        </Card>
      ) : canManage ? (
        <Card className="p-4">
          <p className="text-sm text-gray-600">Chọn 1 đối tác từ trang <a href="/referrals" className="text-primary underline">Hoa hồng giới thiệu</a> để tạo payout mới.</p>
        </Card>
      ) : null}

      <Card className="overflow-hidden">
        <div className="p-4 border-b font-semibold">Lịch sử payout</div>
        {payouts.length === 0 ? (
          <div className="p-6 text-center text-gray-500">Chưa có payout nào.</div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b text-left">
              <tr>
                <th className="px-4 py-2 font-semibold">Ngày tạo</th>
                <th className="px-4 py-2 font-semibold">Đối tác</th>
                <th className="px-4 py-2 font-semibold text-right">Số tiền</th>
                <th className="px-4 py-2 font-semibold text-right">Số record</th>
                <th className="px-4 py-2 font-semibold">Phương thức</th>
                <th className="px-4 py-2 font-semibold">Trạng thái</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {payouts.map((p) => (
                <tr key={p.id} className="border-b hover:bg-gray-50">
                  <td className="px-4 py-2">{new Date(p.created_at).toLocaleDateString("vi-VN")}</td>
                  <td className="px-4 py-2 font-mono text-xs">{p.referrer_user_id.slice(0, 8)}…</td>
                  <td className="px-4 py-2 text-right font-semibold">{fmt(p.total_amount)}</td>
                  <td className="px-4 py-2 text-right">{p.record_count}</td>
                  <td className="px-4 py-2">{p.method}</td>
                  <td className="px-4 py-2">
                    <span className={p.status === "sent" ? "text-blue-700" : "text-orange-600"}>
                      {p.status === "sent" ? "Đã chuyển" : "Chờ chuyển"}
                    </span>
                  </td>
                  <td className="px-4 py-2">
                    {canManage && p.status === "pending" && (
                      <Button size="sm" variant="outline" onClick={() => markSent(p.id)}>
                        Đánh dấu đã chuyển
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  );
}
