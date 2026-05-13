"use client";

import { useEffect, useState, useCallback } from "react";
import { Copy, Share2, Users, CheckCircle2, Clock, Banknote, Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  getReferralStats,
  getReferralHistory,
  becomeAffiliate,
  type ReferralStats,
  type CommissionRecord,
} from "@/services/api";

const fmtVND = (n: number) =>
  new Intl.NumberFormat("vi-VN").format(n) + " đ";

const fmtDate = (iso: string) => {
  try {
    return new Date(iso).toLocaleDateString("vi-VN");
  } catch {
    return iso;
  }
};

const STATUS_META: Record<string, { label: string; color: string }> = {
  pending: { label: "Chờ đối soát", color: "bg-orange-100 text-orange-700" },
  payable: { label: "Có thể nhận", color: "bg-green-100 text-green-700" },
  paid: { label: "Đã nhận", color: "bg-blue-100 text-blue-700" },
  cancelled: { label: "Đã hủy", color: "bg-red-100 text-red-700" },
};

export default function GioiThieuBanPage() {
  const { user } = useAuth();
  const isMember = user?.role === "member";
  const isAff = user?.role === "aff";
  const [stats, setStats] = useState<ReferralStats | null>(null);
  const [history, setHistory] = useState<CommissionRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [activating, setActivating] = useState(false);

  const handleActivate = async () => {
    if (!confirm(
      "Khi bạn giới thiệu bạn bè đăng ký + mua gói, bạn nhận hoa hồng:\n\n" +
      "• Giai đoạn 1 (3 tháng đầu): 50%\n" +
      "• Giai đoạn 2 (6 tháng kế): 30%\n" +
      "• Giai đoạn 3 (vĩnh viễn): 20%\n\n" +
      "Tính trên doanh thu ròng. Thanh toán sau 45 ngày từ giao dịch.\n\nĐồng ý kích hoạt?",
    )) return;
    setActivating(true);
    try {
      await becomeAffiliate();
      toast.success("Đã kích hoạt vai trò đối tác. Tải lại trang…");
      setTimeout(() => window.location.reload(), 1200);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Không thể kích hoạt");
    } finally {
      setActivating(false);
    }
  };

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [s, h] = await Promise.all([getReferralStats(), getReferralHistory(50)]);
      setStats(s);
      setHistory(h.data ?? []);
    } catch {
      toast.error("Không tải được dữ liệu giới thiệu");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) {
    return <div className="p-8 text-center text-gray-500">Đang tải…</div>;
  }
  if (!stats) {
    return <div className="p-8 text-center text-red-600">Không tải được dữ liệu</div>;
  }

  const shareLink = `https://sangiagao.vn/r/${stats.code}`;
  const shareMessage = `Tham gia Sàn Giá Gạo qua link giới thiệu của tôi: ${shareLink}`;

  const copy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`Đã sao chép ${label}`);
  };

  const share = async () => {
    if (typeof navigator !== "undefined" && "share" in navigator) {
      try {
        await navigator.share({ title: "Sàn Giá Gạo - Mã giới thiệu", text: shareMessage, url: shareLink });
      } catch {}
    } else {
      copy(shareMessage, "tin nhắn");
    }
  };

  return (
    <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Giới thiệu bạn bè</h1>
        <p className="text-gray-600 text-sm mt-1">
          Chia sẻ link để bạn bè đăng ký Sàn Giá Gạo và nhận hoa hồng mỗi khi họ mua gói thành viên.
        </p>
      </div>

      {isMember && (
        <Card className="bg-amber-50 border-amber-200">
          <CardContent className="p-4 space-y-3">
            <div className="flex items-center gap-2">
              <Star className="h-5 w-5 text-amber-600" />
              <h2 className="font-semibold">Trở thành đối tác chính thức</h2>
            </div>
            <p className="text-sm text-gray-700">
              Kích hoạt vai trò Đối tác để xem chi tiết người bạn giới thiệu, theo dõi hoa hồng theo từng giai đoạn,
              và nhận tiền khi đạt ngưỡng tối thiểu.
            </p>
            <Button
              onClick={handleActivate}
              disabled={activating}
              className="bg-amber-600 hover:bg-amber-700 text-white"
            >
              {activating ? "Đang kích hoạt…" : "Kích hoạt làm đối tác"}
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Code + Share */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Mã giới thiệu của bạn</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="text-4xl font-bold tracking-widest text-primary">{stats.code}</div>
            <Button variant="outline" size="sm" onClick={() => copy(stats.code, "mã")}>
              <Copy className="h-4 w-4 mr-1" /> Sao chép
            </Button>
          </div>
          <div className="flex items-center gap-2 p-2 bg-gray-50 rounded border text-sm">
            <span className="flex-1 text-blue-700 break-all">{shareLink}</span>
            <Button variant="ghost" size="sm" onClick={() => copy(shareLink, "link")}>
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <Button onClick={share} className="w-full">
            <Share2 className="h-4 w-4 mr-2" /> Chia sẻ ngay
          </Button>
        </CardContent>
      </Card>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatBox icon={<Users className="h-5 w-5" />} label="Đã giới thiệu" value={`${stats.total_referrals} người`} />
        <StatBox icon={<CheckCircle2 className="h-5 w-5 text-green-600" />} label="Đang hoạt động" value={`${stats.active_referees}`} />
        <StatBox icon={<Banknote className="h-5 w-5 text-green-600" />} label="Có thể nhận" value={fmtVND(stats.payable_amount)} />
        <StatBox icon={<Clock className="h-5 w-5 text-orange-500" />} label="Chờ đối soát" value={fmtVND(stats.pending_amount)} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Tổng quan tài chính</CardTitle>
        </CardHeader>
        <CardContent>
          <Row label="Tổng hoa hồng đã ghi nhận" value={fmtVND(stats.total_earned)} />
          <Row label="Có thể nhận" value={fmtVND(stats.payable_amount)} valueClass="text-green-700" />
          <Row label="Chờ đối soát (T+45 ngày)" value={fmtVND(stats.pending_amount)} valueClass="text-orange-600" />
          <Row label="Đã nhận" value={fmtVND(stats.paid_amount)} valueClass="text-blue-700" />
          <div className="text-xs text-gray-500 mt-3">
            Ngưỡng thanh toán tối thiểu: <strong>{fmtVND(stats.minimum_payout)}</strong>. Khi tích đủ, admin sẽ liên hệ chuyển khoản.
          </div>
        </CardContent>
      </Card>

      {/* History — only meaningful for aff */}
      {isAff && (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Lịch sử hoa hồng</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {history.length === 0 ? (
            <div className="p-6 text-center text-sm text-gray-500">
              Chưa có hoa hồng nào. Hãy chia sẻ link để bắt đầu!
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b text-left">
                <tr>
                  <th className="px-4 py-2 font-semibold">Ngày</th>
                  <th className="px-4 py-2 font-semibold">Nguồn</th>
                  <th className="px-4 py-2 font-semibold">Giai đoạn</th>
                  <th className="px-4 py-2 font-semibold text-right">Hoa hồng</th>
                  <th className="px-4 py-2 font-semibold">Trạng thái</th>
                </tr>
              </thead>
              <tbody>
                {history.map((r) => {
                  const meta = STATUS_META[r.status] ?? STATUS_META.pending;
                  return (
                    <tr key={r.id} className="border-b hover:bg-gray-50">
                      <td className="px-4 py-2">{fmtDate(r.created_at)}</td>
                      <td className="px-4 py-2 uppercase text-xs text-gray-500">{r.payment_source}</td>
                      <td className="px-4 py-2">
                        Giai đoạn {r.stage} ({(r.rate * 100).toFixed(0)}%)
                      </td>
                      <td className="px-4 py-2 text-right font-semibold">{fmtVND(r.commission_amount)}</td>
                      <td className="px-4 py-2">
                        <span className={`px-2 py-1 rounded text-xs ${meta.color}`}>{meta.label}</span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>
      )}
    </main>
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

function Row({ label, value, valueClass }: { label: string; value: string; valueClass?: string }) {
  return (
    <div className="flex justify-between py-1.5 text-sm">
      <span className="text-gray-600">{label}</span>
      <span className={`font-semibold ${valueClass ?? ""}`}>{value}</span>
    </div>
  );
}
