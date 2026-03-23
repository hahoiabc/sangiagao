"use client";

import { useEffect, useState } from "react";
import { Crown, Check, Clock, CalendarDays } from "lucide-react";
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

export default function SubscriptionPage() {
  const { token } = useAuth();
  const [status, setStatus] = useState<SubscriptionStatus | null>(null);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [history, setHistory] = useState<SubscriptionHistory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      Promise.all([
        getSubscriptionStatus(token).catch(() => null),
        getSubscriptionPlans(token).then((r) => r.plans).catch(() => []),
        getSubscriptionHistory(token, 1, 20).then((r) => r.data).catch(() => []),
      ]).then(([s, p, h]) => {
        setStatus(s);
        setPlans(p);
        setHistory(h);
        setLoading(false);
      });
    }
  }, [token]);

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full rounded-lg" />
        <Skeleton className="h-60 w-full rounded-lg" />
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Gói dịch vụ</h1>

      <Card className="mb-6">
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
            <Crown className="h-4 w-4" />
            Trạng thái hiện tại
          </CardTitle>
        </CardHeader>
        <CardContent>
          {status?.has_active ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Badge className="bg-green-500">Đang hoạt động</Badge>
                <span className="text-sm font-medium capitalize">{status.plan}</span>
              </div>
              <p className="text-sm text-muted-foreground">
                Hết hạn: {status.expires_at ? formatDate(status.expires_at) : "—"}
              </p>
              <p className="text-sm text-muted-foreground">
                Còn lại: {status.days_remaining} ngày
              </p>
            </div>
          ) : (
            <div>
              <Badge variant="secondary">Chưa kích hoạt</Badge>
              <p className="text-sm text-muted-foreground mt-2">
                Kích hoạt gói thành viên để đăng tin bán gạo
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="text-lg">Các gói dịch vụ</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {plans.filter((p) => p.is_active).map((plan) => {
              const pricePerMonth = Math.round(plan.amount / plan.months);
              return (
                <Card key={plan.id} className="border-2">
                  <CardContent className="p-4 text-center">
                    <h3 className="font-bold text-lg mb-1">{plan.label}</h3>
                    <p className="text-2xl font-bold text-primary mb-1">
                      {new Intl.NumberFormat("vi-VN").format(plan.amount)}đ
                    </p>
                    {plan.months > 1 && (
                      <p className="text-xs text-muted-foreground mb-3">
                        ~ {new Intl.NumberFormat("vi-VN").format(pricePerMonth)}đ/tháng
                      </p>
                    )}
                    <ul className="text-sm text-muted-foreground space-y-1 text-left">
                      <li className="flex items-center gap-2">
                        <Check className="h-4 w-4 text-green-500" />
                        Đăng tin không giới hạn
                      </li>
                      <li className="flex items-center gap-2">
                        <Check className="h-4 w-4 text-green-500" />
                        Chat trực tiếp
                      </li>
                      <li className="flex items-center gap-2">
                        <Check className="h-4 w-4 text-green-500" />
                        Hiển thị trên bảng giá
                      </li>
                    </ul>
                  </CardContent>
                </Card>
              );
            })}
          </div>
          <p className="text-sm text-muted-foreground mt-4 text-center">
            Liên hệ quản trị viên để kích hoạt gói thành viên
          </p>
        </CardContent>
      </Card>

      {/* Subscription History */}
      {history.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Lịch sử gói thành viên
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {history.map((h) => (
                <div key={h.id} className="flex items-center justify-between p-3 rounded-lg border">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{h.plan_months} tháng</span>
                      <Badge variant={h.status === "active" ? "default" : "secondary"} className="text-xs">
                        {h.status === "active" ? "Đang hoạt động" : "Hết hạn"}
                      </Badge>
                    </div>
                    <p className="text-sm font-semibold text-primary">
                      {new Intl.NumberFormat("vi-VN").format(h.amount)}đ
                    </p>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <CalendarDays className="h-3 w-3" />
                        {formatDate(h.starts_at)} - {formatDate(h.expires_at)}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
