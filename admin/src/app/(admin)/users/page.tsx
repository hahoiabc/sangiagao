"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { MoreHorizontal, Eye, ShieldBan, ShieldCheck, CreditCard, UserCog, Users, Shield, Check, X, Lock } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { listUsers, blockUser, unblockUser, activateSubscription, changeUserRole, getPermissions, savePermissions, type User, type PermissionMatrix } from "@/services/api";
import { cn } from "@/lib/utils";

// ── Roles ──
const rolesData = [
  { key: "all", label: "Tất cả", color: "", description: "" },
  { key: "owner", label: "Chủ sở hữu", color: "from-amber-500 to-orange-500", description: "Quyền cao nhất. Quản lý toàn bộ hệ thống, phân quyền admin, xóa/chỉnh sửa mọi dữ liệu." },
  { key: "admin", label: "Quản trị viên", color: "from-red-500 to-pink-500", description: "Quản lý người dùng, tin đăng, báo cáo, gói dịch vụ, danh mục sản phẩm." },
  { key: "editor", label: "Biên tập viên", color: "from-indigo-500 to-purple-500", description: "Duyệt và chỉnh sửa tin đăng, quản lý danh mục sản phẩm, xử lý báo cáo vi phạm." },
  { key: "member", label: "Thành viên", color: "from-emerald-500 to-teal-500", description: "Người dùng đã xác thực. Có thể đăng tin mua/bán, nhắn tin, đánh giá." },
];

// ── Permission matrix ──
const permissionRoles = ["owner", "admin", "editor", "member", "expired", "guest"] as const;

type PermissionGroup = {
  group: string;
  permissions: {
    key: string;
    label: string;
    // Which roles have this permission by default
    defaults: Record<string, boolean>;
    // Owner-only permission that cannot be changed
    locked?: boolean;
  }[];
};

const permissionGroups: PermissionGroup[] = [
  {
    group: "Tổng quan & Giám sát",
    permissions: [
      { key: "dashboard.view", label: "Xem tổng quan (Dashboard)", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "dashboard.charts", label: "Xem biểu đồ thống kê", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "system.monitor", label: "Giám sát hệ thống", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
    ],
  },
  {
    group: "Người dùng",
    permissions: [
      { key: "users.list", label: "Xem danh sách người dùng", defaults: { owner: true, admin: true, editor: false, member: false, expired: false, guest: false } },
      { key: "users.detail", label: "Xem chi tiết người dùng", defaults: { owner: true, admin: true, editor: false, member: false, expired: false, guest: false } },
      { key: "users.block", label: "Khóa / mở khóa tài khoản", defaults: { owner: true, admin: true, editor: false, member: false, expired: false, guest: false } },
      { key: "users.role", label: "Đổi vai trò người dùng", defaults: { owner: true, admin: true, editor: false, member: false, expired: false, guest: false } },
      { key: "users.batch_block", label: "Khóa hàng loạt", defaults: { owner: true, admin: true, editor: false, member: false, expired: false, guest: false } },
    ],
  },
  {
    group: "Tin đăng",
    permissions: [
      { key: "listings.create", label: "Đăng tin mới", defaults: { owner: true, admin: true, editor: true, member: true, expired: false, guest: false } },
      { key: "listings.edit_own", label: "Sửa tin của mình", defaults: { owner: true, admin: true, editor: true, member: true, expired: false, guest: false } },
      { key: "listings.delete_any", label: "Xóa tin của người khác", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "listings.batch_delete", label: "Xóa tin hàng loạt", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
    ],
  },
  {
    group: "Sàn gạo",
    permissions: [
      { key: "marketplace.browse", label: "Xem sàn gạo", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: true } },
      { key: "marketplace.search", label: "Tìm kiếm tin đăng", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: true } },
      { key: "marketplace.detail", label: "Xem chi tiết tin đăng", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: true } },
      { key: "marketplace.priceboard", label: "Xem bảng giá", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: true } },
      { key: "marketplace.seller_profile", label: "Xem hồ sơ người bán", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: true } },
    ],
  },
  {
    group: "Tin nhắn & Đánh giá",
    permissions: [
      { key: "chat.send", label: "Gửi / nhận tin nhắn", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: false } },
      { key: "chat.send_image", label: "Gửi ảnh trong tin nhắn", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: false } },
      { key: "ratings.create", label: "Đánh giá người bán", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: false } },
    ],
  },
  {
    group: "Báo cáo vi phạm",
    permissions: [
      { key: "reports.create", label: "Tạo báo cáo vi phạm", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: false } },
      { key: "reports.manage", label: "Xử lý báo cáo vi phạm", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
    ],
  },
  {
    group: "Gói dịch vụ",
    permissions: [
      { key: "sub.activate", label: "Kích hoạt gói cho người dùng", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "sub.revenue", label: "Xem thống kê doanh thu", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "sub.plans", label: "Quản lý gói dịch vụ (CRUD)", defaults: { owner: true, admin: false, editor: false, member: false, expired: false, guest: false }, locked: true },
    ],
  },
  {
    group: "Danh mục & Tài trợ",
    permissions: [
      { key: "catalog.manage", label: "Quản lý danh mục sản phẩm", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
      { key: "sponsors.manage", label: "Quản lý tài trợ", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
    ],
  },
  {
    group: "Góp ý",
    permissions: [
      { key: "feedback.create", label: "Gửi góp ý", defaults: { owner: true, admin: true, editor: true, member: true, expired: true, guest: false } },
      { key: "feedback.reply", label: "Trả lời góp ý", defaults: { owner: true, admin: true, editor: true, member: false, expired: false, guest: false } },
    ],
  },
];

// ── Role Permission Tab component ──
function RolePermissionsTab() {
  const { token } = useAuth();

  // Build initial state from defaults
  const buildInitialPerms = (): PermissionMatrix => {
    const perms: PermissionMatrix = {};
    for (const role of permissionRoles) {
      perms[role] = {};
      for (const g of permissionGroups) {
        for (const p of g.permissions) {
          perms[role][p.key] = p.defaults[role] ?? false;
        }
      }
    }
    return perms;
  };

  const [perms, setPerms] = useState<PermissionMatrix>(buildInitialPerms);
  const [savedPerms, setSavedPerms] = useState<PermissionMatrix>(buildInitialPerms);
  const [hasChanges, setHasChanges] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Load permissions from API
  useEffect(() => {
    if (!token) return;
    setLoading(true);
    getPermissions(token)
      .then((matrix) => {
        // Merge API data with defaults (ensure all keys exist)
        const merged = buildInitialPerms();
        for (const role of permissionRoles) {
          if (matrix[role]) {
            for (const key of Object.keys(matrix[role])) {
              merged[role][key] = matrix[role][key];
            }
          }
        }
        setPerms(merged);
        setSavedPerms(merged);
      })
      .catch(() => {
        toast.error("Không thể tải cấu hình quyền hạn");
      })
      .finally(() => setLoading(false));
  }, [token]);

  function togglePerm(role: string, key: string) {
    if (role === "owner" || role === "guest") return;
    const perm = permissionGroups.flatMap(g => g.permissions).find(p => p.key === key);
    if (perm?.locked) return;

    setPerms(prev => {
      const updated = { ...prev, [role]: { ...prev[role], [key]: !prev[role][key] } };
      return updated;
    });
    setHasChanges(true);
  }

  async function handleSave() {
    if (!token) return;
    setSaving(true);
    try {
      await savePermissions(token, perms);
      setSavedPerms(perms);
      toast.success("Đã lưu cấu hình quyền hạn");
      setHasChanges(false);
    } catch {
      toast.error("Không thể lưu cấu hình quyền hạn");
    } finally {
      setSaving(false);
    }
  }

  function handleReset() {
    setPerms(savedPerms);
    setHasChanges(false);
  }

  const roleLabels: Record<string, string> = {
    owner: "Chủ sở hữu",
    admin: "QTV",
    editor: "BTV",
    member: "Thành viên",
    expired: "Hết hạn",
    guest: "Khách vãng lai",
  };

  const roleColors: Record<string, string> = {
    owner: "bg-amber-500",
    admin: "bg-red-500",
    editor: "bg-indigo-500",
    member: "bg-emerald-500",
    expired: "bg-gray-400",
    guest: "bg-slate-400",
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <p className="text-sm text-muted-foreground">
          Cấu hình quyền hạn cho từng vai trò. Chủ sở hữu luôn có toàn quyền.
        </p>
        <div className="flex gap-2">
          {hasChanges && (
            <Button size="sm" variant="outline" onClick={handleReset}>Hoàn tác</Button>
          )}
          <Button size="sm" onClick={handleSave} disabled={!hasChanges || saving}>
            {saving ? "Đang lưu..." : "Lưu thay đổi"}
          </Button>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mr-3" />
          Đang tải cấu hình quyền hạn...
        </div>
      ) : (
      <div className="rounded-lg border shadow-sm bg-card overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/30">
              <th className="text-left px-4 py-3 font-semibold min-w-[240px]">Quyền hạn</th>
              {permissionRoles.map(role => (
                <th key={role} className="px-2 py-3 text-center min-w-[80px]">
                  <div className="flex flex-col items-center gap-1">
                    <span className={cn("inline-block w-2.5 h-2.5 rounded-full", roleColors[role])} />
                    <span className="text-xs font-semibold">{roleLabels[role]}</span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {permissionGroups.map((group) => (
              <>
                <tr key={group.group} className="bg-muted/20">
                  <td colSpan={permissionRoles.length + 1} className="px-4 py-2">
                    <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">{group.group}</span>
                  </td>
                </tr>
                {group.permissions.map((perm) => (
                  <tr key={perm.key} className="border-b border-border/50 hover:bg-muted/10 transition-colors">
                    <td className="px-4 py-2.5 text-sm">
                      <div className="flex items-center gap-1.5">
                        {perm.label}
                        {perm.locked && <Lock className="h-3 w-3 text-muted-foreground/50" />}
                      </div>
                    </td>
                    {permissionRoles.map(role => {
                      const checked = perms[role]?.[perm.key] ?? false;
                      const isOwner = role === "owner";
                      const isGuest = role === "guest";
                      const isLocked = perm.locked && role !== "owner";
                      const disabled = isOwner || isGuest || isLocked;

                      return (
                        <td key={role} className="px-2 py-2.5 text-center">
                          <button
                            onClick={() => togglePerm(role, perm.key)}
                            disabled={disabled}
                            className={cn(
                              "inline-flex items-center justify-center w-7 h-7 rounded-md transition-all",
                              checked
                                ? "bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400"
                                : "bg-muted/50 text-muted-foreground/30",
                              disabled
                                ? "cursor-not-allowed opacity-60"
                                : "cursor-pointer hover:ring-2 hover:ring-primary/20"
                            )}
                          >
                            {checked ? <Check className="h-4 w-4" /> : <X className="h-3.5 w-3.5" />}
                          </button>
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </>
            ))}
          </tbody>
        </table>
      </div>
      )}
    </div>
  );
}

// ── Main page ──
export default function UsersPage() {
  const { token, user: currentUser } = useAuth();
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<"users" | "roles">("users");
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [roleFilter, setRoleFilter] = useState("all");

  // Block dialog
  const [blockDialog, setBlockDialog] = useState<User | null>(null);
  const [blockReason, setBlockReason] = useState("");

  // Subscription dialog
  const [subDialog, setSubDialog] = useState<User | null>(null);
  const [subDays, setSubDays] = useState("30");
  const [subError, setSubError] = useState("");

  // Change role dialog
  const [roleDialog, setRoleDialog] = useState<User | null>(null);
  const [newRole, setNewRole] = useState("");

  const limit = 20;

  const fetchUsers = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await listUsers(token, search, page, limit);
      setUsers(res.data);
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token, search, page]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  async function handleBlock() {
    if (!token || !blockDialog || !blockReason) return;
    try {
      await blockUser(token, blockDialog.id, blockReason);
      toast.success("Đã khóa tài khoản");
      setBlockDialog(null);
      setBlockReason("");
      fetchUsers();
    } catch {
      toast.error("Khóa tài khoản thất bại");
    }
  }

  async function handleUnblock(userId: string) {
    if (!token) return;
    try {
      await unblockUser(token, userId);
      toast.success("Đã mở khóa tài khoản");
      fetchUsers();
    } catch {
      toast.error("Mở khóa thất bại");
    }
  }

  async function handleActivateSub() {
    if (!token || !subDialog) return;
    setSubError("");
    const d = parseInt(subDays);
    if (!d || d <= 0 || d > 365) {
      setSubError("Số ngày phải từ 1 đến 365");
      return;
    }
    try {
      await activateSubscription(token, subDialog.id, d);
      toast.success("Đã kích hoạt gói dịch vụ");
      setSubDialog(null);
      setSubDays("30");
      fetchUsers();
    } catch (err) {
      setSubError(err instanceof Error ? err.message : "Không thể kích hoạt gói dịch vụ");
    }
  }

  async function handleChangeRole() {
    if (!token || !roleDialog || !newRole) return;
    try {
      await changeUserRole(token, roleDialog.id, newRole);
      const roleLabel = rolesData.find(r => r.key === newRole)?.label;
      toast.success(`Đã đổi vai trò ${roleDialog.name || roleDialog.phone} thành ${roleLabel}`);
      setRoleDialog(null);
      setNewRole("");
      fetchUsers();
    } catch {
      toast.error("Đổi vai trò thất bại");
    }
  }

  const filteredUsers = roleFilter === "all" ? users : users.filter(u => u.role === roleFilter);
  const totalPages = Math.ceil(total / limit);
  const currentRoleInfo = rolesData.find(r => r.key === roleFilter);

  const roleCounts = rolesData.reduce((acc, r) => {
    acc[r.key] = r.key === "all" ? users.length : users.filter(u => u.role === r.key).length;
    return acc;
  }, {} as Record<string, number>);

  return (
    <div>
      {/* Page tabs */}
      <div className="flex items-center gap-1 mb-5 border-b">
        <button
          onClick={() => setActiveTab("users")}
          className={cn(
            "flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px",
            activeTab === "users"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30"
          )}
        >
          <Users className="h-4 w-4" />
          Quản lý người dùng
        </button>
        <button
          onClick={() => setActiveTab("roles")}
          className={cn(
            "flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px",
            activeTab === "roles"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30"
          )}
        >
          <Shield className="h-4 w-4" />
          Vai trò & Quyền hạn
        </button>
      </div>

      {/* ── Tab: Roles & Permissions ── */}
      {activeTab === "roles" && <RolePermissionsTab />}

      {/* ── Tab: User Management ── */}
      {activeTab === "users" && (
        <>
          {/* Role filter tabs */}
          <div className="flex gap-1.5 mb-4 flex-wrap">
            {rolesData.map(role => (
              <button
                key={role.key}
                onClick={() => setRoleFilter(role.key)}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors border",
                  roleFilter === role.key
                    ? "bg-primary text-primary-foreground border-primary"
                    : "bg-card text-muted-foreground border-border hover:bg-muted hover:text-foreground"
                )}
              >
                {role.label}
                {roleCounts[role.key] > 0 && (
                  <span className={cn(
                    "ml-0.5 inline-flex items-center justify-center rounded-full px-1.5 py-0.5 text-[10px] font-bold min-w-[18px]",
                    roleFilter === role.key ? "bg-white/20" : "bg-muted-foreground/10"
                  )}>
                    {roleCounts[role.key]}
                  </span>
                )}
              </button>
            ))}
          </div>

          {/* Role description */}
          {roleFilter !== "all" && currentRoleInfo?.description && (
            <div className={cn("rounded-lg overflow-hidden border shadow-sm mb-4")}>
              <div className={cn("px-4 py-3 bg-gradient-to-r text-white", currentRoleInfo.color)}>
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-bold">{currentRoleInfo.label}</h3>
                  <span className="ml-1 inline-flex items-center rounded-full bg-white/20 px-2 py-0.5 text-[10px] font-medium">
                    {roleCounts[roleFilter]} thành viên
                  </span>
                </div>
                <p className="text-xs mt-1 text-white/80">{currentRoleInfo.description}</p>
              </div>
            </div>
          )}

          <div className="flex gap-3 mb-4">
            <Input
              placeholder="Tìm theo SĐT hoặc tên..."
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1); }}
              className="max-w-sm shadow-sm"
            />
          </div>

          <div className="rounded-lg border shadow-sm bg-card overflow-x-auto">
            <Table className="min-w-[800px]">
              <TableHeader>
                <TableRow>
                  <TableHead>Thành viên</TableHead>
                  <TableHead>SĐT</TableHead>
                  <TableHead>Vai trò</TableHead>
                  <TableHead>Địa chỉ</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead>Hết hạn gói</TableHead>
                  <TableHead>Ngày tạo</TableHead>
                  <TableHead className="text-right">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
                  </TableRow>
                ) : filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Không tìm thấy người dùng</TableCell>
                  </TableRow>
                ) : (
                  filteredUsers.map((user) => (
                    <TableRow
                      key={user.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => router.push(`/users/${user.id}`)}
                    >
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
                        {user.is_blocked ? (
                          <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border bg-red-50 text-red-700 border-red-200">Đã khóa</span>
                        ) : (
                          <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border bg-emerald-50 text-emerald-700 border-emerald-200">Hoạt động</span>
                        )}
                      </TableCell>
                      <TableCell className="text-sm">
                        {user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date() ? (
                          <span className="text-emerald-600 font-medium">
                            {new Date(user.subscription_expires_at).toLocaleDateString("vi-VN")}
                          </span>
                        ) : (
                          <span className="text-muted-foreground">Chưa có gói</span>
                        )}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(user.created_at).toLocaleDateString("vi-VN")}
                      </TableCell>
                      <TableCell className="text-right" onClick={(e) => e.stopPropagation()}>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => router.push(`/users/${user.id}`)}>
                              <Eye className="h-4 w-4 mr-2" />
                              Xem chi tiết
                            </DropdownMenuItem>
                            {currentUser?.id !== user.id && (
                              <>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={() => { setRoleDialog(user); setNewRole(user.role); }}>
                                  <UserCog className="h-4 w-4 mr-2" />
                                  Đổi vai trò
                                </DropdownMenuItem>
                                {user.is_blocked ? (
                                  <DropdownMenuItem onClick={() => handleUnblock(user.id)}>
                                    <ShieldCheck className="h-4 w-4 mr-2" />
                                    Mở khóa
                                  </DropdownMenuItem>
                                ) : (
                                  <DropdownMenuItem onClick={() => setBlockDialog(user)} className="text-destructive focus:text-destructive">
                                    <ShieldBan className="h-4 w-4 mr-2" />
                                    Khóa tài khoản
                                  </DropdownMenuItem>
                                )}
                              </>
                            )}
                            {!["owner", "admin"].includes(user.role) && (
                              <>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={() => setSubDialog(user)}>
                                  <CreditCard className="h-4 w-4 mr-2" />
                                  {user.subscription_expires_at && new Date(user.subscription_expires_at) > new Date() ? "Gia hạn gói" : "Kích hoạt gói"}
                                </DropdownMenuItem>
                              </>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
            <span className="text-sm text-muted-foreground">
              Trang {page} / {totalPages || 1} ({total} người dùng)
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
        </>
      )}

      {/* Block Dialog */}
      <Dialog open={!!blockDialog} onOpenChange={() => setBlockDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Khóa người dùng: {blockDialog?.phone}</DialogTitle>
          </DialogHeader>
          <Input
            placeholder="Lý do khóa..."
            value={blockReason}
            onChange={(e) => setBlockReason(e.target.value)}
          />
          <DialogFooter>
            <Button variant="ghost" onClick={() => setBlockDialog(null)}>Hủy</Button>
            <Button variant="destructive" onClick={handleBlock} disabled={!blockReason}>
              Khóa người dùng
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Subscription Dialog */}
      <Dialog open={!!subDialog} onOpenChange={() => { setSubDialog(null); setSubError(""); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Kích hoạt gói dịch vụ: {subDialog?.phone}</DialogTitle>
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
            <Button variant="ghost" onClick={() => setSubDialog(null)}>Hủy</Button>
            <Button onClick={handleActivateSub}>Kích hoạt</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Change Role Dialog */}
      <Dialog open={!!roleDialog} onOpenChange={() => setRoleDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Đổi vai trò cho {roleDialog?.name || roleDialog?.phone}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Vai trò mới</label>
              <select
                className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={newRole}
                onChange={e => setNewRole(e.target.value)}
              >
                {rolesData.filter(r => r.key !== "all").map(role => (
                  <option key={role.key} value={role.key}>{role.label} — {role.description.slice(0, 50)}...</option>
                ))}
              </select>
            </div>
            {newRole && (() => {
              const info = rolesData.find(r => r.key === newRole);
              return info?.color ? (
                <div className={cn("rounded-lg p-3 bg-gradient-to-r text-white text-xs", info.color)}>
                  {info.description}
                </div>
              ) : null;
            })()}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setRoleDialog(null)}>Hủy</Button>
            <Button onClick={handleChangeRole} disabled={!newRole || newRole === roleDialog?.role}>
              Xác nhận
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
