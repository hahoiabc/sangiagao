"use client";

import { useEffect, useState, useCallback } from "react";
import { Activity, Cpu, HardDrive, MemoryStick, Users, Clock, Server, Gauge } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { getSystemStats, type SystemStats } from "@/services/api";

function ProgressBar({ value, color }: { value: number; color: string }) {
  return (
    <div className="w-full h-2.5 bg-muted rounded-full overflow-hidden">
      <div
        className={`h-full rounded-full transition-all duration-500 ${color}`}
        style={{ width: `${Math.min(value, 100)}%` }}
      />
    </div>
  );
}

function getColor(percent: number) {
  if (percent < 50) return "bg-emerald-500";
  if (percent < 80) return "bg-amber-500";
  return "bg-red-500";
}

function getTextColor(percent: number) {
  if (percent < 50) return "text-emerald-600";
  if (percent < 80) return "text-amber-600";
  return "text-red-600";
}

export default function MonitoringPage() {
  const { token } = useAuth();
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchStats = useCallback(async () => {
    if (!token) return;
    try {
      const data = await getSystemStats(token);
      setStats(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    fetchStats();
    if (!autoRefresh) return;
    const interval = setInterval(fetchStats, 5000);
    return () => clearInterval(interval);
  }, [fetchStats, autoRefresh]);

  if (loading || !stats) {
    return (
      <div className="flex items-center justify-center py-20 text-muted-foreground">
        Đang tải...
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-primary" />
          <h1 className="text-xl font-semibold">Giám sát hệ thống</h1>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded"
            />
            Tự động cập nhật (5s)
          </label>
          <button
            onClick={fetchStats}
            className="text-sm text-primary hover:underline"
          >
            Làm mới
          </button>
        </div>
      </div>

      {/* Online Users - Prominent */}
      <div className="rounded-lg border shadow-sm bg-card p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-emerald-100 p-3">
              <Users className="h-6 w-6 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Người dùng đang online</p>
              <p className="text-3xl font-bold text-emerald-600">{stats.online_users}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="relative flex h-3 w-3">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500" />
            </span>
            <span className="text-sm text-muted-foreground">Realtime</span>
          </div>
        </div>
      </div>

      {/* System Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* CPU */}
        <div className="rounded-lg border shadow-sm bg-card p-4">
          <div className="flex items-center gap-2 mb-3">
            <Cpu className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">CPU</span>
          </div>
          <div className="space-y-2">
            <div className="flex items-end justify-between">
              <span className={`text-2xl font-bold ${getTextColor(stats.cpu_percent)}`}>
                {stats.cpu_percent.toFixed(1)}%
              </span>
              <span className="text-xs text-muted-foreground">{stats.cpu_cores} cores</span>
            </div>
            <ProgressBar value={stats.cpu_percent} color={getColor(stats.cpu_percent)} />
          </div>
        </div>

        {/* Memory */}
        <div className="rounded-lg border shadow-sm bg-card p-4">
          <div className="flex items-center gap-2 mb-3">
            <MemoryStick className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">RAM</span>
          </div>
          <div className="space-y-2">
            <div className="flex items-end justify-between">
              <span className={`text-2xl font-bold ${getTextColor(stats.mem_percent)}`}>
                {stats.mem_percent.toFixed(1)}%
              </span>
              <span className="text-xs text-muted-foreground">
                {(stats.mem_used_mb / 1024).toFixed(1)} / {(stats.mem_total_mb / 1024).toFixed(1)} GB
              </span>
            </div>
            <ProgressBar value={stats.mem_percent} color={getColor(stats.mem_percent)} />
          </div>
        </div>

        {/* Disk */}
        <div className="rounded-lg border shadow-sm bg-card p-4">
          <div className="flex items-center gap-2 mb-3">
            <HardDrive className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Ổ đĩa</span>
          </div>
          <div className="space-y-2">
            <div className="flex items-end justify-between">
              <span className={`text-2xl font-bold ${getTextColor(stats.disk_percent)}`}>
                {stats.disk_percent.toFixed(1)}%
              </span>
              <span className="text-xs text-muted-foreground">
                {stats.disk_used_gb} / {stats.disk_total_gb} GB
              </span>
            </div>
            <ProgressBar value={stats.disk_percent} color={getColor(stats.disk_percent)} />
          </div>
        </div>
      </div>

      {/* Server Info */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Go Runtime */}
        <div className="rounded-lg border shadow-sm bg-card p-4">
          <div className="flex items-center gap-2 mb-3">
            <Server className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Go Runtime</span>
          </div>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="rounded-md bg-muted/50 px-3 py-2">
              <span className="text-muted-foreground block text-xs">Goroutines</span>
              <span className="font-semibold">{stats.goroutines}</span>
            </div>
            <div className="rounded-md bg-muted/50 px-3 py-2">
              <span className="text-muted-foreground block text-xs">Heap Alloc</span>
              <span className="font-semibold">{stats.heap_alloc_mb} MB</span>
            </div>
            <div className="rounded-md bg-muted/50 px-3 py-2">
              <span className="text-muted-foreground block text-xs">Heap Sys</span>
              <span className="font-semibold">{stats.heap_sys_mb} MB</span>
            </div>
            <div className="rounded-md bg-muted/50 px-3 py-2">
              <span className="text-muted-foreground block text-xs">GC Cycles</span>
              <span className="font-semibold">{stats.gc_cycles}</span>
            </div>
          </div>
        </div>

        {/* Server Info */}
        <div className="rounded-lg border shadow-sm bg-card p-4">
          <div className="flex items-center gap-2 mb-3">
            <Gauge className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Thông tin server</span>
          </div>
          <div className="grid grid-cols-1 gap-2 text-sm">
            <div className="rounded-md bg-muted/50 px-3 py-2 flex justify-between">
              <span className="text-muted-foreground">Hostname</span>
              <span className="font-mono font-medium">{stats.hostname}</span>
            </div>
            <div className="rounded-md bg-muted/50 px-3 py-2 flex justify-between">
              <span className="text-muted-foreground">Go Version</span>
              <span className="font-mono font-medium">{stats.go_version}</span>
            </div>
            <div className="rounded-md bg-muted/50 px-3 py-2 flex justify-between items-center">
              <span className="text-muted-foreground flex items-center gap-1">
                <Clock className="h-3.5 w-3.5" /> Uptime
              </span>
              <span className="font-semibold">{stats.uptime}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
