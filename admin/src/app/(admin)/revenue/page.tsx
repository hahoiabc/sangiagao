"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuth } from "@/lib/auth";
import { getSubscriptionRevenueStats, getDailyRevenue, type SubRevenueStats, type SubDailyRevenueReport } from "@/services/api";
import { cn } from "@/lib/utils";
import { TrendingUp, TrendingDown, Users, CreditCard, Clock, CheckCircle, XCircle, Download, Calendar, Search } from "lucide-react";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from "recharts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

function formatVND(amount: number) {
  return new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(amount);
}

function formatVNDPlain(amount: number) {
  return new Intl.NumberFormat("vi-VN").format(amount);
}

function exportToExcel(stats: SubRevenueStats, daily?: SubDailyRevenueReport | null) {
  const BOM = "\uFEFF";
  const rows: string[][] = [];

  rows.push(["BÁO CÁO DOANH THU GÓI ĐĂNG KÝ THÀNH VIÊN"]);
  rows.push([`Ngày xuất: ${new Date().toLocaleDateString("vi-VN")}`]);
  rows.push([]);
  rows.push(["TỔNG QUAN"]);
  rows.push(["Chỉ số", "Giá trị"]);
  rows.push(["Tổng doanh thu (VND)", formatVNDPlain(stats.total_revenue)]);
  rows.push(["Tổng gói đăng ký", String(stats.total_subscriptions)]);
  rows.push(["Đang hoạt động", String(stats.active_count)]);
  rows.push(["Đã hết hạn", String(stats.expired_count)]);
  rows.push(["Gói trả phí", String(stats.paid_count)]);
  rows.push(["Dùng thử", String(stats.trial_count)]);
  rows.push([]);

  // Monthly breakdown
  rows.push(["CHI TIẾT THEO THÁNG"]);
  rows.push(["Tháng", "Trả phí", "Dùng thử", "Tổng gói", "Doanh thu (VND)"]);

  let totalPaid = 0;
  let totalTrial = 0;
  let totalRevenue = 0;

  if (stats.monthly_revenue) {
    for (const m of stats.monthly_revenue) {
      const total = m.paid_count + m.trial_count;
      rows.push([m.month, String(m.paid_count), String(m.trial_count), String(total), formatVNDPlain(m.revenue)]);
      totalPaid += m.paid_count;
      totalTrial += m.trial_count;
      totalRevenue += m.revenue;
    }
  }

  rows.push(["TỔNG CỘNG", String(totalPaid), String(totalTrial), String(totalPaid + totalTrial), formatVNDPlain(totalRevenue)]);

  // Daily breakdown (if available)
  if (daily && daily.days && daily.days.length > 0) {
    rows.push([]);
    rows.push([`CHI TIẾT THEO NGÀY (${daily.from} đến ${daily.to})`]);
    rows.push(["Ngày", "Trả phí", "Dùng thử", "Tổng gói", "Doanh thu (VND)"]);

    for (const d of daily.days) {
      const total = d.paid_count + d.trial_count;
      rows.push([d.date, String(d.paid_count), String(d.trial_count), String(total), formatVNDPlain(d.revenue)]);
    }

    rows.push(["TỔNG CỘNG", String(daily.total_paid), String(daily.total_trial), String(daily.total_paid + daily.total_trial), formatVNDPlain(daily.total_revenue)]);
  }

  const csv = BOM + rows.map(row =>
    row.map(cell => {
      if (cell.includes(",") || cell.includes('"') || cell.includes("\n")) {
        return `"${cell.replace(/"/g, '""')}"`;
      }
      return cell;
    }).join(",")
  ).join("\n");

  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `bao-cao-doanh-thu-${new Date().toISOString().slice(0, 10)}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function StatCard({ label, value, sub, icon: Icon, color }: {
  label: string;
  value: string | number;
  sub?: string;
  icon: React.ElementType;
  color: string;
}) {
  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm">
      <div className="flex items-center gap-3">
        <div className={cn("flex h-10 w-10 items-center justify-center rounded-lg", color)}>
          <Icon className="h-5 w-5 text-white" />
        </div>
        <div>
          <p className="text-xs text-muted-foreground">{label}</p>
          <p className="text-xl font-bold">{value}</p>
          {sub && <p className="text-[11px] text-muted-foreground">{sub}</p>}
        </div>
      </div>
    </div>
  );
}

function formatDate(d: Date) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function getDefaultDateRange() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1);
  return {
    from: formatDate(from),
    to: formatDate(now),
  };
}

export default function RevenuePage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<SubRevenueStats | null>(null);
  const [loading, setLoading] = useState(true);

  // Trend chart period
  const [trendPeriod, setTrendPeriod] = useState<"day" | "week" | "month" | "quarter" | "year">("month");

  // Daily revenue state
  const defaultRange = getDefaultDateRange();
  const [dailyFrom, setDailyFrom] = useState(defaultRange.from);
  const [dailyTo, setDailyTo] = useState(defaultRange.to);
  const [dailyReport, setDailyReport] = useState<SubDailyRevenueReport | null>(null);
  const [dailyLoading, setDailyLoading] = useState(false);

  // Full daily data for trend chart (365 days)
  const [trendDailyData, setTrendDailyData] = useState<SubDailyRevenueReport | null>(null);

  const fetchStats = useCallback(async () => {
    if (!user) return;
    setLoading(true);
    try {
      const data = await getSubscriptionRevenueStats("");
      setStats(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [user]);

  const fetchDaily = useCallback(async () => {
    if (!user || !dailyFrom || !dailyTo) return;
    setDailyLoading(true);
    try {
      const data = await getDailyRevenue("", dailyFrom, dailyTo);
      setDailyReport(data);
    } catch (err) {
      console.error(err);
    } finally {
      setDailyLoading(false);
    }
  }, [user, dailyFrom, dailyTo]);

  const fetchTrendDaily = useCallback(async () => {
    if (!user) return;
    try {
      const to = new Date();
      const from = new Date();
      from.setFullYear(from.getFullYear() - 1);
      const data = await getDailyRevenue("", formatDate(from), formatDate(to));
      setTrendDailyData(data);
    } catch (err) {
      console.error(err);
    }
  }, [user]);

  useEffect(() => { fetchStats(); }, [fetchStats]);
  useEffect(() => { fetchDaily(); }, [fetchDaily]);
  useEffect(() => { fetchTrendDaily(); }, [fetchTrendDaily]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-muted-foreground">
        Đang tải thống kê doanh thu...
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex items-center justify-center py-20 text-muted-foreground">
        Không thể tải dữ liệu thống kê.
      </div>
    );
  }

  const maxMonthlyRevenue = stats.monthly_revenue?.length
    ? Math.max(...stats.monthly_revenue.map(m => m.revenue), 1)
    : 1;

  const maxMonthlyCount = stats.monthly_revenue?.length
    ? Math.max(...stats.monthly_revenue.map(m => m.paid_count + m.trial_count), 1)
    : 1;

  const maxDailyRevenue = dailyReport?.days?.length
    ? Math.max(...dailyReport.days.map(d => d.revenue), 1)
    : 1;

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold">Thống kê doanh thu</h1>
        <Button variant="outline" size="sm" onClick={() => stats && exportToExcel(stats, dailyReport)}>
          <Download className="h-4 w-4 mr-1.5" />
          Xuất Excel
        </Button>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
        <StatCard
          label="Tổng doanh thu"
          value={formatVND(stats.total_revenue)}
          icon={TrendingUp}
          color="bg-emerald-500"
        />
        <StatCard
          label="Tổng gói đăng ký"
          value={stats.total_subscriptions}
          icon={Users}
          color="bg-indigo-500"
        />
        <StatCard
          label="Đang hoạt động"
          value={stats.active_count}
          icon={CheckCircle}
          color="bg-green-500"
        />
        <StatCard
          label="Đã hết hạn"
          value={stats.expired_count}
          icon={XCircle}
          color="bg-red-500"
        />
        <StatCard
          label="Gói trả phí"
          value={stats.paid_count}
          sub={formatVND(stats.total_revenue)}
          icon={CreditCard}
          color="bg-amber-500"
        />
        <StatCard
          label="Dùng thử"
          value={stats.trial_count}
          icon={Clock}
          color="bg-blue-500"
        />
      </div>

      {/* Revenue trend chart */}
      {(() => {
        type ChartPoint = { label: string; revenue: number; change: number | null };

        function buildChartData(): ChartPoint[] {
          if (!stats) return [];
          if (trendPeriod === "day") {
            const days = trendDailyData?.days ?? [];
            const last30 = days.slice(-30);
            return last30.map((d, i) => {
              const prev = i > 0 ? last30[i - 1].revenue : null;
              const change = prev !== null && prev > 0 ? ((d.revenue - prev) / prev) * 100 : null;
              return { label: d.date.slice(5), revenue: d.revenue, change };
            });
          }

          if (trendPeriod === "week") {
            const days = trendDailyData?.days ?? [];
            const weeks: { label: string; revenue: number }[] = [];
            for (let i = 0; i < days.length; i += 7) {
              const chunk = days.slice(i, i + 7);
              const rev = chunk.reduce((s, d) => s + d.revenue, 0);
              const startDate = chunk[0].date.slice(5);
              const endDate = chunk[chunk.length - 1].date.slice(5);
              weeks.push({ label: `${startDate}~${endDate}`, revenue: rev });
            }
            const last12 = weeks.slice(-12);
            return last12.map((w, i) => {
              const prev = i > 0 ? last12[i - 1].revenue : null;
              const change = prev !== null && prev > 0 ? ((w.revenue - prev) / prev) * 100 : null;
              return { ...w, change };
            });
          }

          if (trendPeriod === "month") {
            const months = stats.monthly_revenue ?? [];
            return months.map((m, i) => {
              const prev = i > 0 ? months[i - 1].revenue : null;
              const change = prev !== null && prev > 0 ? ((m.revenue - prev) / prev) * 100 : null;
              return { label: m.month, revenue: m.revenue, change };
            });
          }

          if (trendPeriod === "quarter") {
            const months = stats.monthly_revenue ?? [];
            const quarters: { label: string; revenue: number }[] = [];
            for (let i = 0; i < months.length; i += 3) {
              const chunk = months.slice(i, i + 3);
              const rev = chunk.reduce((s, m) => s + m.revenue, 0);
              const [y, m] = chunk[0].month.split("-");
              const q = Math.ceil(Number(m) / 3);
              quarters.push({ label: `Q${q}/${y}`, revenue: rev });
            }
            return quarters.map((q, i) => {
              const prev = i > 0 ? quarters[i - 1].revenue : null;
              const change = prev !== null && prev > 0 ? ((q.revenue - prev) / prev) * 100 : null;
              return { ...q, change };
            });
          }

          // year
          const months = stats.monthly_revenue ?? [];
          const yearMap: Record<string, number> = {};
          for (const m of months) {
            const y = m.month.split("-")[0];
            yearMap[y] = (yearMap[y] ?? 0) + m.revenue;
          }
          const years = Object.entries(yearMap).sort(([a], [b]) => a.localeCompare(b)).map(([y, r]) => ({ label: y, revenue: r }));
          return years.map((yr, i) => {
            const prev = i > 0 ? years[i - 1].revenue : null;
            const change = prev !== null && prev > 0 ? ((yr.revenue - prev) / prev) * 100 : null;
            return { ...yr, change };
          });
        }

        const chartData = buildChartData();
        if (chartData.length < 2) return null;

        const last = chartData[chartData.length - 1];
        const prev = chartData[chartData.length - 2];
        const lastChange = prev.revenue > 0 ? ((last.revenue - prev.revenue) / prev.revenue) * 100 : 0;
        const isUp = lastChange >= 0;
        const avgRevenue = chartData.reduce((s, d) => s + d.revenue, 0) / chartData.length;

        const periodLabels: Record<string, string> = {
          day: "Theo ngày",
          week: "Theo tuần",
          month: "Theo tháng",
          quarter: "Theo quý",
          year: "Theo năm",
        };
        const compareLabels: Record<string, string> = {
          day: "so với hôm trước",
          week: "so với tuần trước",
          month: "so với tháng trước",
          quarter: "so với quý trước",
          year: "so với năm trước",
        };

        return (
          <div className="rounded-xl border bg-card shadow-sm overflow-hidden mb-6">
            <div className={cn(
              "px-5 py-3 flex items-center justify-between",
              isUp ? "bg-gradient-to-r from-emerald-500 to-green-500" : "bg-gradient-to-r from-red-500 to-rose-500"
            )}>
              <h3 className="text-sm font-bold text-white flex items-center gap-2">
                {isUp ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />}
                Xu hướng doanh thu
              </h3>
            </div>
            <div className="p-5">
              {/* Period tabs */}
              <div className="flex gap-1 mb-4 bg-muted rounded-lg p-1 w-fit">
                {(["day", "week", "month", "quarter", "year"] as const).map((p) => (
                  <button
                    key={p}
                    onClick={() => setTrendPeriod(p)}
                    className={cn(
                      "px-3 py-1.5 text-xs font-medium rounded-md transition-all",
                      trendPeriod === p
                        ? "bg-white dark:bg-zinc-800 shadow-sm text-foreground"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    {periodLabels[p]}
                  </button>
                ))}
              </div>

              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 10, bottom: 0 }}>
                    <defs>
                      <linearGradient id="revenueGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor={isUp ? "#10b981" : "#ef4444"} stopOpacity={0.3} />
                        <stop offset="95%" stopColor={isUp ? "#10b981" : "#ef4444"} stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis dataKey="label" tick={{ fontSize: 10 }} interval={trendPeriod === "day" ? 2 : 0} angle={trendPeriod === "week" ? -30 : 0} />
                    <YAxis
                      tick={{ fontSize: 11 }}
                      tickFormatter={(v: number) => {
                        if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(0)}tr`;
                        if (v >= 1_000) return `${(v / 1_000).toFixed(0)}k`;
                        return String(v);
                      }}
                    />
                    <Tooltip
                      formatter={(value) => [formatVND(Number(value ?? 0)), "Doanh thu"]}
                      labelStyle={{ fontWeight: 600 }}
                      contentStyle={{ borderRadius: 8, fontSize: 12 }}
                    />
                    <ReferenceLine y={avgRevenue} stroke="#94a3b8" strokeDasharray="5 5" label={{ value: "TB", position: "right", fontSize: 10, fill: "#94a3b8" }} />
                    <Area
                      type="monotone"
                      dataKey="revenue"
                      stroke={isUp ? "#10b981" : "#ef4444"}
                      strokeWidth={2.5}
                      fill="url(#revenueGrad)"
                      dot={{ r: 3, fill: isUp ? "#10b981" : "#ef4444", strokeWidth: 2, stroke: "#fff" }}
                      activeDot={{ r: 6 }}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        );
      })()}

      {/* Daily revenue report */}
      <div className="rounded-xl border bg-card shadow-sm overflow-hidden mb-6">
        <div className="bg-gradient-to-r from-orange-500 to-rose-500 px-5 py-3">
          <h3 className="text-sm font-bold text-white flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            Doanh thu theo ngày
          </h3>
        </div>
        <div className="p-5">
          {/* Date range picker */}
          <div className="flex flex-wrap items-end gap-3 mb-5">
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">Từ ngày</label>
              <Input
                type="date"
                value={dailyFrom}
                onChange={(e) => setDailyFrom(e.target.value)}
                className="w-40"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">Đến ngày</label>
              <Input
                type="date"
                value={dailyTo}
                onChange={(e) => setDailyTo(e.target.value)}
                className="w-40"
              />
            </div>
            <Button size="sm" onClick={fetchDaily} disabled={dailyLoading}>
              <Search className="h-4 w-4 mr-1.5" />
              {dailyLoading ? "Đang tải..." : "Xem báo cáo"}
            </Button>
            {/* Quick range buttons */}
            <div className="flex gap-1.5">
              {[
                { label: "7 ngày", days: 7 },
                { label: "30 ngày", days: 30 },
                { label: "90 ngày", days: 90 },
              ].map((range) => (
                <Button
                  key={range.days}
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    const to = new Date();
                    const from = new Date();
                    from.setDate(from.getDate() - range.days);
                    setDailyFrom(formatDate(from));
                    setDailyTo(formatDate(to));
                  }}
                >
                  {range.label}
                </Button>
              ))}
            </div>
          </div>

          {/* Daily summary */}
          {dailyReport && (
            <div className="grid grid-cols-3 gap-3 mb-5">
              <div className="rounded-lg bg-emerald-50 dark:bg-emerald-950/20 p-3 text-center">
                <p className="text-xs text-muted-foreground">Doanh thu</p>
                <p className="text-lg font-bold text-emerald-600">{formatVND(dailyReport.total_revenue)}</p>
              </div>
              <div className="rounded-lg bg-amber-50 dark:bg-amber-950/20 p-3 text-center">
                <p className="text-xs text-muted-foreground">Trả phí</p>
                <p className="text-lg font-bold text-amber-600">{dailyReport.total_paid}</p>
              </div>
              <div className="rounded-lg bg-blue-50 dark:bg-blue-950/20 p-3 text-center">
                <p className="text-xs text-muted-foreground">Dùng thử</p>
                <p className="text-lg font-bold text-blue-600">{dailyReport.total_trial}</p>
              </div>
            </div>
          )}

          {/* Daily table */}
          {dailyLoading ? (
            <div className="text-center py-8 text-muted-foreground">Đang tải báo cáo theo ngày...</div>
          ) : dailyReport && dailyReport.days && dailyReport.days.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="bg-orange-50/60 dark:bg-orange-950/20">
                    <th className="px-4 py-2.5 text-left text-xs font-semibold text-orange-700 dark:text-orange-300">Ngày</th>
                    <th className="px-4 py-2.5 text-right text-xs font-semibold text-orange-700 dark:text-orange-300">Trả phí</th>
                    <th className="px-4 py-2.5 text-right text-xs font-semibold text-orange-700 dark:text-orange-300">Dùng thử</th>
                    <th className="px-4 py-2.5 text-right text-xs font-semibold text-orange-700 dark:text-orange-300">Tổng</th>
                    <th className="px-4 py-2.5 text-right text-xs font-semibold text-orange-700 dark:text-orange-300">Doanh thu</th>
                    <th className="px-4 py-2.5 text-right text-xs font-semibold text-orange-700 dark:text-orange-300">Biểu đồ</th>
                  </tr>
                </thead>
                <tbody>
                  {dailyReport.days.map((d, i) => (
                    <tr key={d.date} className={cn(i % 2 === 0 ? "bg-card" : "bg-muted/30")}>
                      <td className="px-4 py-2.5 text-sm font-medium">{d.date}</td>
                      <td className="px-4 py-2.5 text-sm text-right">{d.paid_count}</td>
                      <td className="px-4 py-2.5 text-sm text-right">{d.trial_count}</td>
                      <td className="px-4 py-2.5 text-sm text-right font-medium">{d.paid_count + d.trial_count}</td>
                      <td className="px-4 py-2.5 text-sm text-right font-semibold text-emerald-600">{formatVND(d.revenue)}</td>
                      <td className="px-4 py-2.5 w-32">
                        <div className="bg-muted rounded-full h-4 overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-orange-400 to-rose-500 rounded-full transition-all duration-500"
                            style={{ width: `${Math.max((d.revenue / maxDailyRevenue) * 100, 4)}%` }}
                          />
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="bg-orange-50/60 dark:bg-orange-950/20 font-semibold">
                    <td className="px-4 py-2.5 text-sm">Tổng cộng</td>
                    <td className="px-4 py-2.5 text-sm text-right">{dailyReport.total_paid}</td>
                    <td className="px-4 py-2.5 text-sm text-right">{dailyReport.total_trial}</td>
                    <td className="px-4 py-2.5 text-sm text-right">{dailyReport.total_paid + dailyReport.total_trial}</td>
                    <td className="px-4 py-2.5 text-sm text-right text-emerald-600">{formatVND(dailyReport.total_revenue)}</td>
                    <td className="px-4 py-2.5"></td>
                  </tr>
                </tfoot>
              </table>
            </div>
          ) : dailyReport ? (
            <div className="text-center py-8 text-muted-foreground">
              Không có dữ liệu trong khoảng thời gian này.
            </div>
          ) : null}
        </div>
      </div>

      {/* Monthly revenue chart */}
      {stats.monthly_revenue && stats.monthly_revenue.length > 0 && (
        <div className="rounded-xl border bg-card shadow-sm overflow-hidden mb-6">
          <div className="bg-gradient-to-r from-emerald-500 to-teal-500 px-5 py-3">
            <h3 className="text-sm font-bold text-white">Doanh thu theo tháng (12 tháng gần nhất)</h3>
          </div>
          <div className="p-5">
            <div className="space-y-3">
              {stats.monthly_revenue.map((m) => (
                <div key={m.month} className="flex items-center gap-3">
                  <span className="text-xs text-muted-foreground w-16 shrink-0">{m.month}</span>
                  <div className="flex-1 flex items-center gap-2">
                    <div className="flex-1 bg-muted rounded-full h-6 overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-emerald-400 to-emerald-600 rounded-full flex items-center justify-end pr-2 transition-all duration-500"
                        style={{ width: `${Math.max((m.revenue / maxMonthlyRevenue) * 100, 8)}%` }}
                      >
                        <span className="text-[10px] font-semibold text-white whitespace-nowrap">
                          {formatVND(m.revenue)}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Monthly breakdown table */}
      {stats.monthly_revenue && stats.monthly_revenue.length > 0 && (
        <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
          <div className="bg-gradient-to-r from-indigo-500 to-purple-500 px-5 py-3">
            <h3 className="text-sm font-bold text-white">Chi tiết theo tháng</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-indigo-50/60 dark:bg-indigo-950/20">
                  <th className="px-4 py-2.5 text-left text-xs font-semibold text-indigo-700 dark:text-indigo-300">Tháng</th>
                  <th className="px-4 py-2.5 text-right text-xs font-semibold text-indigo-700 dark:text-indigo-300">Trả phí</th>
                  <th className="px-4 py-2.5 text-right text-xs font-semibold text-indigo-700 dark:text-indigo-300">Dùng thử</th>
                  <th className="px-4 py-2.5 text-right text-xs font-semibold text-indigo-700 dark:text-indigo-300">Tổng</th>
                  <th className="px-4 py-2.5 text-right text-xs font-semibold text-indigo-700 dark:text-indigo-300">Doanh thu</th>
                  <th className="px-4 py-2.5 text-right text-xs font-semibold text-indigo-700 dark:text-indigo-300">Biểu đồ</th>
                </tr>
              </thead>
              <tbody>
                {stats.monthly_revenue.map((m, i) => (
                  <tr key={m.month} className={cn(i % 2 === 0 ? "bg-card" : "bg-muted/30")}>
                    <td className="px-4 py-2.5 text-sm font-medium">{m.month}</td>
                    <td className="px-4 py-2.5 text-sm text-right">{m.paid_count}</td>
                    <td className="px-4 py-2.5 text-sm text-right">{m.trial_count}</td>
                    <td className="px-4 py-2.5 text-sm text-right font-medium">{m.paid_count + m.trial_count}</td>
                    <td className="px-4 py-2.5 text-sm text-right font-semibold text-emerald-600">{formatVND(m.revenue)}</td>
                    <td className="px-4 py-2.5 w-32">
                      <div className="flex gap-0.5 h-4">
                        <div
                          className="bg-amber-400 rounded-sm"
                          style={{ width: `${(m.paid_count / maxMonthlyCount) * 100}%` }}
                          title={`Trả phí: ${m.paid_count}`}
                        />
                        <div
                          className="bg-blue-400 rounded-sm"
                          style={{ width: `${(m.trial_count / maxMonthlyCount) * 100}%` }}
                          title={`Dùng thử: ${m.trial_count}`}
                        />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {(!stats.monthly_revenue || stats.monthly_revenue.length === 0) && (
        <div className="rounded-xl border bg-card p-8 text-center text-muted-foreground">
          Chưa có dữ liệu doanh thu theo tháng.
        </div>
      )}
    </div>
  );
}
