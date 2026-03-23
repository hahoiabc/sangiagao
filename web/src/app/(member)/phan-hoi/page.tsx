"use client";

import { useEffect, useState } from "react";
import { MessageSquare, Send, Clock, Reply } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { createFeedback, getMyFeedbacks, type Feedback } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { formatDate, timeAgo } from "@/lib/utils";
import { toast } from "sonner";

export default function FeedbackPage() {
  const { token } = useAuth();
  const [feedbacks, setFeedbacks] = useState<Feedback[]>([]);
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (token) {
      getMyFeedbacks(token, 1, 50)
        .then((res) => setFeedbacks(res.data))
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [token]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    if (content.trim().length < 10) {
      toast.error("Nội dung phản hồi phải có ít nhất 10 ký tự");
      return;
    }
    setSubmitting(true);
    try {
      const fb = await createFeedback(token, content.trim());
      setFeedbacks((prev) => [fb, ...prev]);
      setContent("");
      toast.success("Gửi phản hồi thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Gửi phản hồi thất bại");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-40 w-full rounded-lg" />
        <Skeleton className="h-60 w-full rounded-lg" />
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Phản Hồi & Góp Ý</h1>

      {/* Submit form */}
      <Card className="mb-6">
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Gửi phản hồi mới
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-3">
            Ý kiến của bạn rất quan trọng để chúng tôi cải thiện dịch vụ. Hãy chia sẻ những góp ý, đề xuất hoặc vấn đề bạn gặp phải.
          </p>
          <form onSubmit={handleSubmit} className="space-y-3">
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Nhập nội dung phản hồi... (tối thiểu 10 ký tự, tối đa 2000 ký tự)"
              maxLength={2000}
              className="w-full min-h-24 rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">{content.length}/2000</span>
              <Button type="submit" className="gap-1.5" disabled={submitting || content.trim().length < 10}>
                <Send className="h-4 w-4" />
                {submitting ? "Đang gửi..." : "Gửi phản hồi"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* History */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Lịch sử phản hồi
          </CardTitle>
        </CardHeader>
        <CardContent>
          {feedbacks.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">Chưa có phản hồi nào</p>
          ) : (
            <div className="space-y-4">
              {feedbacks.map((fb) => (
                <div key={fb.id} className="border rounded-lg p-4 space-y-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <MessageSquare className="h-4 w-4 text-primary" />
                      <span className="text-xs text-muted-foreground">{timeAgo(fb.created_at)}</span>
                    </div>
                    {fb.reply ? (
                      <Badge className="bg-green-500 text-xs">Đã trả lời</Badge>
                    ) : (
                      <Badge variant="secondary" className="text-xs">Chờ trả lời</Badge>
                    )}
                  </div>
                  <p className="text-sm whitespace-pre-wrap">{fb.content}</p>

                  {fb.reply && (
                    <div className="bg-muted/50 rounded-lg p-3 space-y-1">
                      <div className="flex items-center gap-2">
                        <Reply className="h-3.5 w-3.5 text-primary" />
                        <span className="text-xs font-medium text-primary">Phản hồi từ quản trị viên</span>
                        {fb.replied_at && (
                          <span className="text-xs text-muted-foreground">{formatDate(fb.replied_at)}</span>
                        )}
                      </div>
                      <p className="text-sm">{fb.reply}</p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
