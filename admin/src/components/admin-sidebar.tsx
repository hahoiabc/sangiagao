"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState, useCallback } from "react";
import { LayoutDashboard, Users, ShoppingBasket, Flag, CreditCard, LogOut, Wheat, Megaphone, MessageSquareText, Menu, Activity, Package, GripVertical, TrendingUp, Bell, Mail, MessageCircle, Type } from "lucide-react";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import {
  DndContext, closestCenter, PointerSensor, useSensor, useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove, SortableContext, useSortable, verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";

const defaultNavItems = [
  { href: "/dashboard", label: "Tổng quan", icon: LayoutDashboard, adminOnly: false },
  { href: "/users", label: "Người dùng", icon: Users, adminOnly: true },
  { href: "/listings", label: "Tin đăng", icon: ShoppingBasket, adminOnly: false },
  { href: "/reports", label: "Báo cáo", icon: Flag, adminOnly: false },
  { href: "/subscriptions", label: "Gói dịch vụ", icon: CreditCard, adminOnly: false },
  { href: "/payments", label: "Đơn thanh toán", icon: CreditCard, adminOnly: false },
  { href: "/catalog", label: "Danh mục SP", icon: Package, adminOnly: false },
  { href: "/sponsors", label: "Tài trợ", icon: Megaphone, adminOnly: false },
  { href: "/feedbacks", label: "Góp ý", icon: MessageSquareText, adminOnly: false },
  { href: "/revenue", label: "Doanh thu", icon: TrendingUp, adminOnly: true },
  { href: "/inbox", label: "Hộp thư", icon: Mail, adminOnly: false },
  { href: "/notifications", label: "Thông báo", icon: Bell, adminOnly: true },
  { href: "/zalo-zns", label: "OTP Zalo ZNS", icon: MessageCircle, adminOnly: true },
  { href: "/slogan", label: "Slogan", icon: Type, adminOnly: false },
  { href: "/monitoring", label: "Giám sát", icon: Activity, adminOnly: false },
];

const NAV_ORDER_KEY = "admin_nav_order";

function getOrderedNavItems() {
  if (typeof window === "undefined") return defaultNavItems;
  try {
    const saved = localStorage.getItem(NAV_ORDER_KEY);
    if (!saved) return defaultNavItems;
    const order: string[] = JSON.parse(saved);
    const itemMap = new Map(defaultNavItems.map(item => [item.href, item]));
    const ordered = order.filter(href => itemMap.has(href)).map(href => itemMap.get(href)!);
    // Append any new items not in saved order
    for (const item of defaultNavItems) {
      if (!order.includes(item.href)) ordered.push(item);
    }
    return ordered;
  } catch {
    return defaultNavItems;
  }
}

function saveNavOrder(items: typeof defaultNavItems) {
  try {
    localStorage.setItem(NAV_ORDER_KEY, JSON.stringify(items.map(i => i.href)));
  } catch { /* ignore */ }
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

function useUnrepliedCount(token: string | null) {
  const [unrepliedCount, setUnrepliedCount] = useState(0);

  const fetchUnreplied = useCallback(async () => {
    if (!token) return;
    try {
      const res = await fetch(`${API_BASE}/admin/feedbacks/unreplied-count`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setUnrepliedCount(data.count || 0);
      }
    } catch {
      // ignore
    }
  }, [token]);

  useEffect(() => {
    fetchUnreplied();
    const interval = setInterval(fetchUnreplied, 60000);
    return () => clearInterval(interval);
  }, [fetchUnreplied]);

  return unrepliedCount;
}

function SortableNavItem({ item, pathname, unrepliedCount, onNavigate, reorderMode }: {
  item: typeof defaultNavItems[0];
  pathname: string;
  unrepliedCount: number;
  onNavigate?: () => void;
  reorderMode: boolean;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: item.href });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const Icon = item.icon;
  const active = pathname.startsWith(item.href);
  const showBadge = item.href === "/feedbacks" && unrepliedCount > 0;

  return (
    <div ref={setNodeRef} style={style} className="flex items-center gap-0">
      {reorderMode && (
        <button {...attributes} {...listeners} className="cursor-grab active:cursor-grabbing p-1 text-sidebar-foreground/30 hover:text-sidebar-foreground/60 shrink-0">
          <GripVertical className="h-3.5 w-3.5" />
        </button>
      )}
      <Link
        href={item.href}
        onClick={onNavigate}
        className={cn(
          "flex flex-1 items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-150",
          active
            ? "bg-sidebar-accent text-sidebar-primary border-l-[3px] border-sidebar-primary -ml-px"
            : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
        )}
      >
        <Icon className={cn("h-[18px] w-[18px]", active && "text-sidebar-primary")} />
        {item.label}
        {showBadge && (
          <span className="ml-auto inline-flex items-center justify-center rounded-full bg-red-500 px-1.5 py-0.5 text-[10px] font-bold text-white min-w-[18px]">
            {unrepliedCount > 99 ? "99+" : unrepliedCount}
          </span>
        )}
      </Link>
    </div>
  );
}

function SidebarNav({ pathname, unrepliedCount, userRole, onNavigate }: { pathname: string; unrepliedCount: number; userRole?: string; onNavigate?: () => void }) {
  const [navItems, setNavItems] = useState(defaultNavItems);
  const [reorderMode, setReorderMode] = useState(false);

  useEffect(() => {
    setNavItems(getOrderedNavItems());
  }, []);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = navItems.findIndex(i => i.href === active.id);
    const newIndex = navItems.findIndex(i => i.href === over.id);
    const reordered = arrayMove(navItems, oldIndex, newIndex);
    setNavItems(reordered);
    saveNavOrder(reordered);
  }

  const filteredItems = navItems.filter((item) => !item.adminOnly || userRole === "owner" || userRole === "admin");

  return (
    <nav className="flex-1 space-y-1 px-3 py-4">
      <div className="flex items-center justify-between mb-2 px-3">
        <p className="text-[11px] font-semibold uppercase tracking-wider text-sidebar-foreground/40">
          Quản lý
        </p>
        <button
          onClick={() => setReorderMode(!reorderMode)}
          className={cn(
            "text-[10px] px-1.5 py-0.5 rounded transition-colors",
            reorderMode
              ? "bg-sidebar-primary/20 text-sidebar-primary"
              : "text-sidebar-foreground/30 hover:text-sidebar-foreground/60"
          )}
          title={reorderMode ? "Xong" : "Sắp xếp"}
        >
          {reorderMode ? "Xong" : "Sắp xếp"}
        </button>
      </div>
      {reorderMode ? (
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd} modifiers={[restrictToVerticalAxis]}>
          <SortableContext items={filteredItems.map(i => i.href)} strategy={verticalListSortingStrategy}>
            {filteredItems.map((item) => (
              <SortableNavItem key={item.href} item={item} pathname={pathname} unrepliedCount={unrepliedCount} onNavigate={onNavigate} reorderMode={reorderMode} />
            ))}
          </SortableContext>
        </DndContext>
      ) : (
        filteredItems.map((item) => (
          <SortableNavItem key={item.href} item={item} pathname={pathname} unrepliedCount={unrepliedCount} onNavigate={onNavigate} reorderMode={reorderMode} />
        ))
      )}
    </nav>
  );
}

function SidebarFooter({ user, logout }: { user: ReturnType<typeof useAuth>["user"]; logout: () => void }) {
  return (
    <div className="border-t border-sidebar-border p-3">
      <div className="flex items-center justify-between rounded-lg px-3 py-2">
        <Link
          href="/profile"
          className="flex items-center gap-2 truncate max-w-[160px] hover:opacity-80 transition-opacity"
          title="Tài khoản của tôi"
        >
          <Avatar className="h-6 w-6">
            <AvatarImage src={user?.avatar_url} alt={user?.name || user?.phone} />
            <AvatarFallback className="text-[10px] bg-sidebar-primary/20 text-sidebar-primary">
              {(user?.name || user?.phone || "?").slice(0, 1).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="text-xs text-sidebar-foreground/50 truncate">{user?.name || user?.phone}</span>
        </Link>
        <button
          onClick={logout}
          className="flex items-center justify-center rounded-md p-1.5 text-sidebar-foreground/40 hover:bg-sidebar-accent hover:text-sidebar-foreground transition-colors"
          title="Đăng xuất"
        >
          <LogOut className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}

function SidebarBrand() {
  return (
    <div className="flex h-16 items-center gap-2.5 border-b border-sidebar-border px-5">
      <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-sidebar-primary/20">
        <Wheat className="h-4.5 w-4.5 text-sidebar-primary" />
      </div>
      <span className="text-lg font-bold text-white tracking-tight">SanGiaGao.Vn</span>
    </div>
  );
}

export function AdminSidebar() {
  const pathname = usePathname();
  const { user, token, logout } = useAuth();
  const unrepliedCount = useUnrepliedCount(token);

  return (
    <aside className="hidden lg:flex h-screen w-64 flex-col bg-sidebar">
      <SidebarBrand />
      <SidebarNav pathname={pathname} unrepliedCount={unrepliedCount} userRole={user?.role} />
      <SidebarFooter user={user} logout={logout} />
    </aside>
  );
}

export function MobileSidebar() {
  const pathname = usePathname();
  const { user, token, logout } = useAuth();
  const unrepliedCount = useUnrepliedCount(token);
  const [open, setOpen] = useState(false);

  // Close on route change
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  return (
    <div className="lg:hidden">
      <Button variant="ghost" size="icon" onClick={() => setOpen(true)} className="h-9 w-9">
        <Menu className="h-5 w-5" />
        <span className="sr-only">Menu</span>
      </Button>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="left" className="w-64 p-0 bg-sidebar border-sidebar-border" showCloseButton={false}>
          <SheetTitle className="sr-only">Menu điều hướng</SheetTitle>
          <SidebarBrand />
          <SidebarNav pathname={pathname} unrepliedCount={unrepliedCount} userRole={user?.role} onNavigate={() => setOpen(false)} />
          <SidebarFooter user={user} logout={logout} />
        </SheetContent>
      </Sheet>
    </div>
  );
}
