"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArrowLeft, ChevronDown } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  getUserDetail, blockUser, unblockUser, activateSubscription, listUserListings, listUserSubscriptions, changeUserRole,
  type User, type Listing, type Subscription,
} from "@/services/api";

export default function UserDetailPage() {
  const { token, user: currentUser } = useAuth();
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [user, setUser] = useState<User | null>(null);
  const [listings, setListings] = useState<Listing[]>([]);
  const [listingsTotal, setListingsTotal] = useState(0);
  const [listingsPage, setListingsPage] = useState(1);
  const listingsLimit = 10;

  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [subsTotal, setSubsTotal] = useState(0);
  const [subsPage, setSubsPage] = useState(1);
  const subsLimit = 10;

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Block dialog
  const [blockDialog, setBlockDialog] = useState(false);
  const [blockReason, setBlockReason] = useState("");

  // Subscription dialog
  const [subDialog, setSubDialog] = useState(false);
  const [subDays, setSubDays] = useState("30");

  const fetchUser = useCallback(async () => {
    if (!token || !id) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getUserDetail(token, id);
      setUser(data);
    } catch {
      setError("Không tìm thấy người dùng.");
    } finally {
      setLoading(false);
    }
  }, [token, id]);

  const fetchListings = useCallback(async () => {
    if (!token) return;
    try {
      const res = await listUserListings(token, id, listingsPage, listingsLimit);
      setListings(res.data || []);
      setListingsTotal(res.total);
    } catch {
      // Non-critical
    }
  }, [token, id, listingsPage]);

  const fetchSubscriptions = useCallback(async () => {
    if (!token) return;
    try {
      const res = await listUserSubscriptions(token, id, subsPage, subsLimit);
      setSubscriptions(res.data || []);
      setSubsTotal(res.total);
    } catch {
      // Non-critical
    }
  }, [token, id, subsPage]);

  useEffect(() => {
    fetchUser();
    fetchListings();
    fetchSubscriptions();
  }, [fetchUser, fetchListings, fetchSubscriptions]);

  async function handleBlock() {
    if (!token || !blockReason) return;
    try {
      await blockUser(token, id, blockReason);
      toast.success("Đã khóa tài khoản");
      setBlockDialog(false);
      setBlockReason("");
      fetchUser();
    } catch {
      toast.error("Khóa tài khoản thất bại");
    }
  }

  async function handleUnblock() {
    if (!token) return;
    try {
      await unblockUser(token, id);
      toast.success("Đã mở khóa tài khoản");
      fetchUser();
    } catch {
      toast.error("Mở khóa thất bại");
    }
  }

  // Role change confirmation
  const [roleChangeDialog, setRoleChangeDialog] = useState<string | null>(null);

  const [subError, setSubError] = useState("");

  async function handleActivateSub() {
    if (!token) return;
    setSubError("");
    const d = parseInt(subDays);
    if (!d || d <= 0 || d > 365) {
      setSubError("Số ngày phải từ 1 đến 365");
      return;
    }
    try {
      await activateSubscription(token, id, d);
      toast.success("Đã kích hoạt gói dịch vụ");
      setSubDialog(false);
      setSubDays("30");
      fetchUser();
      fetchSubscriptions();
    } catch (err) {
      setSubError(err instanceof Error ? err.message : "Không thể kích hoạt gói dịch vụ");
    }
  }

  const allRoles = [
    { value: "owner", label: "Chủ sở hữu", color: "bg-amber-50 text-amber-700 border-amber-200" },
    { value: "admin", label: "Quản trị viên", color: "bg-purple-50 text-purple-700 border-purple-200" },
    { value: "editor", label: "Biên tập viên", color: "bg-cyan-50 text-cyan-700 border-cyan-200" },
    { value: "member", label: "Thành viên", color: "bg-violet-50 text-violet-700 border-violet-200" },
  ];

  // Only owner can assign owner/admin roles
  const roleOptions = currentUser?.role === "owner"
    ? allRoles
    : allRoles.filter((r) => r.value !== "owner");

  function roleLabel(role: string) {
    return allRoles.find((r) => r.value === role)?.label || role;
  }

  function roleColor(role: string) {
    return allRoles.find((r) => r.value === role)?.color || "bg-violet-50 text-violet-700 border-violet-200";
  }

  function requestChangeRole(newRole: string) {
    if (!user || user.role === newRole) return;
    setRoleChangeDialog(newRole);
  }

  async function confirmChangeRole() {
    if (!token || !roleChangeDialog) return;
    try {
      const updated = await changeUserRole(token, id, roleChangeDialog);
      setUser(updated);
      toast.success("Đã thay đổi vai trò");
    } catch {
      toast.error("Thay đổi vai trò thất bại");
    } finally {
      setRoleChangeDialog(null);
    }
  }

  function formatPrice(price: number) {
    return new Intl.NumberFormat("vi-VN").format(price) + "đ/kg";
  }

  function formatQty(qty: number) {
    return new Intl.NumberFormat("vi-VN").format(qty) + " kg";
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-muted-foreground">
        Đang tải...
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" onClick={() => router.push("/users")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Quay lại
        </Button>
        <p className="text-center py-20 text-muted-foreground">{error || "Không tìm thấy người dùng."}</p>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => router.push("/users")}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Quay lại
          </Button>
          <h1 className="text-xl font-semibold">Chi tiết người dùng</h1>
        </div>
        <div className="flex gap-2">
          {currentUser?.id !== user.id && (
            user.is_blocked ? (
              <Button variant="outline" size="sm" onClick={handleUnblock}>
                Mở khóa
              </Button>
            ) : (
              <Button variant="destructive" size="sm" onClick={() => setBlockDialog(true)}>
                Khóa tài khoản
              </Button>
            )
          )}
          {!["owner", "admin"].includes(user.role) && (
            <Button variant="outline" size="sm" onClick={() => setSubDialog(true)}>
              Kích hoạt gói
            </Button>
          )}
        </div>
      </div>

      {/* Profile card - always visible */}
      <div className="rounded-lg border shadow-sm bg-card p-5">
        <div className="flex items-center gap-4">
          <Avatar className="h-18 w-18">
            <AvatarImage src={user.avatar_url} alt={user.name || user.phone} />
            <AvatarFallback className="text-lg">
              {(user.name || user.phone || "?").charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1">
            <h2 className="text-base font-semibold leading-tight">{user.name || "-"}</h2>
            <p className="font-mono text-sm text-muted-foreground leading-tight">{user.phone}</p>
            <div className="flex gap-2 mt-1.5">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium border cursor-pointer hover:opacity-80 transition-opacity ${roleColor(user.role)}`}>
                    {roleLabel(user.role)}
                    <ChevronDown className="h-3 w-3" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start">
                  {roleOptions.map((opt) => (
                    <DropdownMenuItem
                      key={opt.value}
                      onClick={() => requestChangeRole(opt.value)}
                      className={user.role === opt.value ? "font-semibold" : ""}
                    >
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium border ${opt.color} mr-2`}>
                        {opt.label}
                      </span>
                      {user.role === opt.value && "✓"}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              {user.is_blocked ? (
                <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border bg-red-50 text-red-700 border-red-200">Đã khóa</span>
              ) : (
                <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border bg-emerald-50 text-emerald-700 border-emerald-200">Hoạt động</span>
              )}
            </div>
          </div>
          {/* Subscription status summary */}
          {!["owner", "admin"].includes(user.role) && (
            <div className={`rounded-lg border px-4 py-3 text-right ${user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date() ? "bg-emerald-50 border-emerald-200" : "bg-amber-50 border-amber-200"}`}>
              <p className="text-xs text-muted-foreground leading-none">Gói dịch vụ</p>
              <p className="text-sm font-semibold leading-tight">
                {user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date()
                  ? "Đang hoạt động"
                  : "Chưa kích hoạt"}
              </p>
              {user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date() && (
                <p className="text-xs text-muted-foreground mt-0.5">
                  Hết hạn: {new Date(user.subscription_expires_at).toLocaleDateString("vi-VN")}
                </p>
              )}
            </div>
          )}
        </div>

        {user.is_blocked && user.block_reason && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm mt-3">
            <span className="text-muted-foreground">Lý do khóa: </span>
            <span className="font-medium leading-tight">{user.block_reason}</span>
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultValue="info" className="space-y-3">
        <TabsList>
          <TabsTrigger value="info">Thông tin tài khoản</TabsTrigger>
          <TabsTrigger value="listings">Sản phẩm đã đăng ({listingsTotal})</TabsTrigger>
          <TabsTrigger value="subscriptions">Lịch sử gói ({subsTotal})</TabsTrigger>
        </TabsList>

        {/* Tab 1: Thông tin tài khoản */}
        <TabsContent value="info" className="space-y-2">
          <div className="rounded-lg border shadow-sm bg-card p-3">
            <h3 className="text-base font-semibold mb-1">Thông tin cá nhân</h3>
            <div className="grid grid-cols-2 gap-1 text-sm">
              {user.org_name && (
                <div className="rounded-lg bg-muted/50 px-2 py-1">
                  <span className="text-muted-foreground block leading-none text-xs">Tổ chức</span>
                  <span className="font-medium leading-tight">{user.org_name}</span>
                </div>
              )}
              <div className="rounded-lg bg-muted/50 px-2 py-1">
                <span className="text-muted-foreground block leading-none text-xs">Tỉnh/TP</span>
                <span className="font-medium leading-tight">{user.province || "-"}</span>
              </div>
              <div className="rounded-lg bg-muted/50 px-2 py-1">
                <span className="text-muted-foreground block leading-none text-xs">Xã/Phường</span>
                <span className="font-medium leading-tight">{user.ward || "-"}</span>
              </div>
              {user.address && (
                <div className="rounded-lg bg-muted/50 px-2 py-1">
                  <span className="text-muted-foreground block leading-none text-xs">Địa chỉ chi tiết</span>
                  <span className="font-medium leading-tight">{user.address}</span>
                </div>
              )}
              <div className="rounded-lg bg-muted/50 px-2 py-1">
                <span className="text-muted-foreground block leading-none text-xs">Ngày tạo</span>
                <span className="font-medium leading-tight">{new Date(user.created_at).toLocaleString("vi-VN")}</span>
              </div>
              {!["owner", "admin"].includes(user.role) && (
                <div className="rounded-lg bg-muted/50 px-2 py-1">
                  <span className="text-muted-foreground block leading-none text-xs">Hết hạn gói</span>
                  {user.subscription_expires_at ? (
                    <span className={`font-medium ${new Date(user.subscription_expires_at) > new Date() ? "text-emerald-600" : "text-red-600"}`}>
                      {new Date(user.subscription_expires_at).toLocaleString("vi-VN")}
                    </span>
                  ) : (
                    <span className="font-medium text-muted-foreground">Chưa có gói</span>
                  )}
                </div>
              )}
            </div>

            {user.description && (
              <div className="mt-1">
                <span className="text-sm text-muted-foreground">Giới thiệu</span>
                <p className="mt-0.5 text-sm whitespace-pre-wrap bg-muted/50 rounded-md px-2 py-1">
                  {user.description}
                </p>
              </div>
            )}
          </div>

          {/* System metadata */}
          <div className="rounded-lg border shadow-sm bg-card p-2.5 text-xs text-muted-foreground space-y-0.5">
            <h3 className="text-sm font-medium mb-1">Thông tin hệ thống</h3>
            <p>ID: <span className="font-mono">{user.id}</span></p>
            <p>Tạo lúc: {new Date(user.created_at).toLocaleString("vi-VN")}</p>
          </div>
        </TabsContent>

        {/* Tab 2: Sản phẩm đã đăng */}
        <TabsContent value="listings">
          <div className="rounded-lg border shadow-sm bg-card p-4">
            {listings.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">Chưa có tin đăng nào</p>
            ) : (
              <div className="space-y-2">
                {listings.map((listing) => (
                  <div
                    key={listing.id}
                    className="flex items-center justify-between rounded-md border p-3 cursor-pointer hover:bg-muted/50"
                    onClick={() => router.push(`/listings/${listing.id}`)}
                  >
                    <div className="flex items-center gap-3">
                      {listing.images && listing.images.length > 0 ? (
                        <img
                          src={listing.images[0]}
                          alt={listing.title}
                          className="h-10 w-10 rounded object-cover border"
                        />
                      ) : (
                        <div className="h-10 w-10 rounded bg-muted flex items-center justify-center text-xs text-muted-foreground">
                          ?
                        </div>
                      )}
                      <div>
                        <p className="text-sm font-medium">{listing.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {listing.rice_type} &middot; {formatQty(listing.quantity_kg)} &middot; {formatPrice(listing.price_per_kg)}
                        </p>
                      </div>
                    </div>
                    <Badge variant={listing.status === "active" ? "default" : "secondary"} className="text-xs">
                      {listing.status === "active" ? "Đang hiển thị" : listing.status}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
            {listingsTotal > listingsLimit && (
              <div className="flex items-center justify-between mt-3 pt-3 border-t">
                <span className="text-xs text-muted-foreground">
                  Trang {listingsPage} / {Math.ceil(listingsTotal / listingsLimit)} ({listingsTotal} tin đăng)
                </span>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" disabled={listingsPage <= 1} onClick={() => setListingsPage(listingsPage - 1)}>
                    Trước
                  </Button>
                  <Button size="sm" variant="outline" disabled={listingsPage >= Math.ceil(listingsTotal / listingsLimit)} onClick={() => setListingsPage(listingsPage + 1)}>
                    Sau
                  </Button>
                </div>
              </div>
            )}
          </div>
        </TabsContent>

        {/* Tab 3: Lịch sử gói */}
        <TabsContent value="subscriptions">
          <div className="rounded-lg border shadow-sm bg-card p-4">
            {subscriptions.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">Chưa có gói dịch vụ nào</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Gói</TableHead>
                    <TableHead>Ngày bắt đầu</TableHead>
                    <TableHead>Ngày hết hạn</TableHead>
                    <TableHead>Trạng thái</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {subscriptions.map((sub) => (
                    <TableRow key={sub.id}>
                      <TableCell className="text-sm font-medium">
                        {sub.plan === "paid" ? "Trả phí" : sub.plan === "free_trial" ? "Dùng thử" : sub.plan}
                      </TableCell>
                      <TableCell className="text-sm">
                        {new Date(sub.started_at).toLocaleString("vi-VN")}
                      </TableCell>
                      <TableCell className="text-sm">
                        {new Date(sub.expires_at).toLocaleString("vi-VN")}
                      </TableCell>
                      <TableCell>
                        <Badge variant={sub.status === "active" ? "default" : "secondary"}>
                          {sub.status === "active" ? "Đang hoạt động" : "Hết hạn"}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
            {subsTotal > subsLimit && (
              <div className="flex items-center justify-between mt-3 pt-3 border-t">
                <span className="text-xs text-muted-foreground">
                  Trang {subsPage} / {Math.ceil(subsTotal / subsLimit)} ({subsTotal} gói)
                </span>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" disabled={subsPage <= 1} onClick={() => setSubsPage(subsPage - 1)}>
                    Trước
                  </Button>
                  <Button size="sm" variant="outline" disabled={subsPage >= Math.ceil(subsTotal / subsLimit)} onClick={() => setSubsPage(subsPage + 1)}>
                    Sau
                  </Button>
                </div>
              </div>
            )}
          </div>
        </TabsContent>
      </Tabs>

      {/* Block Dialog */}
      <Dialog open={blockDialog} onOpenChange={setBlockDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Khóa người dùng: {user.phone}</DialogTitle>
          </DialogHeader>
          <Input
            placeholder="Lý do khóa..."
            value={blockReason}
            onChange={(e) => setBlockReason(e.target.value)}
          />
          <DialogFooter>
            <Button variant="ghost" onClick={() => setBlockDialog(false)}>Hủy</Button>
            <Button variant="destructive" onClick={handleBlock} disabled={!blockReason}>
              Khóa người dùng
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Subscription Dialog */}
      <Dialog open={subDialog} onOpenChange={(open) => { setSubDialog(open); if (!open) setSubError(""); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Kích hoạt gói dịch vụ: {user.phone}</DialogTitle>
          </DialogHeader>
          <Input
            type="number"
            placeholder="Số ngày (tối đa 365)"
            value={subDays}
            onChange={(e) => setSubDays(e.target.value)}
            min={1}
            max={365}
          />
          {subError && (
            <p className="text-sm text-destructive">{subError}</p>
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setSubDialog(false)}>Hủy</Button>
            <Button onClick={handleActivateSub}>Kích hoạt</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Role Change Confirmation Dialog */}
      <Dialog open={!!roleChangeDialog} onOpenChange={() => setRoleChangeDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Thay đổi vai trò</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Bạn có chắc muốn đổi vai trò của <strong>{user.name || user.phone}</strong> từ{" "}
            <strong>{roleLabel(user.role)}</strong> thành{" "}
            <strong>{roleChangeDialog ? roleLabel(roleChangeDialog) : ""}</strong>?
          </p>
          {(roleChangeDialog === "owner" || roleChangeDialog === "admin") && (
            <p className="text-sm text-amber-600 font-medium">
              {roleChangeDialog === "owner"
                ? "Vai trò Chủ sở hữu có quyền cao nhất trên hệ thống. Hãy cân nhắc kỹ."
                : "Vai trò Quản trị viên có toàn quyền trên hệ thống. Hãy cân nhắc kỹ."}
            </p>
          )}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setRoleChangeDialog(null)}>Hủy</Button>
            <Button onClick={confirmChangeRole}>Xác nhận</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
