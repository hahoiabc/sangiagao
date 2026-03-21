"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { listFeedbacks, replyFeedback, type Feedback } from "@/services/api";

export default function FeedbacksPage() {
  const { token } = useAuth();
  const [feedbacks, setFeedbacks] = useState<Feedback[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [replyDialog, setReplyDialog] = useState<Feedback | null>(null);
  const [replyText, setReplyText] = useState("");
  const [sending, setSending] = useState(false);

  const limit = 20;

  const fetchFeedbacks = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await listFeedbacks(token, page, limit);
      setFeedbacks(res.data);
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token, page]);

  useEffect(() => {
    fetchFeedbacks();
  }, [fetchFeedbacks]);

  function openReply(fb: Feedback) {
    setReplyDialog(fb);
    setReplyText(fb.reply || "");
  }

  async function handleReply() {
    if (!token || !replyDialog || !replyText.trim()) return;
    setSending(true);
    try {
      await replyFeedback(token, replyDialog.id, replyText.trim());
      toast.success("Đã gửi phản hồi");
      setReplyDialog(null);
      setReplyText("");
      fetchFeedbacks();
    } catch {
      toast.error("Gửi phản hồi thất bại");
    } finally {
      setSending(false);
    }
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      <h1 className="text-xl font-semibold mb-5">Góp ý từ thành viên</h1>

      <div className="rounded-lg border shadow-sm bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]">#</TableHead>
              <TableHead>Thành viên</TableHead>
              <TableHead>Nội dung góp ý</TableHead>
              <TableHead>Phản hồi</TableHead>
              <TableHead>Ngày gửi</TableHead>
              <TableHead className="text-right">Thao tác</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
              </TableRow>
            ) : feedbacks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Chưa có góp ý nào</TableCell>
              </TableRow>
            ) : (
              feedbacks.map((fb, i) => (
                <TableRow key={fb.id}>
                  <TableCell className="text-muted-foreground">{(page - 1) * limit + i + 1}</TableCell>
                  <TableCell>
                    <div className="font-medium text-sm">{fb.user_name}</div>
                    <div className="text-xs text-muted-foreground">{fb.user_phone}</div>
                  </TableCell>
                  <TableCell className="max-w-[300px]">
                    <p className="text-sm whitespace-pre-wrap line-clamp-3">{fb.content}</p>
                  </TableCell>
                  <TableCell className="max-w-[250px]">
                    {fb.reply ? (
                      <div>
                        <p className="text-sm whitespace-pre-wrap line-clamp-2 text-green-700">{fb.reply}</p>
                        {fb.replied_at && (
                          <p className="text-xs text-muted-foreground mt-1">
                            {new Date(fb.replied_at).toLocaleDateString("vi-VN")}
                          </p>
                        )}
                      </div>
                    ) : (
                      <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200">
                        Chờ phản hồi
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(fb.created_at).toLocaleDateString("vi-VN")}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button size="sm" variant={fb.reply ? "outline" : "default"} onClick={() => openReply(fb)}>
                      {fb.reply ? "Sửa phản hồi" : "Phản hồi"}
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
        <span className="text-sm text-muted-foreground">
          Trang {page} / {totalPages || 1} ({total} góp ý)
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

      <Dialog open={!!replyDialog} onOpenChange={() => setReplyDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Phản hồi góp ý</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <p className="text-sm font-medium mb-1">Từ: {replyDialog?.user_name} ({replyDialog?.user_phone})</p>
              <div className="bg-muted/50 rounded-md p-3 text-sm whitespace-pre-wrap">
                {replyDialog?.content}
              </div>
            </div>
            <div>
              <p className="text-sm font-medium mb-1">Phản hồi của bạn:</p>
              <textarea
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                rows={4}
                placeholder="Nhập nội dung phản hồi..."
                value={replyText}
                onChange={(e) => setReplyText(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setReplyDialog(null)}>Hủy</Button>
            <Button onClick={handleReply} disabled={!replyText.trim() || sending}>
              {sending ? "Đang gửi..." : "Gửi phản hồi"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
