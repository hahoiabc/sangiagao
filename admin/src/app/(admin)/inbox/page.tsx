"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Pencil, Trash2, Pin, ImageIcon, X } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  listInbox,
  createInbox,
  updateInbox,
  deleteInbox,
  type InboxMessage,
  type CreateInboxRequest,
  type UpdateInboxRequest,
} from "@/services/api";

type FormMode = "create" | "edit" | null;

const TARGET_OPTIONS = [
  { value: "all_users", label: "Tất cả người dùng" },
  { value: "role:member", label: "Thành viên (member)" },
  { value: "role:seller", label: "Người bán (seller)" },
];

export default function InboxPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<InboxMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [formMode, setFormMode] = useState<FormMode>(null);
  const [editId, setEditId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  // Form fields
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [target, setTarget] = useState("all_users");
  const [isPinned, setIsPinned] = useState(false);
  const [expiresAt, setExpiresAt] = useState("");

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  async function load() {
    if (!user) return;
    setLoading(true);
    try {
      const res = await listInbox(1, 50);
      setItems(res.data || []);
    } catch {
      toast.error("Không thể tải danh sách hộp thư");
    } finally {
      setLoading(false);
    }
  }

  function resetForm() {
    setTitle("");
    setBody("");
    setImageUrl("");
    setTarget("all_users");
    setIsPinned(false);
    setExpiresAt("");
    setFormMode(null);
    setEditId(null);
  }

  function startEdit(item: InboxMessage) {
    setTitle(item.title);
    setBody(item.body);
    setImageUrl(item.image_url || "");
    setTarget(item.target);
    setIsPinned(item.is_pinned);
    setExpiresAt(item.expires_at ? item.expires_at.slice(0, 16) : "");
    setEditId(item.id);
    setFormMode("edit");
  }

  async function handleSubmit() {
    if (!user || !title.trim() || !body.trim()) return;

    try {
      if (formMode === "create") {
        const data: CreateInboxRequest = {
          title: title.trim(),
          body: body.trim(),
          target,
          is_pinned: isPinned,
        };
        if (imageUrl.trim()) data.image_url = imageUrl.trim();
        if (expiresAt) data.expires_at = new Date(expiresAt).toISOString();
        await createInbox(data);
        toast.success("Đã tạo thông báo + gửi push");
      } else if (formMode === "edit" && editId) {
        const data: UpdateInboxRequest = {
          title: title.trim(),
          body: body.trim(),
          is_pinned: isPinned,
        };
        if (imageUrl.trim()) data.image_url = imageUrl.trim();
        if (expiresAt) data.expires_at = new Date(expiresAt).toISOString();
        await updateInbox(editId, data);
        toast.success("Đã cập nhật thông báo");
      }
      resetForm();
      load();
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Lỗi";
      toast.error(`Thao tác thất bại: ${msg}`);
    }
  }

  async function handleDelete(id: string) {
    if (!user) return;
    try {
      await deleteInbox(id);
      toast.success("Đã xóa thông báo");
      setConfirmDeleteId(null);
      load();
    } catch {
      toast.error("Xóa thất bại");
    }
  }

  function formatDate(iso: string) {
    const d = new Date(iso);
    return d.toLocaleDateString("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" });
  }

  const canSubmit = title.trim().length > 0 && body.trim().length > 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Hộp thư hệ thống</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Tạo thông báo hiển thị trong inbox của user (1 bản ghi, nhiều người đọc)
          </p>
        </div>
        {!formMode && (
          <Button onClick={() => setFormMode("create")}>
            <Plus className="h-4 w-4 mr-2" />
            Tạo thông báo
          </Button>
        )}
      </div>

      {/* Form */}
      {formMode && (
        <div className="rounded-lg border p-6 space-y-4 max-w-2xl">
          <h2 className="text-lg font-semibold">
            {formMode === "create" ? "Tạo thông báo mới" : "Sửa thông báo"}
          </h2>

          <div>
            <label className="text-sm font-medium mb-1.5 block">Tiêu đề *</label>
            <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Nhập tiêu đề..." maxLength={200} />
            <p className="text-xs text-muted-foreground mt-1">{title.length}/200</p>
          </div>

          <div>
            <label className="text-sm font-medium mb-1.5 block">Nội dung *</label>
            <textarea
              className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              value={body}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setBody(e.target.value)}
              placeholder="Nhập nội dung..."
              rows={5}
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-1.5 block">Đối tượng</label>
            <select
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              {TARGET_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-6">
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={isPinned} onChange={(e) => setIsPinned(e.target.checked)} className="rounded" />
              <Pin className="h-4 w-4" />
              Ghim lên đầu
            </label>
          </div>

          <div>
            <label className="text-sm font-medium mb-1.5 block">Hết hạn (tùy chọn)</label>
            <Input type="datetime-local" value={expiresAt} onChange={(e) => setExpiresAt(e.target.value)} />
          </div>

          <div>
            <label className="text-sm font-medium mb-1.5 block">
              <ImageIcon className="h-4 w-4 inline mr-1.5" />
              Hình ảnh (tùy chọn)
            </label>
            <div className="flex gap-2">
              <Input value={imageUrl} onChange={(e) => setImageUrl(e.target.value)} placeholder="URL hình ảnh (https://...)..." type="url" />
              {imageUrl && (
                <Button variant="ghost" size="icon" onClick={() => setImageUrl("")} className="shrink-0">
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
            {imageUrl.trim() && (
              <div className="mt-2 rounded-lg border overflow-hidden max-w-[300px]">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={imageUrl.trim()}
                  alt="Preview"
                  className="w-full h-auto max-h-[200px] object-cover"
                  onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
                  onLoad={(e) => { (e.target as HTMLImageElement).style.display = "block"; }}
                />
              </div>
            )}
          </div>

          <div className="flex gap-2 pt-2">
            <Button onClick={handleSubmit} disabled={!canSubmit}>
              {formMode === "create" ? "Tạo + Gửi push" : "Lưu thay đổi"}
            </Button>
            <Button variant="outline" onClick={resetForm}>Hủy</Button>
          </div>
        </div>
      )}

      {/* List */}
      {loading ? (
        <p className="text-muted-foreground">Đang tải...</p>
      ) : items.length === 0 ? (
        <p className="text-muted-foreground">Chưa có thông báo nào</p>
      ) : (
        <div className="rounded-md border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3 font-medium">Tiêu đề</th>
                <th className="text-left px-4 py-3 font-medium w-[120px]">Đối tượng</th>
                <th className="text-center px-4 py-3 font-medium w-[60px]">Ghim</th>
                <th className="text-left px-4 py-3 font-medium w-[140px]">Ngày tạo</th>
                <th className="text-left px-4 py-3 font-medium w-[140px]">Hết hạn</th>
                <th className="text-right px-4 py-3 font-medium w-[100px]">Thao tác</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b hover:bg-muted/30">
                  <td className="px-4 py-3">
                    <p className="font-medium">{item.title}</p>
                    <p className="text-muted-foreground text-xs mt-0.5 line-clamp-1">{item.body}</p>
                  </td>
                  <td className="px-4 py-3 text-xs">
                    {TARGET_OPTIONS.find((o) => o.value === item.target)?.label || item.target}
                  </td>
                  <td className="px-4 py-3 text-center">
                    {item.is_pinned && <Pin className="h-4 w-4 text-orange-500 mx-auto" />}
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">{formatDate(item.created_at)}</td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">
                    {item.expires_at ? formatDate(item.expires_at) : "—"}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => startEdit(item)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      {confirmDeleteId === item.id ? (
                        <div className="flex gap-1">
                          <Button variant="destructive" size="sm" className="h-8 text-xs" onClick={() => handleDelete(item.id)}>
                            Xóa
                          </Button>
                          <Button variant="outline" size="sm" className="h-8 text-xs" onClick={() => setConfirmDeleteId(null)}>
                            Hủy
                          </Button>
                        </div>
                      ) : (
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => setConfirmDeleteId(item.id)}>
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
