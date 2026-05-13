"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { getLeaderboard, type LeaderboardRow } from "@/services/api";
import { Card } from "@/components/ui/card";
import { useAuth } from "@/lib/auth";

const fmt = (n: number) => new Intl.NumberFormat("vi-VN").format(n) + " đ";

export default function ReferralsOverviewPage() {
  const { user } = useAuth();
  const canManage = user?.role === "owner" || user?.role === "admin";
  const [rows, setRows] = useState<LeaderboardRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const r = await getLeaderboard();
      setRows(r.data ?? []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Lỗi tải dữ liệu");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Hoa hồng giới thiệu</h1>
          <p className="text-sm text-gray-500">Quản lý đối tác, hoa hồng, payout</p>
        </div>
        <div className="flex gap-2">
          <Link href="/referrals/rules" className="px-3 py-1.5 border rounded text-sm hover:bg-gray-50">
            {canManage ? "Cài đặt quy tắc" : "Xem quy tắc"}
          </Link>
          <Link href="/referrals/payouts" className="px-3 py-1.5 bg-primary text-primary-foreground rounded text-sm">
            {canManage ? "Quản lý payout" : "Lịch sử payout"}
          </Link>
        </div>
      </div>

      <Card className="overflow-hidden">
        <div className="p-4 border-b">
          <h2 className="font-semibold">Bảng xếp hạng đối tác</h2>
        </div>
        {loading ? (
          <div className="p-6 text-center text-gray-500">Đang tải…</div>
        ) : err ? (
          <div className="p-6 text-center text-red-600">{err}</div>
        ) : rows.length === 0 ? (
          <div className="p-6 text-center text-gray-500">Chưa có hoa hồng nào.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b text-left">
                <tr>
                  <th className="px-4 py-2 font-semibold">Đối tác</th>
                  <th className="px-4 py-2 font-semibold">Mã</th>
                  <th className="px-4 py-2 font-semibold text-right">Đã giới thiệu</th>
                  <th className="px-4 py-2 font-semibold text-right">Tổng hoa hồng</th>
                  <th className="px-4 py-2 font-semibold text-right">Có thể trả</th>
                  <th className="px-4 py-2 font-semibold text-right">Chờ đối soát</th>
                  <th className="px-4 py-2 font-semibold text-right">Đã trả</th>
                  <th className="px-4 py-2 font-semibold"></th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => (
                  <tr key={r.referrer_user_id} className="border-b hover:bg-gray-50">
                    <td className="px-4 py-2">
                      <Link
                        href={`/referrals/${r.referrer_user_id}`}
                        className="font-medium hover:text-primary hover:underline"
                      >
                        {r.name || "—"}
                      </Link>
                      <div className="text-xs text-gray-500">{r.phone}</div>
                    </td>
                    <td className="px-4 py-2 font-mono">{r.code}</td>
                    <td className="px-4 py-2 text-right">{r.total_referrals}</td>
                    <td className="px-4 py-2 text-right font-semibold">{fmt(r.total_earned)}</td>
                    <td className="px-4 py-2 text-right text-green-700">{fmt(r.payable_amount)}</td>
                    <td className="px-4 py-2 text-right text-orange-600">{fmt(r.pending_amount)}</td>
                    <td className="px-4 py-2 text-right text-blue-700">{fmt(r.paid_amount)}</td>
                    <td className="px-4 py-2 text-right">
                      {canManage && r.payable_amount > 0 && (
                        <Link
                          href={`/referrals/payouts?referrer=${r.referrer_user_id}`}
                          className="text-xs text-primary hover:underline"
                        >
                          Tạo payout →
                        </Link>
                      )}
                    </td>
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
