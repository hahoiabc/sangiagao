"use client";

import { useEffect, useState } from "react";
import { Mail, Pin, ChevronRight, ArrowLeft, Image as ImageIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getInbox, getInboxDetail, markInboxRead, type InboxMessage } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { timeAgo, formatDate } from "@/lib/utils";
import NextImage from "next/image";

export default function SystemInboxPage() {
  const { user } = useAuth();
  const [messages, setMessages] = useState<InboxMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [unreadCount, setUnreadCount] = useState(0);

  // Detail view
  const [selected, setSelected] = useState<InboxMessage | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  useEffect(() => {
    if (user) {
      getInbox("", 1, 50)
        .then((res) => {
          setMessages(res.data ?? []);
          setUnreadCount(res.unread_count ?? 0);
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [user]);

  async function handleSelect(msg: InboxMessage) {
    // Mark read in list
    if (!msg.is_read) {
      setMessages((prev) =>
        prev.map((m) => (m.id === msg.id ? { ...m, is_read: true } : m))
      );
      setUnreadCount((prev) => Math.max(0, prev - 1));
      markInboxRead("", msg.id).catch(() => {});
    }

    // Load detail
    setDetailLoading(true);
    try {
      const detail = await getInboxDetail("", msg.id);
      setSelected(detail);
    } catch {
      setSelected(msg);
    } finally {
      setDetailLoading(false);
    }
  }

  function handleBack() {
    setSelected(null);
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-8 w-40" />
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-lg" />
        ))}
      </div>
    );
  }

  // Detail view
  if (selected) {
    return (
      <div className="max-w-2xl mx-auto">
        <button
          onClick={handleBack}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeft className="h-4 w-4" />
          Quay lại
        </button>

        <Card>
          <CardContent className="p-6">
            {detailLoading ? (
              <div className="space-y-4">
                <Skeleton className="h-6 w-3/4" />
                <Skeleton className="h-4 w-1/4" />
                <Skeleton className="h-40 w-full" />
              </div>
            ) : (
              <div className="space-y-4">
                {selected.is_pinned && (
                  <Badge variant="secondary" className="gap-1">
                    <Pin className="h-3 w-3" />
                    Ghim
                  </Badge>
                )}

                <h1 className="text-xl font-bold">{selected.title}</h1>

                <p className="text-sm text-muted-foreground">
                  {formatDate(selected.created_at)}
                </p>

                <hr />

                {selected.image_url && (
                  <div className="relative w-full aspect-video rounded-lg overflow-hidden bg-muted">
                    <NextImage
                      src={selected.image_url}
                      alt={selected.title}
                      fill
                      className="object-cover"
                    />
                  </div>
                )}

                <p className="text-sm leading-relaxed whitespace-pre-wrap">
                  {selected.body}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    );
  }

  // List view
  return (
    <div className="max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Hộp thư</h1>
        {unreadCount > 0 && (
          <Badge variant="secondary">{unreadCount} chưa đọc</Badge>
        )}
      </div>

      {messages.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Mail className="h-12 w-12 mx-auto text-muted-foreground/30 mb-3" />
            <p className="text-muted-foreground">Chưa có thông báo</p>
            <p className="text-sm text-muted-foreground/60 mt-1">
              Thông báo từ hệ thống sẽ hiển thị ở đây
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {messages.map((msg) => (
            <Card
              key={msg.id}
              className={`cursor-pointer transition-colors hover:bg-muted/50 ${
                !msg.is_read ? "bg-primary/[0.04] border-primary/20" : ""
              }`}
              onClick={() => handleSelect(msg)}
            >
              <CardContent className="p-4 flex items-start gap-3">
                <div className="mt-0.5">
                  {msg.is_pinned ? (
                    <Pin className="h-5 w-5 text-primary" />
                  ) : (
                    <Mail
                      className={`h-5 w-5 ${
                        msg.is_read ? "text-muted-foreground/40" : "text-primary"
                      }`}
                    />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <h3
                      className={`text-sm truncate ${
                        msg.is_read ? "font-normal" : "font-bold"
                      }`}
                    >
                      {msg.title}
                    </h3>
                    <span className="text-xs text-muted-foreground whitespace-nowrap">
                      {timeAgo(msg.created_at)}
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground line-clamp-2 mt-1">
                    {msg.body}
                  </p>
                </div>
                <ChevronRight className="h-4 w-4 text-muted-foreground/40 mt-1 shrink-0" />
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
