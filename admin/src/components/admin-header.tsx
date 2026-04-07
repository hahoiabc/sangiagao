"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { MobileSidebar } from "@/components/admin-sidebar";
import { useAuth } from "@/lib/auth";

const pageTitles: Record<string, string> = {
  "/dashboard": "Tổng quan",
  "/users": "Quản lý người dùng",
  "/listings": "Quản lý tin đăng",
  "/reports": "Hàng đợi báo cáo",
  "/subscriptions": "Quản lý gói dịch vụ",
  "/sponsors": "Quản lý tài trợ",
  "/feedbacks": "Góp ý",
  "/profile": "Tài khoản của tôi",
  "/slogan": "Quản lý Slogan",
  "/gioi-thieu": "Trang giới thiệu",
  "/monitoring": "Giám sát hệ thống",
};

export function AdminHeader() {
  const pathname = usePathname();
  const { user } = useAuth();

  // Match the longest prefix
  const title = Object.entries(pageTitles).find(([path]) => pathname.startsWith(path))?.[1]
    || "Quản trị";

  // Build breadcrumb for detail pages
  const isDetail = /\/[a-f0-9-]{36}$/.test(pathname);
  const parentPath = isDetail ? pathname.replace(/\/[^/]+$/, "") : null;
  const parentTitle = parentPath ? pageTitles[parentPath] : null;

  return (
    <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-card/80 backdrop-blur-sm px-4 lg:px-6">
      <div className="flex items-center gap-3 text-sm">
        <MobileSidebar />
        {parentTitle ? (
          <>
            <span className="text-muted-foreground">{parentTitle}</span>
            <span className="text-muted-foreground/40">/</span>
            <span className="font-semibold">Chi tiết</span>
          </>
        ) : (
          <h2 className="text-base font-semibold">{title}</h2>
        )}
      </div>
      <Link href="/profile" className="flex items-center gap-3 rounded-lg px-2 py-1 -mr-2 hover:bg-muted/50 transition-colors">
        <Avatar className="h-8 w-8">
          <AvatarImage src={user?.avatar_url} alt={user?.name || user?.phone} />
          <AvatarFallback className="text-xs font-semibold bg-primary/10 text-primary">
            {(user?.name || user?.phone || "?").slice(0, 1).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm text-muted-foreground hidden sm:block">{user?.name || user?.phone}</span>
      </Link>
    </header>
  );
}
