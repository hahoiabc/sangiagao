"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { MessageCircle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getConversations, type Conversation, type PaginatedResponse } from "@/services/api";
import { timeAgo } from "@/lib/utils";
import { useAuth } from "@/lib/auth";

export default function ConversationsPage() {
  const { token } = useAuth();
  const [result, setResult] = useState<PaginatedResponse<Conversation> | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      getConversations(token, 1, 50)
        .then(setResult)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [token]);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Tin Nhắn</h1>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-20 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <div className="space-y-2">
          {result.data.map((conv) => (
            <Link key={conv.id} href={`/tin-nhan/${conv.id}`}>
              <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
                <CardContent className="p-4 flex items-center gap-3">
                  <Avatar className="h-10 w-10">
                    <AvatarFallback className="bg-primary/10 text-primary text-sm">
                      {(conv.other_user?.name || "?").charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <p className="font-medium text-sm truncate">
                        {conv.other_user?.name || "Người dùng"}
                      </p>
                      <span className="text-xs text-muted-foreground flex-shrink-0">
                        {timeAgo(conv.last_message_at)}
                      </span>
                    </div>
                    {conv.other_user?.org_name && (
                      <p className="text-xs text-muted-foreground truncate">{conv.other_user.org_name}</p>
                    )}
                  </div>
                  {conv.unread_count > 0 && (
                    <Badge className="h-5 min-w-5 flex items-center justify-center text-xs rounded-full">
                      {conv.unread_count}
                    </Badge>
                  )}
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <MessageCircle className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
          <p className="text-muted-foreground">Chưa có cuộc trò chuyện nào</p>
          <p className="text-sm text-muted-foreground mt-1">
            Liên hệ người bán từ trang tin đăng để bắt đầu trò chuyện
          </p>
        </div>
      )}
    </div>
  );
}
