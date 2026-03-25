"use client";

import { useAuth } from "@/lib/auth";
import { useRouter, usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";
import { AlertTriangle, X } from "lucide-react";
import { Navbar } from "@/components/navbar";
import { Footer } from "@/components/footer";
import { Button } from "@/components/ui/button";
import { getSubscriptionStatus, type SubscriptionStatus } from "@/services/api";

// Routes that are always accessible (like mobile's allowed list)
const ALWAYS_ALLOWED = ["/tai-khoan", "/goi-thanh-vien", "/thong-bao", "/phan-hoi", "/tin-nhan"];
// Privileged roles that skip subscription gate
const PRIVILEGED_ROLES = ["editor", "admin", "owner"];

export default function MemberLayout({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [subStatus, setSubStatus] = useState<SubscriptionStatus | null>(null);
  const [subLoaded, setSubLoaded] = useState(false);
  const [bannerDismissed, setBannerDismissed] = useState(false);

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace("/dang-nhap");
    }
  }, [user, isLoading, router]);

  // Check subscription status
  useEffect(() => {
    if (!user) return;
    if (PRIVILEGED_ROLES.includes(user.role)) {
      setSubLoaded(true);
      return;
    }
    getSubscriptionStatus("")
      .then(setSubStatus)
      .catch(() => {})
      .finally(() => setSubLoaded(true));
  }, [user]);

  if (isLoading || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  const isPrivileged = PRIVILEGED_ROLES.includes(user.role);
  const isAllowedRoute = ALWAYS_ALLOWED.some((r) => pathname.startsWith(r));
  const isBlocked = !isPrivileged && subLoaded && subStatus && !subStatus.has_active && !isAllowedRoute;

  // Subscription expiry banner
  const showBanner = !isPrivileged && subLoaded && subStatus && !bannerDismissed;
  const isExpired = showBanner && !subStatus.has_active;
  const isExpiringSoon = showBanner && subStatus.has_active && subStatus.days_remaining <= 15;

  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />

      {/* Subscription expiry banner - like mobile */}
      {(isExpired || isExpiringSoon) && (
        <div className={`px-4 py-2.5 text-sm flex items-center justify-between ${isExpired ? "bg-red-50 text-red-700 border-b border-red-200" : "bg-orange-50 text-orange-700 border-b border-orange-200"}`}>
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 flex-shrink-0" />
            <span>
              {isExpired
                ? "Gói dịch vụ đã hết hạn. Tin đăng đã bị tạm ẩn."
                : `Gói dịch vụ còn ${subStatus!.days_remaining} ngày. `}
            </span>
            <Link href="/goi-thanh-vien" className="font-medium underline underline-offset-2">
              Xem chi tiết
            </Link>
          </div>
          <button onClick={() => setBannerDismissed(true)} className="p-1 hover:opacity-70">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      <main className="flex-1 mx-auto w-full max-w-7xl px-4 py-6">
        {isBlocked ? (
          <div className="text-center py-16">
            <AlertTriangle className="h-12 w-12 text-orange-400 mx-auto mb-4" />
            <h2 className="text-xl font-bold mb-2">Gói dịch vụ đã hết hạn</h2>
            <p className="text-muted-foreground mb-2">
              Tin đăng của bạn đã bị tạm ẩn khỏi sàn.
            </p>
            <p className="text-muted-foreground mb-6">
              Gia hạn gói để tin đăng được hiển thị trở lại.
            </p>
            <Link href="/goi-thanh-vien">
              <Button>Xem gói dịch vụ</Button>
            </Link>
          </div>
        ) : (
          children
        )}
      </main>
      <Footer />
    </div>
  );
}
