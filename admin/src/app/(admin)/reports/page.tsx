"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { ExternalLink } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { listReports, resolveReport, type Report } from "@/services/api";

const actionOptions = [
  { value: "delete_listing", label: "Xóa tin đăng" },
  { value: "block_user", label: "Khóa người dùng" },
  { value: "warn_user", label: "Cảnh cáo" },
  { value: "dismiss", label: "Bỏ qua báo cáo" },
];

const statusTabs = [
  { value: "pending", label: "Chờ xử lý" },
  { value: "resolved", label: "Đã xử lý" },
  { value: "dismissed", label: "Đã bỏ qua" },
  { value: "all", label: "Tất cả" },
];

export default function ReportsPage() {
  const { token } = useAuth();
  const [reports, setReports] = useState<Report[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("pending");
  const [loading, setLoading] = useState(false);
  const [resolveDialog, setResolveDialog] = useState<Report | null>(null);
  const [selectedAction, setSelectedAction] = useState("");
  const [adminNote, setAdminNote] = useState("");

  const limit = 20;

  const fetchReports = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await listReports(token, page, limit, status);
      setReports(res.data);
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token, page, status]);

  useEffect(() => {
    fetchReports();
  }, [fetchReports]);

  function handleStatusChange(newStatus: string) {
    setStatus(newStatus);
    setPage(1);
  }

  async function handleResolve() {
    if (!token || !resolveDialog || !selectedAction) return;
    try {
      await resolveReport(token, resolveDialog.id, selectedAction, adminNote || undefined);
      toast.success("Đã xử lý báo cáo");
      setResolveDialog(null);
      setSelectedAction("");
      setAdminNote("");
      fetchReports();
    } catch {
      toast.error("Xử lý báo cáo thất bại");
    }
  }

  function targetLabel(type: string) {
    switch (type) {
      case "listing": return "Tin đăng";
      case "user": return "Người dùng";
      case "rating": return "Đánh giá";
      default: return type;
    }
  }

  function targetBadgeColor(type: string) {
    switch (type) {
      case "listing": return "default" as const;
      case "user": return "secondary" as const;
      case "rating": return "outline" as const;
      default: return "outline" as const;
    }
  }

  function targetLink(type: string, id: string): string | null {
    switch (type) {
      case "listing": return `/listings/${id}`;
      case "user": return `/users/${id}`;
      default: return null;
    }
  }

  function actionLabel(action: string) {
    const opt = actionOptions.find((o) => o.value === action);
    return opt?.label || action;
  }

  const showExtraColumns = status !== "pending";
  const colSpan = showExtraColumns ? 9 : 7;
  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      <h1 className="text-xl font-semibold mb-5">Quản lý báo cáo</h1>

      {/* Status tabs */}
      <div className="flex gap-2 mb-4">
        {statusTabs.map((tab) => (
          <Button
            key={tab.value}
            size="sm"
            variant={status === tab.value ? "default" : "outline"}
            onClick={() => handleStatusChange(tab.value)}
          >
            {tab.label}
          </Button>
        ))}
      </div>

      <div className="rounded-lg border shadow-sm bg-card overflow-x-auto">
        <Table className="min-w-[800px]">
          <TableHeader>
            <TableRow>
              <TableHead>Đối tượng</TableHead>
              <TableHead>ID đối tượng</TableHead>
              <TableHead>Lý do</TableHead>
              <TableHead>Mô tả</TableHead>
              <TableHead>Trạng thái</TableHead>
              {showExtraColumns && <TableHead>Hành động</TableHead>}
              {showExtraColumns && <TableHead>Ghi chú admin</TableHead>}
              <TableHead>Ngày tạo</TableHead>
              <TableHead className="text-right">Thao tác</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={colSpan} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
              </TableRow>
            ) : reports.length === 0 ? (
              <TableRow>
                <TableCell colSpan={colSpan} className="text-center py-8 text-muted-foreground">Không có báo cáo</TableCell>
              </TableRow>
            ) : (
              reports.map((report) => (
                <TableRow key={report.id}>
                  <TableCell>
                    <Badge variant={targetBadgeColor(report.target_type)}>
                      {targetLabel(report.target_type)}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {targetLink(report.target_type, report.target_id) ? (
                      <Link
                        href={targetLink(report.target_type, report.target_id)!}
                        className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline"
                      >
                        Xem nội dung
                        <ExternalLink className="h-3 w-3" />
                      </Link>
                    ) : (
                      <span className="font-mono text-xs text-muted-foreground truncate max-w-[120px] block">
                        {report.target_id}
                      </span>
                    )}
                  </TableCell>
                  <TableCell>{report.reason}</TableCell>
                  <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                    {report.description || "-"}
                  </TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border ${report.status === "pending" ? "bg-amber-50 text-amber-700 border-amber-200" : report.status === "resolved" ? "bg-emerald-50 text-emerald-700 border-emerald-200" : "bg-gray-50 text-gray-600 border-gray-200"}`}>
                      {report.status === "pending" ? "Chờ xử lý" : report.status === "resolved" ? "Đã xử lý" : report.status === "dismissed" ? "Đã bỏ qua" : report.status}
                    </span>
                  </TableCell>
                  {showExtraColumns && (
                    <TableCell className="text-sm">
                      {report.admin_action ? actionLabel(report.admin_action) : "-"}
                    </TableCell>
                  )}
                  {showExtraColumns && (
                    <TableCell className="max-w-[150px] truncate text-sm text-muted-foreground">
                      {report.admin_note || "-"}
                    </TableCell>
                  )}
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(report.created_at).toLocaleDateString("vi-VN")}
                  </TableCell>
                  <TableCell className="text-right">
                    {report.status === "pending" && (
                      <Button size="sm" onClick={() => setResolveDialog(report)}>
                        Xử lý
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
        <span className="text-sm text-muted-foreground">
          Trang {page} / {totalPages || 1} ({total} báo cáo)
        </span>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>
            Trước
          </Button>
          <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
            Sau
          </Button>
        </div>
      </div>

      <Dialog open={!!resolveDialog} onOpenChange={() => setResolveDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Xử lý báo cáo</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">
              Đối tượng: <strong>{resolveDialog ? targetLabel(resolveDialog.target_type) : ""}</strong> — Lý do: <strong>{resolveDialog?.reason}</strong>
            </p>
            {resolveDialog && targetLink(resolveDialog.target_type, resolveDialog.target_id) && (
              <Link
                href={targetLink(resolveDialog.target_type, resolveDialog.target_id)!}
                target="_blank"
                className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
              >
                Xem nội dung bị báo cáo
                <ExternalLink className="h-3.5 w-3.5" />
              </Link>
            )}
            {resolveDialog?.description && (
              <p className="text-sm bg-muted/50 rounded-md p-2">{resolveDialog.description}</p>
            )}
            <p className="text-sm font-medium">Chọn hành động:</p>
            <div className="grid grid-cols-2 gap-2">
              {actionOptions.map((opt) => (
                <Button
                  key={opt.value}
                  size="sm"
                  variant={selectedAction === opt.value ? "default" : "outline"}
                  onClick={() => setSelectedAction(opt.value)}
                >
                  {opt.label}
                </Button>
              ))}
            </div>
            <p className="text-sm font-medium pt-1">Ghi chú cho người dùng:</p>
            <textarea
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              rows={3}
              placeholder="Nhập ghi chú gửi kèm thông báo cho người gửi và người bị báo cáo..."
              value={adminNote}
              onChange={(e) => setAdminNote(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setResolveDialog(null)}>Hủy</Button>
            <Button onClick={handleResolve} disabled={!selectedAction}>
              Xác nhận
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
