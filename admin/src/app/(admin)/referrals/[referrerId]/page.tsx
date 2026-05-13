"use client";

import { useEffect, useState, useCallback, use } from "react";
import Link from "next/link";
import { ArrowLeft, Users, CheckCircle2, Banknote } from "lucide-react";
import { getReferees, type RefereeRow } from "@/services/api";
import { Card } from "@/components/ui/card";
import { toast } from "sonner";

const fmtVND = (n: number) => new Intl.NumberFormat("vi-VN").format(n) + " đ";

const fmtDate = (iso: string) => {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleDateString("vi-VN");
  } catch {
    return iso;
  }
};

type Props = { params: Promise<{ referrerId: string }> };

export default function RefereesPage({ params }: Props) {
  const { referrerId } = use(params);
  const [referees, setReferees] = useState<RefereeRow[]>([]);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const r = await getReferees(referrerId);
      setReferees(r.data ?? []);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Không tải được danh sách");
    } finally {
      setLoading(false);
    }
  }, [referrerId]);

  useEffect(() => {
    load();
  }, [load]);

  const stats = {
    total: referees.length,
    active: referees.filter((r) => r.sub_status === "active").length,
    totalEarned: referees.reduce((s, r) => s + (r.total_commission || 0), 0),
    totalPaid: referees.reduce((s, r) => s + (r.paid_commission || 0), 0),
  };

  return (
    <div className="space-y-4">
      <Link href="/referrals" className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-primary">
        <ArrowLeft className="h-4 w-4" /> Quay lại
      </Link>
      <h1 className="text-2xl font-bold">Danh sách người được giới thiệu</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatBox icon={<Users className="h-5 w-5" />} label="Tổng" value={`${stats.total} người`} />
        <StatBox icon={<CheckCircle2 className="h-5 w-5 text-green-600" />} label="Đang hoạt động" value={`${stats.active} người`} />
        <StatBox icon={<Banknote className="h-5 w-5 text-green-700" />} label="Tổng hoa hồng" value={fmtVND(stats.totalEarned)} />
        <StatBox icon={<Banknote className="h-5 w-5 text-blue-700" />} label="Đã trả" value={fmtVND(stats.totalPaid)} />
      </div>

      <Card className="overflow-hidden">
        {loading ? (
          <div className="p-6 text-center text-gray-500">Đang tải…</div>
        ) : referees.length === 0 ? (
          <div className="p-6 text-center text-gray-500">Chưa có ai đăng ký qua link giới thiệu.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b text-left">
                <tr>
                  <th className="px-4 py-2 font-semibold">SĐT</th>
                  <th className="px-4 py-2 font-semibold">Tên</th>
                  <th className="px-4 py-2 font-semibold">Đăng ký</th>
                  <th className="px-4 py-2 font-semibold">Gói dịch vụ</th>
                  <th className="px-4 py-2 font-semibold text-right">Số lượt mua</th>
                  <th className="px-4 py-2 font-semibold text-right">Tổng hoa hồng</th>
                  <th className="px-4 py-2 font-semibold text-right">Đã trả</th>
                </tr>
              </thead>
              <tbody>
                {referees.map((r, i) => (
                  <tr key={r.id || i} className="border-b hover:bg-gray-50">
                    <td className="px-4 py-2 font-mono">{r.phone}</td>
                    <td className="px-4 py-2">{r.name || <span className="text-gray-400">(ẩn)</span>}</td>
                    <td className="px-4 py-2">{fmtDate(r.registered_at)}</td>
                    <td className="px-4 py-2">
                      {r.sub_status === "active" ? (
                        <span className="text-green-700">
                          Active{r.sub_expires_at && <span className="text-xs text-gray-500"> · hết {fmtDate(r.sub_expires_at)}</span>}
                        </span>
                      ) : r.sub_status === "expired" ? (
                        <span className="text-gray-500">Hết hạn</span>
                      ) : (
                        <span className="text-gray-400">Chưa mua</span>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right">{r.commission_count}</td>
                    <td className="px-4 py-2 text-right font-semibold">{fmtVND(r.total_commission)}</td>
                    <td className="px-4 py-2 text-right text-blue-700">{fmtVND(r.paid_commission)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}

function StatBox({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="border rounded-lg p-3 bg-white">
      <div className="flex items-center gap-2 text-gray-500 text-xs">
        {icon}
        <span>{label}</span>
      </div>
      <div className="text-lg font-semibold mt-1">{value}</div>
    </div>
  );
}
