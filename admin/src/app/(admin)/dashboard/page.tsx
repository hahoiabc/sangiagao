"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/lib/auth";
import { getDashboardStats, getDashboardCharts, type DashboardCharts } from "@/services/api";
import { Users, ShoppingBasket, CreditCard, Flag, Eye, Star } from "lucide-react";
import {
  AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend,
} from "recharts";

const statCards = [
  { key: "total_users", label: "Tổng người dùng", icon: Users, color: "stat-blue" },
  { key: "total_listings", label: "Tổng tin đăng", icon: ShoppingBasket, color: "stat-amber" },
  { key: "active_listings", label: "Tin đang hoạt động", icon: Eye, color: "stat-green" },
  { key: "active_subscriptions", label: "Gói đang hoạt động", icon: CreditCard, color: "stat-cyan" },
  { key: "pending_reports", label: "Báo cáo chờ xử lý", icon: Flag, color: "stat-rose" },
  { key: "total_ratings", label: "Đánh giá", icon: Star, color: "stat-purple" },
];

const CHART_COLORS = [
  "oklch(0.588 0.200 264)",   // indigo
  "oklch(0.6 0.180 185)",     // teal
  "oklch(0.646 0.222 41)",    // orange
  "oklch(0.628 0.200 304)",   // purple
  "oklch(0.70 0.150 145)",    // green
  "oklch(0.60 0.200 25)",     // red
  "oklch(0.75 0.150 85)",     // amber
  "oklch(0.55 0.200 220)",    // blue
];

const ROLE_LABELS: Record<string, string> = {
  owner: "Chủ sở hữu",
  admin: "Quản trị viên",
  editor: "Biên tập viên",
  member: "Thành viên",
};

export default function DashboardPage() {
  const { token } = useAuth();
  const [stats, setStats] = useState<Record<string, number>>({});
  const [charts, setCharts] = useState<DashboardCharts | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    Promise.all([
      getDashboardStats(token),
      getDashboardCharts(token),
    ])
      .then(([s, c]) => { setStats(s); setCharts(c); })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [token]);

  return (
    <div className="space-y-6">
      {/* KPI Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {statCards.map(({ key, label, icon: Icon, color }) => (
          <Card key={key} className={`${color} shadow-sm hover:shadow-md transition-shadow`}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{label}</CardTitle>
              <div className="rounded-lg p-2" style={{ backgroundColor: "var(--stat-bg)" }}>
                <Icon className="h-3.5 w-3.5" style={{ color: "var(--stat-icon)" }} />
              </div>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="h-8 w-16 animate-pulse rounded bg-muted" />
              ) : (
                <div className="text-2xl font-bold">{(stats[key] ?? 0).toLocaleString("vi-VN")}</div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Row 1: Area charts - trends */}
      <div className="grid gap-5 lg:grid-cols-2">
        {/* Người dùng mới theo tháng */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Người dùng mới theo tháng</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[260px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <AreaChart data={charts.users_by_month}>
                  <defs>
                    <linearGradient id="gradUsers" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="oklch(0.588 0.200 264)" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="oklch(0.588 0.200 264)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="oklch(0.9 0.01 264)" />
                  <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                  <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                  <Tooltip formatter={(v) => [v, "Người dùng"]} />
                  <Area type="monotone" dataKey="count" stroke="oklch(0.588 0.200 264)" fill="url(#gradUsers)" strokeWidth={2} />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        {/* Tin đăng mới theo tháng */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Tin đăng mới theo tháng</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[260px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <AreaChart data={charts.listings_by_month}>
                  <defs>
                    <linearGradient id="gradListings" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="oklch(0.6 0.180 185)" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="oklch(0.6 0.180 185)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="oklch(0.9 0.01 264)" />
                  <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                  <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                  <Tooltip formatter={(v) => [v, "Tin đăng"]} />
                  <Area type="monotone" dataKey="count" stroke="oklch(0.6 0.180 185)" fill="url(#gradListings)" strokeWidth={2} />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Row 2: Bar + Pie */}
      <div className="grid gap-5 lg:grid-cols-3">
        {/* Loại gạo phổ biến - bar chart */}
        <Card className="lg:col-span-2 shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Loại gạo phổ biến</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[280px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <BarChart data={charts.listings_by_rice_type} layout="vertical" margin={{ left: 10 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="oklch(0.9 0.01 264)" horizontal={false} />
                  <XAxis type="number" allowDecimals={false} tick={{ fontSize: 12 }} />
                  <YAxis dataKey="label" type="category" width={110} tick={{ fontSize: 11 }} />
                  <Tooltip formatter={(v) => [v, "Tin đăng"]} />
                  <Bar dataKey="count" radius={[0, 4, 4, 0]} barSize={20}>
                    {charts.listings_by_rice_type.map((_, i) => (
                      <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        {/* Người dùng theo vai trò - pie chart */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Vai trò người dùng</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[280px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <PieChart>
                  <Pie
                    data={charts.users_by_role.map(r => ({ ...r, name: ROLE_LABELS[r.label] || r.label }))}
                    dataKey="count"
                    nameKey="name"
                    cx="50%"
                    cy="45%"
                    outerRadius={85}
                    innerRadius={45}
                    paddingAngle={3}
                    label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                    labelLine={false}
                  >
                    {charts.users_by_role.map((_, i) => (
                      <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(v, name) => [v, name]} />
                  <Legend
                    verticalAlign="bottom"
                    formatter={(value) => <span className="text-xs">{value}</span>}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Row 3: Subscriptions + Province */}
      <div className="grid gap-5 lg:grid-cols-2">
        {/* Gói dịch vụ theo tháng */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Gói dịch vụ đăng ký theo tháng</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[260px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={charts.subs_by_month}>
                  <CartesianGrid strokeDasharray="3 3" stroke="oklch(0.9 0.01 264)" />
                  <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                  <YAxis allowDecimals={false} tick={{ fontSize: 12 }} />
                  <Tooltip formatter={(v) => [v, "Gói dịch vụ"]} />
                  <Bar dataKey="count" fill="oklch(0.628 0.200 304)" radius={[4, 4, 0, 0]} barSize={32} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        {/* Tin đăng theo tỉnh thành */}
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-semibold">Tin đăng theo tỉnh thành</CardTitle>
          </CardHeader>
          <CardContent>
            {loading || !charts ? (
              <div className="h-[260px] animate-pulse rounded bg-muted" />
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={charts.listings_by_province} layout="vertical" margin={{ left: 10 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="oklch(0.9 0.01 264)" horizontal={false} />
                  <XAxis type="number" allowDecimals={false} tick={{ fontSize: 12 }} />
                  <YAxis dataKey="label" type="category" width={90} tick={{ fontSize: 11 }} />
                  <Tooltip formatter={(v) => [v, "Tin đăng"]} />
                  <Bar dataKey="count" fill="oklch(0.70 0.150 145)" radius={[0, 4, 4, 0]} barSize={20} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
