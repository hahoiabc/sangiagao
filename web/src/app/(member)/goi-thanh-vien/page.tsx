"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { Crown, Check, Clock, CalendarDays, AlertTriangle, Info, CheckCircle, XCircle, QrCode, X, Loader2 } from "lucide-react";
import Image from "next/image";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getSubscriptionStatus,
  getSubscriptionPlans,
  getSubscriptionHistory,
  createPaymentOrder,
  getPaymentStatus,
  type SubscriptionStatus,
  type SubscriptionPlan,
  type SubscriptionHistory,
  type PaymentQRInfo,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { formatDate } from "@/lib/utils";
import { toast } from "sonner";

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("vi-VN").format(amount) + "đ";
}

export default function SubscriptionPage() {
  const { user } = useAuth();
  const [status, setStatus] = useState<SubscriptionStatus | null>(null);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [history, setHistory] = useState<SubscriptionHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [paymentQR, setPaymentQR] = useState<PaymentQRInfo | null>(null);
  const [paymentLoading, setPaymentLoading] = useState(false);
  const [paymentStatus, setPaymentStatus] = useState<string>("pending");
  const pollRef = useRef<NodeJS.Timeout | null>(null);

  const loadData = useCallback(() => {
    if (!user) return;
    Promise.all([
      getSubscriptionStatus("").catch(() => null),
      getSubscriptionPlans("").then((r) => r.plans).catch(() => []),
      getSubscriptionHistory("", 1, 20).then((r) => r.data).catch(() => []),
    ]).then(([s, p, h]) => {
      setStatus(s);
      setPlans(p);
      setHistory(h);
      setLoading(false);
    });
  }, [user]);

  useEffect(() => { loadData(); }, [loadData]);

  // Cleanup poll on unmount
  useEffect(() => { return () => { if (pollRef.current) clearInterval(pollRef.current); }; }, []);

  async function handlePayment(months: number) {
    setPaymentLoading(true);
    try {
      const qr = await createPaymentOrder("", months);
      setPaymentQR(qr);
      setPaymentStatus("pending");
      // Start polling every 5 seconds
      if (pollRef.current) clearInterval(pollRef.current);
      pollRef.current = setInterval(async () => {
        try {
          const order = await getPaymentStatus("", qr.order_id);
          if (order.status === "paid") {
            setPaymentStatus("paid");
            if (pollRef.current) clearInterval(pollRef.current);
            loadData(); // Refresh subscription status
          } else if (order.status === "expired") {
            setPaymentStatus("expired");
            if (pollRef.current) clearInterval(pollRef.current);
          }
        } catch { /* ignore */ }
      }, 5000);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tạo đơn thất bại");
    } finally {
      setPaymentLoading(false);
    }
  }

  function closePayment() {
    if (pollRef.current) clearInterval(pollRef.current);
    setPaymentQR(null);
    setPaymentStatus("pending");
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full rounded-lg" />
        <Skeleton className="h-60 w-full rounded-lg" />
      </div>
    );
  }

  // Calculate discount based on 1-month plan price (like mobile)
  const activePlans = plans.filter((p) => p.is_active);
  const oneMonthPlan = activePlans.find((p) => p.months === 1);
  const baseMonthlyPrice = oneMonthPlan ? oneMonthPlan.amount : 0;

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Gói dịch vụ</h1>

      {/* Status card - gradient like mobile */}
      <div className={`rounded-xl p-6 mb-6 text-white ${status?.has_active ? "bg-gradient-to-br from-primary to-primary/70" : "bg-gradient-to-br from-red-500 to-red-400"}`}>
        <div className="flex items-center gap-4">
          <div className="h-14 w-14 rounded-full bg-white/20 flex items-center justify-center">
            {status?.has_active ? (
              <CheckCircle className="h-8 w-8" />
            ) : (
              <XCircle className="h-8 w-8" />
            )}
          </div>
          <div>
            <p className="text-lg font-bold">
              {status?.has_active ? "Đang hoạt động" : "Đã hết hạn"}
            </p>
            {status?.has_active && status.plan && (
              <p className="text-sm text-white/80">{status.plan === 'free_trial' ? 'Dùng thử miễn phí' : 'Gói trả phí'}</p>
            )}
            {status?.has_active && status.days_remaining > 0 && (
              <p className="text-2xl font-bold mt-1">Còn {status.days_remaining} ngày</p>
            )}
            {status?.expires_at && (
              <p className="text-sm text-white/80">Hạn: {formatDate(status.expires_at)}</p>
            )}
            {!status?.has_active && (
              <p className="text-sm text-white/80 mt-1">Kích hoạt gói thành viên để đăng tin bán gạo</p>
            )}
          </div>
        </div>
      </div>

      {/* Warning card (if expired) - like mobile */}
      {status && !status.has_active && (
        <div className="rounded-lg border border-orange-200 bg-orange-50 p-4 mb-6 flex items-start gap-3">
          <AlertTriangle className="h-5 w-5 text-orange-500 flex-shrink-0 mt-0.5" />
          <div>
            <p className="font-medium text-orange-800">Gói đã hết hạn</p>
            <p className="text-sm text-orange-700 mt-1">
              Tin đăng của bạn đã bị tạm ẩn khỏi sàn. Gia hạn gói để tin đăng được hiển thị trở lại.
            </p>
          </div>
        </div>
      )}

      {/* Plans grid - with discount badges like mobile */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="text-lg">Các gói dịch vụ</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            {activePlans.map((plan) => {
              const pricePerMonth = Math.round(plan.amount / plan.months);
              const discount = baseMonthlyPrice > 0 && plan.months > 1
                ? Math.round((1 - pricePerMonth / baseMonthlyPrice) * 100)
                : 0;
              return (
                <div key={plan.id} className="relative rounded-xl border-2 p-4 text-center hover:border-primary transition-colors">
                  {/* Discount badge - like mobile */}
                  {discount > 0 && (
                    <div className="absolute -top-2.5 -right-2.5 bg-red-500 text-white text-xs font-bold px-2 py-0.5 rounded-full">
                      -{discount}%
                    </div>
                  )}
                  <h3 className="font-bold text-lg mb-2">{plan.label}</h3>
                  {/* Original price crossed out (if discount) */}
                  {discount > 0 && (
                    <p className="text-sm text-muted-foreground line-through">
                      {formatCurrency(baseMonthlyPrice * plan.months)}
                    </p>
                  )}
                  <p className="text-2xl font-bold text-primary mb-1">
                    {formatCurrency(plan.amount)}
                  </p>
                  {plan.months > 1 && (
                    <p className="text-xs text-muted-foreground mb-3">
                      ~ {formatCurrency(pricePerMonth)}/tháng
                    </p>
                  )}
                  <Button
                    className="w-full mt-3 gap-2"
                    onClick={() => handlePayment(plan.months)}
                    disabled={paymentLoading}
                  >
                    {paymentLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <QrCode className="h-4 w-4" />}
                    Thanh toán
                  </Button>
                </div>
              );
            })}
          </div>

          <p className="text-xs text-muted-foreground text-center mt-4">
            Chọn gói và thanh toán bằng QR chuyển khoản ngân hàng. Gói kích hoạt tự động sau khi thanh toán.
          </p>
        </CardContent>
      </Card>

      {/* Subscription History - like mobile */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Lịch sử gia hạn ({history.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {history.length > 0 ? (
            <div className="space-y-3">
              {history.map((h) => (
                <div key={h.id} className="flex items-center gap-4 p-3 rounded-lg border">
                  {/* Status icon - like mobile */}
                  <div className={`h-10 w-10 rounded-full flex items-center justify-center flex-shrink-0 ${h.status === "active" ? "bg-green-100" : "bg-gray-100"}`}>
                    {h.status === "active" ? (
                      <CheckCircle className="h-5 w-5 text-green-600" />
                    ) : (
                      <Clock className="h-5 w-5 text-gray-500" />
                    )}
                  </div>
                  <div className="flex-1 space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{h.plan === 'free_trial' ? 'Dùng thử 30 ngày' : `Gói ${h.plan_months} tháng`}</span>
                      <Badge variant={h.status === "active" ? "default" : "secondary"} className="text-xs">
                        {h.status === "active" ? "Hoạt động" : "Hết hạn"}
                      </Badge>
                    </div>
                    <p className={`text-sm font-bold ${h.plan === 'free_trial' ? 'text-green-600' : 'text-primary'}`}>
                      {h.plan === 'free_trial' ? 'Miễn phí' : formatCurrency(h.amount)}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <CalendarDays className="h-3 w-3" />
                      {formatDate(h.starts_at)} - {formatDate(h.expires_at)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">Chưa có lịch sử gia hạn</p>
          )}
        </CardContent>
      </Card>
      {/* Payment QR Modal */}
      {paymentQR && (
        <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4" onClick={closePayment}>
          <div className="bg-background rounded-xl max-w-md w-full p-6 relative" onClick={(e) => e.stopPropagation()}>
            <button onClick={closePayment} className="absolute top-3 right-3 text-muted-foreground hover:text-foreground">
              <X className="h-5 w-5" />
            </button>

            {paymentStatus === "paid" ? (
              <div className="text-center py-8">
                <CheckCircle className="h-16 w-16 text-green-500 mx-auto mb-4" />
                <h3 className="text-xl font-bold mb-2">Thanh toán thành công!</h3>
                <p className="text-muted-foreground">Gói dịch vụ đã được kích hoạt tự động.</p>
                <Button className="mt-4" onClick={closePayment}>Đóng</Button>
              </div>
            ) : paymentStatus === "expired" ? (
              <div className="text-center py-8">
                <XCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
                <h3 className="text-xl font-bold mb-2">Đơn đã hết hạn</h3>
                <p className="text-muted-foreground">Vui lòng tạo đơn mới để thanh toán.</p>
                <Button className="mt-4" onClick={closePayment}>Đóng</Button>
              </div>
            ) : (
              <>
                <h3 className="text-lg font-bold mb-4 text-center">Thanh toán chuyển khoản</h3>
                <div className="flex justify-center mb-4">
                  <Image
                    src={paymentQR.qr_url}
                    alt="QR Thanh toán"
                    width={250}
                    height={250}
                    className="rounded-lg"
                    unoptimized
                  />
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Ngân hàng</span>
                    <span className="font-medium">{paymentQR.bank_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Số tài khoản</span>
                    <span className="font-mono font-medium">{paymentQR.account_no}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Chủ tài khoản</span>
                    <span className="font-medium">{paymentQR.account_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Số tiền</span>
                    <span className="font-bold text-primary">{formatCurrency(paymentQR.amount)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Nội dung CK</span>
                    <span className="font-mono font-bold text-primary">{paymentQR.order_code}</span>
                  </div>
                </div>
                <div className="mt-4 p-3 rounded-lg bg-amber-50 border border-amber-200">
                  <p className="text-xs text-amber-800">
                    <strong>Lưu ý:</strong> Vui lòng chuyển khoản <strong>đúng số tiền</strong> và <strong>đúng nội dung</strong> để hệ thống tự động kích hoạt gói.
                  </p>
                </div>
                <div className="mt-4 flex items-center justify-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Đang chờ thanh toán...
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
