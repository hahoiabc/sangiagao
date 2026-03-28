"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Send, ImageIcon, X } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { broadcastNotification } from "@/services/api";

export default function NotificationsPage() {
  const { token } = useAuth();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [sending, setSending] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);

  async function handleSend() {
    if (!title.trim() || !body.trim()) return;
    setSending(true);
    try {
      const payload: { title: string; body: string; image_url?: string } = {
        title: title.trim(),
        body: body.trim(),
      };
      if (imageUrl.trim()) {
        payload.image_url = imageUrl.trim();
      }
      const result = await broadcastNotification(token ?? "", payload);
      toast.success(`Gửi thành công tới ${result.sent_to} thành viên`);
      setTitle("");
      setBody("");
      setImageUrl("");
      setConfirmOpen(false);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Lỗi không xác định";
      toast.error(`Gửi thông báo thất bại: ${msg}`);
      console.error("Broadcast error:", err);
    } finally {
      setSending(false);
    }
  }

  const canSend = title.trim().length > 0 && body.trim().length > 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Gửi thông báo</h1>
        <p className="text-muted-foreground text-sm mt-1">
          Gửi thông báo đẩy tới tất cả thành viên
        </p>
      </div>

      <div className="max-w-xl space-y-4">
        <div>
          <label className="text-sm font-medium mb-1.5 block">Tiêu đề</label>
          <Input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Nhập tiêu đề thông báo..."
            maxLength={200}
          />
          <p className="text-xs text-muted-foreground mt-1">{title.length}/200</p>
        </div>

        <div>
          <label className="text-sm font-medium mb-1.5 block">Nội dung</label>
          <textarea
            className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            value={body}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setBody(e.target.value)}
            placeholder="Nhập nội dung thông báo..."
            rows={5}
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-1.5 block">
            <ImageIcon className="h-4 w-4 inline mr-1.5" />
            Hình ảnh (tùy chọn)
          </label>
          <div className="flex gap-2">
            <Input
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="Dán URL hình ảnh (https://...)..."
              type="url"
            />
            {imageUrl && (
              <Button variant="ghost" size="icon" onClick={() => setImageUrl("")} className="shrink-0">
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            URL hình ảnh công khai (HTTPS). Hiển thị dưới dạng ảnh lớn trên thông báo đẩy.
          </p>
          {imageUrl.trim() && (
            <div className="mt-2 rounded-lg border overflow-hidden max-w-[300px]">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={imageUrl.trim()}
                alt="Preview"
                className="w-full h-auto max-h-[200px] object-cover"
                onError={(e) => {
                  (e.target as HTMLImageElement).style.display = "none";
                }}
                onLoad={(e) => {
                  (e.target as HTMLImageElement).style.display = "block";
                }}
              />
            </div>
          )}
        </div>

        {!confirmOpen ? (
          <Button onClick={() => setConfirmOpen(true)} disabled={!canSend}>
            <Send className="h-4 w-4 mr-2" />
            Gửi thông báo
          </Button>
        ) : (
          <div className="rounded-lg border border-orange-200 bg-orange-50 p-4 space-y-3">
            <p className="text-sm font-medium text-orange-800">
              Xác nhận gửi thông báo tới tất cả thành viên?
            </p>
            <div className="text-sm text-orange-700 space-y-1">
              <p><strong>Tiêu đề:</strong> {title}</p>
              <p><strong>Nội dung:</strong> {body}</p>
              {imageUrl.trim() && <p><strong>Hình ảnh:</strong> Có đính kèm</p>}
            </div>
            <div className="flex gap-2">
              <Button onClick={handleSend} disabled={sending} size="sm">
                {sending ? "Đang gửi..." : "Xác nhận gửi"}
              </Button>
              <Button variant="outline" size="sm" onClick={() => setConfirmOpen(false)} disabled={sending}>
                Hủy
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
