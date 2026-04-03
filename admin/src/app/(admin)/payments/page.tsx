"use client";

import { useEffect, useState, useCallback } from "react";
import { CreditCard, ChevronLeft, ChevronRight, CheckCircle, Clock, XCircle, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getPaymentOrders, type PaymentOrder, type PaginatedResponse } from "@/services/api";

const statusConfig: Record<string, { label: string; icon: typeof CheckCircle; color: string }> = {
  paid: { label: "Đã thanh toán", icon: CheckCircle, color: "text-green-600 bg-green-50" },
  pending: { label: "Đang chờ", icon: Clock, color: "text-orange-600 bg-orange-50" },
  expired: { label: "Hết hạn", icon: XCircle, color: "text-gray-500 bg-gray-100" },
  cancelled: { label: "Đã hủy", icon: AlertTriangle, color: "text-red-500 bg-red-50" },
};

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("vi-VN").format(amount) + "đ";
}

function formatTime(dateStr: string) {
  return new Date(dateStr).toLocaleString("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" });
}

export default function PaymentsPage() {
  const [result, setResult] = useState<PaginatedResponse<PaymentOrder> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  const fetchPage = useCallback((p: number) => {
    setLoading(true);
    getPaymentOrders(p, 20)
      .then((r) => { setResult(r); setPage(p); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { fetchPage(1); }, [fetchPage]);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <CreditCard className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-2xl font-bold">Đơn thanh toán</h1>
          <p className="text-sm text-muted-foreground">Quản lý đơn thanh toán SePay</p>
        </div>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-16 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <>
          <div className="rounded-lg border overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="text-left py-3 px-4 font-medium">User</th>
                  <th className="text-left py-3 px-4 font-medium">Gói</th>
                  <th className="text-left py-3 px-4 font-medium">Số tiền</th>
                  <th className="text-left py-3 px-4 font-medium">Mã đơn</th>
                  <th className="text-left py-3 px-4 font-medium">Trạng thái</th>
                  <th className="text-left py-3 px-4 font-medium">Thời gian</th>
                </tr>
              </thead>
              <tbody>
                {result.data.map((order) => {
                  const sc = statusConfig[order.status] || statusConfig.pending;
                  const Icon = sc.icon;
                  return (
                    <tr key={order.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="py-3 px-4">
                        <p className="font-medium">{order.user_name || "Ẩn danh"}</p>
                        <p className="text-xs text-muted-foreground">{order.user_phone || ""}</p>
                      </td>
                      <td className="py-3 px-4">{order.plan_months} tháng</td>
                      <td className="py-3 px-4 font-medium">{formatCurrency(order.amount)}</td>
                      <td className="py-3 px-4 font-mono text-xs">{order.order_code}</td>
                      <td className="py-3 px-4">
                        <Badge variant="outline" className={`gap-1 ${sc.color}`}>
                          <Icon className="h-3 w-3" />
                          {sc.label}
                        </Badge>
                      </td>
                      <td className="py-3 px-4 text-xs text-muted-foreground">
                        <p>{formatTime(order.created_at)}</p>
                        {order.paid_at && <p className="text-green-600">Paid: {formatTime(order.paid_at)}</p>}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {Math.ceil(result.total / result.limit) > 1 && (
            <div className="flex items-center justify-center gap-3">
              <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => fetchPage(page - 1)}>
                <ChevronLeft className="h-4 w-4 mr-1" /> Trước
              </Button>
              <span className="text-sm text-muted-foreground">
                Trang {page} / {Math.ceil(result.total / result.limit)}
              </span>
              <Button variant="outline" size="sm" disabled={page >= Math.ceil(result.total / result.limit)} onClick={() => fetchPage(page + 1)}>
                Sau <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12 text-muted-foreground">
          <CreditCard className="h-12 w-12 mx-auto mb-3 opacity-40" />
          <p>Chưa có đơn thanh toán nào</p>
        </div>
      )}
    </div>
  );
}
