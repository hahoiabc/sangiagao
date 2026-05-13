"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { getAllReferees, type AllRefereeRow } from "@/services/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";

const fmtVND = (n: number) => new Intl.NumberFormat("vi-VN").format(n) + " đ";
const fmtDate = (iso: string) => {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleDateString("vi-VN");
  } catch {
    return iso;
  }
};

const PAGE_SIZE_OPTIONS = [20, 50, 100, 200, 500];

export default function ReferralsDetailPage() {
  const { user } = useAuth();
  const canSeeAll = user?.role === "owner" || user?.role === "admin";

  const [rows, setRows] = useState<AllRefereeRow[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [referrerFilter, setReferrerFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const r = await getAllReferees(page, limit, referrerFilter || undefined);
      setRows(r.data ?? []);
      setTotal(r.total ?? 0);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Không tải được dữ liệu");
    } finally {
      setLoading(false);
    }
  }, [page, limit, referrerFilter]);

  useEffect(() => {
    load();
  }, [load]);

  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Chi tiết người được giới thiệu</h1>
        <p className="text-sm text-gray-500">
          {canSeeAll
            ? "Danh sách tất cả thành viên đã đăng ký qua link giới thiệu."
            : "Danh sách thành viên bạn đã giới thiệu thành công."}
        </p>
      </div>

      <Card className="p-3 flex flex-wrap items-center gap-3">
        {canSeeAll && (
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">Lọc theo đối tác (user ID):</label>
            <Input
              type="text"
              placeholder="UUID đối tác (bỏ trống = tất cả)"
              value={referrerFilter}
              onChange={(e) => {
                setReferrerFilter(e.target.value);
                setPage(1);
              }}
              className="w-72"
            />
          </div>
        )}
        <div className="flex items-center gap-2 ml-auto">
          <label className="text-sm text-gray-600">Hiển thị:</label>
          <select
            value={limit}
            onChange={(e) => {
              setLimit(Number(e.target.value));
              setPage(1);
            }}
            className="border rounded px-2 py-1 text-sm"
          >
            {PAGE_SIZE_OPTIONS.map((n) => (
              <option key={n} value={n}>
                {n} / trang
              </option>
            ))}
          </select>
        </div>
      </Card>

      <Card className="overflow-hidden">
        {loading ? (
          <div className="p-6 text-center text-gray-500">Đang tải…</div>
        ) : rows.length === 0 ? (
          <div className="p-6 text-center text-gray-500">Không có thành viên nào.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b text-left">
                <tr>
                  <th className="px-4 py-2 font-semibold">STT</th>
                  <th className="px-4 py-2 font-semibold">SĐT</th>
                  <th className="px-4 py-2 font-semibold">Tên</th>
                  <th className="px-4 py-2 font-semibold">Đăng ký</th>
                  {canSeeAll && <th className="px-4 py-2 font-semibold">Đối tác giới thiệu</th>}
                  <th className="px-4 py-2 font-semibold">Mã</th>
                  <th className="px-4 py-2 font-semibold">Gói</th>
                  <th className="px-4 py-2 font-semibold text-right">Số lượt mua</th>
                  <th className="px-4 py-2 font-semibold text-right">Hoa hồng</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r, i) => (
                  <tr key={r.id || `${page}-${i}`} className="border-b hover:bg-gray-50">
                    <td className="px-4 py-2 text-gray-500">{(page - 1) * limit + i + 1}</td>
                    <td className="px-4 py-2 font-mono">{r.phone}</td>
                    <td className="px-4 py-2">{r.name || <span className="text-gray-400">(ẩn)</span>}</td>
                    <td className="px-4 py-2">{fmtDate(r.registered_at)}</td>
                    {canSeeAll && (
                      <td className="px-4 py-2">
                        <Link
                          href={`/referrals/${r.referrer_user_id}`}
                          className="text-primary hover:underline"
                        >
                          {r.referrer_name || r.referrer_phone || "—"}
                        </Link>
                      </td>
                    )}
                    <td className="px-4 py-2 font-mono text-xs">{r.referrer_code}</td>
                    <td className="px-4 py-2">
                      {r.sub_status === "active" ? (
                        <span className="text-green-700">Active</span>
                      ) : r.sub_status === "expired" ? (
                        <span className="text-gray-500">Hết hạn</span>
                      ) : (
                        <span className="text-gray-400">Chưa mua</span>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right">{r.commission_count}</td>
                    <td className="px-4 py-2 text-right font-semibold">{fmtVND(r.total_commission)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {total > 0 && (
        <div className="flex items-center justify-between text-sm">
          <div className="text-gray-600">
            Tổng <strong>{total}</strong> thành viên · Trang <strong>{page}</strong> / {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              ← Trước
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            >
              Sau →
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
