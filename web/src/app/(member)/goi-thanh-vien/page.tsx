"use client";

import { useEffect, useState } from "react";
import { Crown, Check, Clock, CalendarDays, AlertTriangle, Info, CheckCircle, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getSubscriptionStatus,
  getSubscriptionPlans,
  getSubscriptionHistory,
  type SubscriptionStatus,
  type SubscriptionPlan,
  type SubscriptionHistory,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { formatDate } from "@/lib/utils";

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("vi-VN").format(amount) + "đ";
}

export default function SubscriptionPage() {
  const { user } = useAuth();
  const [status, setStatus] = useState<SubscriptionStatus | null>(null);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [history, setHistory] = useState<SubscriptionHistory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user) {
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
    }
  }, [user]);

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
                  <ul className="text-sm text-muted-foreground space-y-1 text-left mt-3">
                    <li className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-green-500 flex-shrink-0" />
                      Đăng tin không giới hạn
                    </li>
                    <li className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-green-500 flex-shrink-0" />
                      Chat trực tiếp
                    </li>
                    <li className="flex items-center gap-2">
                      <Check className="h-4 w-4 text-green-500 flex-shrink-0" />
                      Hiển thị trên bảng giá
                    </li>
                  </ul>
                </div>
              );
            })}
          </div>

          {/* Renewal info card - like mobile */}
          <div className="rounded-lg border bg-blue-50 border-blue-200 p-4 mt-4 flex items-start gap-3">
            <Info className="h-5 w-5 text-blue-500 flex-shrink-0 mt-0.5" />
            <div>
              <p className="font-medium text-blue-800">Gia hạn gói dịch vụ</p>
              <p className="text-sm text-blue-700 mt-1">
                Để gia hạn hoặc đăng ký gói dịch vụ, vui lòng liên hệ bộ phận hỗ trợ qua mục &quot;Góp ý&quot; trong trang tài khoản.
              </p>
            </div>
          </div>
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
    </div>
  );
}
