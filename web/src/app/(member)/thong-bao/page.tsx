"use client";

import { useEffect, useState } from "react";
import { Bell, MessageCircle, Star, CreditCard, Flag } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { getNotifications, markNotificationRead, type AppNotification, type PaginatedResponse } from "@/services/api";
import { timeAgo } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

// Match mobile notification type icons
function getNotificationIcon(type: string) {
  switch (type) {
    case "message":
      return <MessageCircle className="h-5 w-5" />;
    case "rating":
      return <Star className="h-5 w-5" />;
    case "subscription":
      return <CreditCard className="h-5 w-5" />;
    case "report":
      return <Flag className="h-5 w-5" />;
    default:
      return <Bell className="h-5 w-5" />;
  }
}

export default function NotificationsPage() {
  const { token } = useAuth();
  const [result, setResult] = useState<PaginatedResponse<AppNotification> | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      getNotifications(token, 1, 50)
        .then(setResult)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [token]);

  async function handleRead(id: string) {
    if (!token) return;
    try {
      await markNotificationRead(token, id);
      setResult((prev) =>
        prev
          ? {
              ...prev,
              data: prev.data.map((n) => (n.id === id ? { ...n, is_read: true } : n)),
            }
          : prev
      );
    } catch {
      // ignore
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Thông báo</h1>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-16 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <div className="space-y-2">
          {result.data.map((notif) => (
            <Card
              key={notif.id}
              className={cn("cursor-pointer transition-colors", !notif.is_read && "bg-primary/5 border-primary/20")}
              onClick={() => !notif.is_read && handleRead(notif.id)}
            >
              <CardContent className="p-4">
                <div className="flex items-start gap-3">
                  <div className={cn(
                    "flex-shrink-0 mt-0.5",
                    notif.is_read ? "text-muted-foreground" : "text-primary"
                  )}>
                    {getNotificationIcon(notif.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-start justify-between gap-2">
                      <p className={cn("text-sm", !notif.is_read && "font-semibold")}>{notif.title}</p>
                      <span className="text-xs text-muted-foreground flex-shrink-0">
                        {timeAgo(notif.created_at)}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-0.5 line-clamp-2">{notif.body}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <Bell className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
          <p className="text-muted-foreground">Chưa có thông báo</p>
          <p className="text-sm text-muted-foreground mt-1">
            Thông báo mới sẽ hiển thị ở đây
          </p>
        </div>
      )}
    </div>
  );
}
