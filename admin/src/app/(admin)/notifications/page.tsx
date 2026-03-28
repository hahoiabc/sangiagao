"use client";

import { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Send, ImageIcon, X, Search, User as UserIcon } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { broadcastNotification, sendNotification, listUsers, type User } from "@/services/api";

type Tab = "broadcast" | "individual";

export default function NotificationsPage() {
  const { token } = useAuth();
  const [tab, setTab] = useState<Tab>("broadcast");

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Gửi thông báo</h1>
        <p className="text-muted-foreground text-sm mt-1">
          Gửi thông báo đẩy tới thành viên
        </p>
      </div>

      <div className="flex gap-1 border-b">
        <button
          onClick={() => setTab("broadcast")}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            tab === "broadcast"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          Gửi tất cả
        </button>
        <button
          onClick={() => setTab("individual")}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            tab === "individual"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          Gửi cá nhân
        </button>
      </div>

      {tab === "broadcast" ? (
        <BroadcastForm token={token} />
      ) : (
        <IndividualForm token={token} />
      )}
    </div>
  );
}

function BroadcastForm({ token }: { token: string | null }) {
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
    <div className="max-w-xl space-y-4">
      <NotificationFields
        title={title}
        setTitle={setTitle}
        body={body}
        setBody={setBody}
        imageUrl={imageUrl}
        setImageUrl={setImageUrl}
      />

      {!confirmOpen ? (
        <Button onClick={() => setConfirmOpen(true)} disabled={!canSend}>
          <Send className="h-4 w-4 mr-2" />
          Gửi tới tất cả
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
  );
}

function IndividualForm({ token }: { token: string | null }) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [sending, setSending] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  async function handleSend() {
    if (!selectedUser || !title.trim() || !body.trim()) return;
    setSending(true);
    try {
      await sendNotification(token ?? "", {
        user_id: selectedUser.id,
        title: title.trim(),
        body: body.trim(),
        ...(imageUrl.trim() ? { image_url: imageUrl.trim() } : {}),
      });
      toast.success(`Đã gửi thông báo tới ${selectedUser.name || selectedUser.phone}`);
      setTitle("");
      setBody("");
      setImageUrl("");
      setSelectedUser(null);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Lỗi không xác định";
      toast.error(`Gửi thất bại: ${msg}`);
      console.error("Send individual error:", err);
    } finally {
      setSending(false);
    }
  }

  const canSend = !!selectedUser && title.trim().length > 0 && body.trim().length > 0;

  return (
    <div className="max-w-xl space-y-4">
      <div>
        <label className="text-sm font-medium mb-1.5 block">
          <UserIcon className="h-4 w-4 inline mr-1.5" />
          Người nhận
        </label>
        {selectedUser ? (
          <div className="flex items-center gap-2 rounded-md border px-3 py-2">
            <span className="text-sm flex-1">
              {selectedUser.name || "Chưa đặt tên"} — {selectedUser.phone}
            </span>
            <Button variant="ghost" size="icon" onClick={() => setSelectedUser(null)} className="h-6 w-6">
              <X className="h-3 w-3" />
            </Button>
          </div>
        ) : (
          <UserSearch token={token} onSelect={setSelectedUser} />
        )}
      </div>

      <NotificationFields
        title={title}
        setTitle={setTitle}
        body={body}
        setBody={setBody}
        imageUrl={imageUrl}
        setImageUrl={setImageUrl}
      />

      <Button onClick={handleSend} disabled={!canSend || sending}>
        <Send className="h-4 w-4 mr-2" />
        {sending ? "Đang gửi..." : "Gửi thông báo"}
      </Button>
    </div>
  );
}

function UserSearch({ token, onSelect }: { token: string | null; onSelect: (u: User) => void }) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setOpen(false);
      return;
    }
    setLoading(true);
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      try {
        const res = await listUsers(token ?? "", query.trim(), 1, 10);
        setResults(res.data || []);
        setOpen(true);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 300);
    return () => clearTimeout(debounceRef.current);
  }, [query, token]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={containerRef}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => results.length > 0 && setOpen(true)}
          placeholder="Tìm theo tên hoặc SĐT..."
          className="pl-9"
        />
      </div>
      {open && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover shadow-md max-h-[240px] overflow-y-auto">
          {loading ? (
            <p className="p-3 text-sm text-muted-foreground">Đang tìm...</p>
          ) : results.length === 0 ? (
            <p className="p-3 text-sm text-muted-foreground">Không tìm thấy</p>
          ) : (
            results.map((u) => (
              <button
                key={u.id}
                onClick={() => { onSelect(u); setQuery(""); setOpen(false); }}
                className="w-full text-left px-3 py-2 text-sm hover:bg-accent transition-colors flex items-center gap-2"
              >
                <span className="font-medium">{u.name || "Chưa đặt tên"}</span>
                <span className="text-muted-foreground">{u.phone}</span>
                <span className="ml-auto text-xs text-muted-foreground">{u.role}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}

function NotificationFields({
  title, setTitle, body, setBody, imageUrl, setImageUrl,
}: {
  title: string; setTitle: (v: string) => void;
  body: string; setBody: (v: string) => void;
  imageUrl: string; setImageUrl: (v: string) => void;
}) {
  return (
    <>
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
    </>
  );
}
