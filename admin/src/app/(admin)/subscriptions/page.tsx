"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  listUsers, activateSubscription, getSubscriptionPlans,
  listAllPlans, createPlan, updatePlan, deletePlan,
  type User, type SubscriptionPlan,
} from "@/services/api";
import { cn } from "@/lib/utils";

function formatVND(amount: number) {
  return new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(amount);
}

type PlanFormData = {
  months: number;
  amount: number;
  label: string;
  is_active?: boolean;
};

export default function SubscriptionsPage() {
  const { token, user: currentUser } = useAuth();
  const isOwner = currentUser?.role === "owner";

  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);

  const [activateDialog, setActivateDialog] = useState<User | null>(null);
  const [selectedMonths, setSelectedMonths] = useState(1);
  const [activateError, setActivateError] = useState("");
  const [activating, setActivating] = useState(false);

  // Plan CRUD state (owner only)
  const [allPlans, setAllPlans] = useState<SubscriptionPlan[]>([]);
  const [plansLoading, setPlansLoading] = useState(false);
  const [planDialogOpen, setPlanDialogOpen] = useState(false);
  const [editingPlan, setEditingPlan] = useState<SubscriptionPlan | null>(null);
  const [planForm, setPlanForm] = useState<PlanFormData>({ months: 1, amount: 0, label: "" });
  const [planSaving, setPlanSaving] = useState(false);
  const [planError, setPlanError] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<SubscriptionPlan | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [showPlanManager, setShowPlanManager] = useState(false);

  const limit = 20;

  const fetchPlans = useCallback(async () => {
    if (!token) return;
    try {
      const res = await getSubscriptionPlans(token);
      setPlans(res.plans || []);
    } catch (err) {
      console.error(err);
    }
  }, [token]);

  const fetchAllPlans = useCallback(async () => {
    if (!token || !isOwner) return;
    setPlansLoading(true);
    try {
      const res = await listAllPlans(token);
      setAllPlans(res.plans || []);
    } catch (err) {
      console.error(err);
      toast.error("Không thể tải danh sách gói");
    } finally {
      setPlansLoading(false);
    }
  }, [token, isOwner]);

  const fetchUsers = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await listUsers(token, search, page, limit);
      setUsers(res.data.filter((u) => u.role !== "admin"));
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token, search, page]);

  useEffect(() => {
    fetchPlans();
    fetchUsers();
    if (isOwner) fetchAllPlans();
  }, [fetchPlans, fetchUsers, fetchAllPlans, isOwner]);

  async function handleActivate() {
    if (!token || !activateDialog) return;
    setActivateError("");
    setActivating(true);
    try {
      await activateSubscription(token, activateDialog.id, selectedMonths);
      const plan = plans.find(p => p.months === selectedMonths);
      toast.success(`Đã kích hoạt gói ${plan?.label} (${formatVND(plan?.amount || 0)}) cho ${activateDialog.name || activateDialog.phone}`);
      setActivateDialog(null);
      setSelectedMonths(1);
      fetchUsers();
    } catch (err) {
      setActivateError(err instanceof Error ? err.message : "Không thể kích hoạt gói dịch vụ");
    } finally {
      setActivating(false);
    }
  }

  // Plan CRUD handlers
  function openCreatePlan() {
    setEditingPlan(null);
    setPlanForm({ months: 1, amount: 0, label: "" });
    setPlanError("");
    setPlanDialogOpen(true);
  }

  function openEditPlan(plan: SubscriptionPlan) {
    setEditingPlan(plan);
    setPlanForm({ months: plan.months, amount: plan.amount, label: plan.label, is_active: plan.is_active });
    setPlanError("");
    setPlanDialogOpen(true);
  }

  async function handleSavePlan() {
    if (!token) return;
    if (!planForm.label.trim() || planForm.months <= 0 || planForm.amount < 0) {
      setPlanError("Vui lòng nhập đầy đủ thông tin hợp lệ");
      return;
    }
    setPlanSaving(true);
    setPlanError("");
    try {
      if (editingPlan) {
        await updatePlan(token, editingPlan.id, {
          months: planForm.months,
          amount: planForm.amount,
          label: planForm.label,
          is_active: planForm.is_active,
        });
        toast.success("Đã cập nhật gói dịch vụ");
      } else {
        await createPlan(token, {
          months: planForm.months,
          amount: planForm.amount,
          label: planForm.label,
        });
        toast.success("Đã tạo gói dịch vụ mới");
      }
      setPlanDialogOpen(false);
      fetchAllPlans();
      fetchPlans();
    } catch (err) {
      setPlanError(err instanceof Error ? err.message : "Không thể lưu gói dịch vụ");
    } finally {
      setPlanSaving(false);
    }
  }

  async function handleDeletePlan() {
    if (!token || !deleteTarget) return;
    setDeleting(true);
    try {
      await deletePlan(token, deleteTarget.id);
      toast.success(`Đã xóa gói "${deleteTarget.label}"`);
      setDeleteTarget(null);
      fetchAllPlans();
      fetchPlans();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Không thể xóa gói");
    } finally {
      setDeleting(false);
    }
  }

  const totalPages = Math.ceil(total / limit);
  const selectedPlan = plans.find(p => p.months === selectedMonths);

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold">Quản lý gói dịch vụ</h1>
        {isOwner && (
          <Button variant={showPlanManager ? "default" : "outline"} onClick={() => setShowPlanManager(!showPlanManager)}>
            {showPlanManager ? "Ẩn cài đặt gói" : "Cài đặt gói"}
          </Button>
        )}
      </div>

      {/* Plan cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
        {plans.map((plan) => {
          const basePrice = (plans.find(p => p.months === 1)?.amount || plan.amount / plan.months) * plan.months;
          const discount = basePrice > 0 ? Math.round((1 - plan.amount / basePrice) * 100) : 0;
          return (
            <div key={plan.months} className="rounded-xl border bg-card p-4 shadow-sm text-center relative overflow-hidden">
              {discount > 0 && (
                <span className="absolute top-2 right-2 bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full">
                  -{discount}%
                </span>
              )}
              <p className="text-lg font-bold">{plan.label}</p>
              {discount > 0 && (
                <p className="text-xs text-muted-foreground line-through mt-1">{formatVND(basePrice)}</p>
              )}
              <p className="text-2xl font-bold text-emerald-600 mt-0.5">{formatVND(plan.amount)}</p>
              <p className="text-xs text-muted-foreground mt-1">
                {formatVND(Math.round(plan.amount / plan.months))}/tháng
              </p>
            </div>
          );
        })}
      </div>

      {/* Plan Manager (owner only) */}
      {isOwner && showPlanManager && (
        <div className="mb-6 rounded-lg border border-dashed border-primary/30 bg-primary/5 p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-base font-semibold">Cài đặt gói dịch vụ</h2>
            <Button size="sm" onClick={openCreatePlan}>Thêm gói mới</Button>
          </div>
          <div className="rounded-lg border shadow-sm bg-card">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tên gói</TableHead>
                  <TableHead>Số tháng</TableHead>
                  <TableHead>Giá</TableHead>
                  <TableHead>Giá/tháng</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="text-right">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {plansLoading ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
                  </TableRow>
                ) : allPlans.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Chưa có gói dịch vụ nào</TableCell>
                  </TableRow>
                ) : (
                  allPlans.map((plan) => (
                    <TableRow key={plan.id}>
                      <TableCell className="font-medium">{plan.label}</TableCell>
                      <TableCell>{plan.months} tháng</TableCell>
                      <TableCell className="font-semibold text-emerald-600">{formatVND(plan.amount)}</TableCell>
                      <TableCell className="text-muted-foreground">{formatVND(Math.round(plan.amount / plan.months))}</TableCell>
                      <TableCell>
                        {plan.is_active ? (
                          <Badge variant="default">Hoạt động</Badge>
                        ) : (
                          <Badge variant="secondary">Tạm ẩn</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex gap-2 justify-end">
                          <Button size="sm" variant="outline" onClick={() => openEditPlan(plan)}>Sửa</Button>
                          <Button size="sm" variant="destructive" onClick={() => setDeleteTarget(plan)}>Xóa</Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      <div className="flex gap-3 mb-4">
        <Input
          placeholder="Tìm thành viên theo SĐT hoặc tên..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          className="max-w-sm shadow-sm"
        />
      </div>

      <div className="rounded-lg border shadow-sm bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Thành viên</TableHead>
              <TableHead>SĐT</TableHead>
              <TableHead>Vai trò</TableHead>
              <TableHead>Địa chỉ</TableHead>
              <TableHead>Gói hiện tại</TableHead>
              <TableHead>Hết hạn</TableHead>
              <TableHead>Trạng thái</TableHead>
              <TableHead className="text-right">Thao tác</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
              </TableRow>
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Không tìm thấy thành viên</TableCell>
              </TableRow>
            ) : (
              users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-8 w-8">
                        <AvatarImage src={user.avatar_url} alt={user.name || user.phone} />
                        <AvatarFallback className="text-xs">
                          {(user.name || user.phone || "?").charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <span className="font-medium text-sm">{user.name || "-"}</span>
                    </div>
                  </TableCell>
                  <TableCell className="font-mono text-sm">{user.phone}</TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border ${user.role === "owner" ? "bg-amber-50 text-amber-700 border-amber-200" : user.role === "admin" ? "bg-purple-50 text-purple-700 border-purple-200" : user.role === "editor" ? "bg-cyan-50 text-cyan-700 border-cyan-200" : "bg-violet-50 text-violet-700 border-violet-200"}`}>
                      {user.role === "owner" ? "Chủ sở hữu" : user.role === "admin" ? "Quản trị viên" : user.role === "editor" ? "Biên tập viên" : "Thành viên"}
                    </span>
                  </TableCell>
                  <TableCell className="text-sm">
                    {[user.ward, user.province].filter(Boolean).join(", ") || "-"}
                  </TableCell>
                  <TableCell>
                    {user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date() ? (
                      <Badge variant="default">Đang hoạt động</Badge>
                    ) : (
                      <Badge variant="secondary">Chưa có gói</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-sm">
                    {user.subscription_expires_at ? (
                      <span className={new Date(user.subscription_expires_at) > new Date() ? "text-emerald-600 font-medium" : "text-red-600"}>
                        {new Date(user.subscription_expires_at).toLocaleDateString("vi-VN")}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>
                  <TableCell>
                    {user.is_blocked ? (
                      <Badge variant="destructive">Đã khóa</Badge>
                    ) : (
                      <Badge variant="default">Hoạt động</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button size="sm" onClick={() => { setActivateDialog(user); setSelectedMonths(1); }}>
                      Gia hạn
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
        <span className="text-sm text-muted-foreground">
          Trang {page} / {totalPages || 1} ({total} thành viên)
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

      {/* Activate Dialog */}
      <Dialog open={!!activateDialog} onOpenChange={() => { setActivateDialog(null); setActivateError(""); }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Gia hạn gói dịch vụ</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Thành viên: <strong>{activateDialog?.name || activateDialog?.phone}</strong>
          </p>

          {/* Plan selection */}
          <div>
            <label className="text-sm font-medium mb-2 block">Chọn gói đăng ký</label>
            <div className="grid grid-cols-2 gap-2">
              {plans.map((plan) => {
                const basePrice = (plans.find(p => p.months === 1)?.amount || plan.amount / plan.months) * plan.months;
                const discount = basePrice > 0 ? Math.round((1 - plan.amount / basePrice) * 100) : 0;
                return (
                  <button
                    key={plan.months}
                    onClick={() => setSelectedMonths(plan.months)}
                    className={cn(
                      "rounded-lg border p-3 text-left transition-all relative overflow-hidden",
                      selectedMonths === plan.months
                        ? "border-primary bg-primary/5 ring-2 ring-primary/20"
                        : "border-border hover:border-primary/30 hover:bg-muted/50"
                    )}
                  >
                    {discount > 0 && (
                      <span className="absolute top-1 right-1 bg-red-500 text-white text-[9px] font-bold px-1 py-0.5 rounded-full">
                        -{discount}%
                      </span>
                    )}
                    <p className="text-sm font-semibold">{plan.label}</p>
                    {discount > 0 && (
                      <p className="text-[11px] text-muted-foreground line-through">{formatVND(basePrice)}</p>
                    )}
                    <p className={cn(
                      "text-lg font-bold mt-0.5",
                      selectedMonths === plan.months ? "text-primary" : "text-emerald-600"
                    )}>
                      {formatVND(plan.amount)}
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      {formatVND(Math.round(plan.amount / plan.months))}/tháng
                    </p>
                  </button>
                );
              })}
            </div>
          </div>

          {selectedPlan && (
            <div className="rounded-lg bg-emerald-50 dark:bg-emerald-950/20 p-3 text-sm">
              <p>Gói: <strong>{selectedPlan.label}</strong></p>
              <p>Thanh toán: <strong className="text-emerald-700">{formatVND(selectedPlan.amount)}</strong></p>
              <p>Thời hạn thêm: <strong>{selectedPlan.months * 30} ngày</strong></p>
            </div>
          )}

          {activateError && (
            <p className="text-sm text-destructive">{activateError}</p>
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setActivateDialog(null)}>Hủy</Button>
            <Button onClick={handleActivate} disabled={activating}>
              {activating ? "Đang xử lý..." : "Kích hoạt"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Plan Create/Edit Dialog */}
      <Dialog open={planDialogOpen} onOpenChange={(open) => { if (!open) setPlanDialogOpen(false); }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingPlan ? "Chỉnh sửa gói" : "Thêm gói mới"}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Tên gói</label>
              <Input
                value={planForm.label}
                onChange={(e) => setPlanForm({ ...planForm, label: e.target.value })}
                placeholder="VD: 1 tháng, 3 tháng..."
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium mb-1 block">Số tháng</label>
                <Input
                  type="number"
                  min={1}
                  value={planForm.months}
                  onChange={(e) => setPlanForm({ ...planForm, months: parseInt(e.target.value) || 0 })}
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Giá (VNĐ)</label>
                <Input
                  type="number"
                  min={0}
                  value={planForm.amount}
                  onChange={(e) => setPlanForm({ ...planForm, amount: parseInt(e.target.value) || 0 })}
                />
              </div>
            </div>
            {editingPlan && (
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={planForm.is_active ?? true}
                  onChange={(e) => setPlanForm({ ...planForm, is_active: e.target.checked })}
                  className="rounded border-gray-300"
                />
                <label htmlFor="is_active" className="text-sm">Đang hoạt động</label>
              </div>
            )}
            {planForm.months > 0 && planForm.amount > 0 && (
              <div className="rounded-lg bg-muted/50 p-3 text-sm">
                <p>Giá mỗi tháng: <strong className="text-emerald-600">{formatVND(Math.round(planForm.amount / planForm.months))}</strong></p>
              </div>
            )}
            {planError && <p className="text-sm text-destructive">{planError}</p>}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setPlanDialogOpen(false)}>Hủy</Button>
            <Button onClick={handleSavePlan} disabled={planSaving}>
              {planSaving ? "Đang lưu..." : editingPlan ? "Cập nhật" : "Tạo gói"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Plan Delete Confirmation */}
      <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Xác nhận xóa</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Bạn có chắc chắn muốn xóa gói <strong>&quot;{deleteTarget?.label}&quot;</strong>? Thao tác này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteTarget(null)}>Hủy</Button>
            <Button variant="destructive" onClick={handleDeletePlan} disabled={deleting}>
              {deleting ? "Đang xóa..." : "Xóa"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
