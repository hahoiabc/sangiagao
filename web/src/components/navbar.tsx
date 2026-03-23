"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Wheat, Search, MessageCircle, Bell, User, LogOut, Menu, X, Crown, MessageSquareText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { useState, useEffect, useRef } from "react";
import { getConversations } from "@/services/api";

const publicLinks = [
  { href: "/bang-gia", label: "Sàn gạo", perm: "marketplace.priceboard" },
  { href: "/san-giao-dich", label: "Sàn giao dịch", perm: "marketplace.browse" },
];

const memberLinks = [
  { href: "/tin-dang", label: "Tin của tôi", icon: Search, perm: "listings.create" },
  { href: "/tin-nhan", label: "Tin nhắn", icon: MessageCircle, perm: "chat.send" },
  { href: "/thong-bao", label: "Thông báo", icon: Bell, perm: null },
  { href: "/goi-thanh-vien", label: "Gói dịch vụ", icon: Crown, perm: null },
  { href: "/phan-hoi", label: "Góp ý", icon: MessageSquareText, perm: "feedback.create" },
];

export function Navbar() {
  const { user, token, logout, hasPermission } = useAuth();
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined);

  // Poll unread count like mobile (every 10 seconds)
  useEffect(() => {
    if (!token) {
      setUnreadCount(0);
      return;
    }

    async function fetchUnread() {
      try {
        const res = await getConversations(token!, 1, 50);
        const total = (res.data ?? []).reduce((sum, c) => sum + c.unread_count, 0);
        setUnreadCount(total);
      } catch {
        // ignore
      }
    }

    fetchUnread();
    intervalRef.current = setInterval(fetchUnread, 10000);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [token]);

  function renderBadge(href: string) {
    if (href === "/tin-nhan" && unreadCount > 0) {
      return (
        <Badge className="h-5 min-w-5 flex items-center justify-center text-[10px] rounded-full px-1.5 ml-1">
          {unreadCount > 99 ? "99+" : unreadCount}
        </Badge>
      );
    }
    return null;
  }

  return (
    <header className="sticky top-0 z-50 border-b bg-white/95 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
        <div className="flex items-center gap-6">
          <Link href="/bang-gia" className="flex items-center gap-2">
            <Wheat className="h-6 w-6 text-primary" />
            <span className="text-lg font-bold text-primary hidden sm:inline">SanGiaGao</span>
          </Link>
          <nav className="hidden md:flex items-center gap-1">
            {publicLinks.filter((l) => !l.perm || hasPermission(l.perm)).map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className={cn(
                  "px-3 py-2 rounded-md text-sm font-medium transition-colors",
                  pathname.startsWith(l.href)
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                {l.label}
              </Link>
            ))}
            {user &&
              memberLinks.filter((l) => !l.perm || hasPermission(l.perm)).map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  className={cn(
                    "px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-1.5",
                    pathname.startsWith(l.href)
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  <l.icon className="h-4 w-4" />
                  {l.label}
                  {renderBadge(l.href)}
                </Link>
              ))}
          </nav>
        </div>

        <div className="hidden md:flex items-center gap-2">
          {user ? (
            <>
              <Link href="/tai-khoan">
                <Button variant="ghost" size="sm" className="gap-1.5">
                  <User className="h-4 w-4" />
                  {user.name || user.phone}
                </Button>
              </Link>
              <Button variant="ghost" size="sm" onClick={logout}>
                <LogOut className="h-4 w-4" />
              </Button>
            </>
          ) : (
            <>
              <Link href="/dang-nhap">
                <Button variant="ghost" size="sm">Đăng nhập</Button>
              </Link>
              <Link href="/dang-ky">
                <Button size="sm">Đăng ký</Button>
              </Link>
            </>
          )}
        </div>

        <Button variant="ghost" size="icon" className="md:hidden" onClick={() => setMobileOpen(!mobileOpen)}>
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </Button>
      </div>

      {mobileOpen && (
        <div className="md:hidden border-t bg-white p-4 space-y-2">
          {publicLinks.filter((l) => !l.perm || hasPermission(l.perm)).map((l) => (
            <Link
              key={l.href}
              href={l.href}
              onClick={() => setMobileOpen(false)}
              className={cn(
                "block px-3 py-2 rounded-md text-sm font-medium",
                pathname.startsWith(l.href) ? "bg-primary/10 text-primary" : "text-muted-foreground"
              )}
            >
              {l.label}
            </Link>
          ))}
          {user &&
            memberLinks.filter((l) => !l.perm || hasPermission(l.perm)).map((l) => (
              <Link
                key={l.href}
                href={l.href}
                onClick={() => setMobileOpen(false)}
                className={cn(
                  "flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium",
                  pathname.startsWith(l.href) ? "bg-primary/10 text-primary" : "text-muted-foreground"
                )}
              >
                <l.icon className="h-4 w-4" />
                {l.label}
                {renderBadge(l.href)}
              </Link>
            ))}
          <div className="border-t pt-2 mt-2">
            {user ? (
              <>
                <Link href="/tai-khoan" onClick={() => setMobileOpen(false)} className="block px-3 py-2 text-sm">
                  Tài khoản
                </Link>
                <button onClick={logout} className="block px-3 py-2 text-sm text-destructive">
                  Đăng xuất
                </button>
              </>
            ) : (
              <>
                <Link href="/dang-nhap" onClick={() => setMobileOpen(false)} className="block px-3 py-2 text-sm">
                  Đăng nhập
                </Link>
                <Link href="/dang-ky" onClick={() => setMobileOpen(false)} className="block px-3 py-2 text-sm font-medium text-primary">
                  Đăng ký
                </Link>
              </>
            )}
          </div>
        </div>
      )}
    </header>
  );
}
