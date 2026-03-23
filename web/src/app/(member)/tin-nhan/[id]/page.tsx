"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Send, ImagePlus, X, Loader2, Trash2, RotateCcw, CheckSquare, Square } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getMessages,
  sendMessage as apiSendMessage,
  markConversationRead,
  uploadImage,
  deleteMessage,
  recallMessage,
  batchDeleteMessages,
  batchRecallMessages,
  type Message,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { timeAgo } from "@/lib/utils";
import { toast } from "sonner";

const MAX_CHAT_IMAGES = 10;

export default function ChatRoomPage() {
  const { id: convId } = useParams<{ id: string }>();
  const { user, token } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const [uploadingImages, setUploadingImages] = useState(false);
  const [selectedImages, setSelectedImages] = useState<{ file: File; preview: string }[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Message actions state
  const [selectMode, setSelectMode] = useState(false);
  const [selectedMsgIds, setSelectedMsgIds] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; msg: Message } | null>(null);

  useEffect(() => {
    if (!token || !convId) return;

    async function fetchMessages() {
      try {
        const res = await getMessages(token!, convId!, 1, 100);
        setMessages(res.data.reverse());
        markConversationRead(token!, convId!).catch(() => {});
      } catch {
        // ignore
      } finally {
        setLoading(false);
      }
    }

    fetchMessages();

    intervalRef.current = setInterval(async () => {
      try {
        const res = await getMessages(token!, convId!, 1, 100);
        setMessages(res.data.reverse());
      } catch {
        // ignore
      }
    }, 3000);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [token, convId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    return () => {
      selectedImages.forEach((img) => URL.revokeObjectURL(img.preview));
    };
  }, [selectedImages]);

  // Close context menu on click outside
  useEffect(() => {
    function handleClick() {
      setContextMenu(null);
    }
    if (contextMenu) {
      window.addEventListener("click", handleClick);
      return () => window.removeEventListener("click", handleClick);
    }
  }, [contextMenu]);

  function handleImageSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files;
    if (!files) return;
    const remaining = MAX_CHAT_IMAGES - selectedImages.length;
    const selected = Array.from(files).slice(0, remaining);
    const newImages = selected.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));
    setSelectedImages((prev) => [...prev, ...newImages]);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  function removeImage(index: number) {
    setSelectedImages((prev) => {
      URL.revokeObjectURL(prev[index].preview);
      return prev.filter((_, i) => i !== index);
    });
  }

  async function handleSendImages() {
    if (!token || !convId || selectedImages.length === 0) return;
    setUploadingImages(true);
    try {
      for (const img of selectedImages) {
        const { url } = await uploadImage(token, img.file, "listings");
        const msg = await apiSendMessage(token, convId, url, "image");
        setMessages((prev) => [...prev, msg]);
      }
      selectedImages.forEach((img) => URL.revokeObjectURL(img.preview));
      setSelectedImages([]);
    } catch {
      // ignore
    } finally {
      setUploadingImages(false);
    }
  }

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !convId) return;

    if (selectedImages.length > 0) {
      await handleSendImages();
    }

    if (input.trim()) {
      setSending(true);
      try {
        const msg = await apiSendMessage(token, convId, input.trim());
        setMessages((prev) => [...prev, msg]);
        setInput("");
      } catch {
        // ignore
      } finally {
        setSending(false);
      }
    }
  }

  function handleContextMenu(e: React.MouseEvent, msg: Message) {
    if (msg.type === "recalled") return;
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, msg });
  }

  async function handleDeleteMsg(msg: Message) {
    if (!token || !convId) return;
    try {
      await deleteMessage(token, convId, msg.id);
      setMessages((prev) => prev.filter((m) => m.id !== msg.id));
      toast.success("Đã xóa tin nhắn");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa thất bại");
    }
    setContextMenu(null);
  }

  async function handleRecallMsg(msg: Message) {
    if (!token || !convId) return;
    try {
      await recallMessage(token, convId, msg.id);
      setMessages((prev) =>
        prev.map((m) => (m.id === msg.id ? { ...m, type: "recalled", content: "Tin nhắn đã được thu hồi" } : m))
      );
      toast.success("Đã thu hồi tin nhắn");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Thu hồi thất bại");
    }
    setContextMenu(null);
  }

  function toggleSelectMsg(msgId: string) {
    setSelectedMsgIds((prev) => {
      const next = new Set(prev);
      if (next.has(msgId)) next.delete(msgId);
      else next.add(msgId);
      return next;
    });
  }

  async function handleBatchDelete() {
    if (!token || !convId || selectedMsgIds.size === 0) return;
    try {
      await batchDeleteMessages(token, convId, Array.from(selectedMsgIds));
      setMessages((prev) => prev.filter((m) => !selectedMsgIds.has(m.id)));
      toast.success(`Đã xóa ${selectedMsgIds.size} tin nhắn`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa thất bại");
    }
    setSelectedMsgIds(new Set());
    setSelectMode(false);
  }

  async function handleBatchRecall() {
    if (!token || !convId || selectedMsgIds.size === 0) return;
    const myMsgIds = Array.from(selectedMsgIds).filter((id) => {
      const msg = messages.find((m) => m.id === id);
      return msg && msg.sender_id === user?.id;
    });
    if (myMsgIds.length === 0) {
      toast.error("Chỉ có thể thu hồi tin nhắn của bạn");
      return;
    }
    try {
      await batchRecallMessages(token, convId, myMsgIds);
      setMessages((prev) =>
        prev.map((m) =>
          myMsgIds.includes(m.id) ? { ...m, type: "recalled", content: "Tin nhắn đã được thu hồi" } : m
        )
      );
      toast.success(`Đã thu hồi ${myMsgIds.length} tin nhắn`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Thu hồi thất bại");
    }
    setSelectedMsgIds(new Set());
    setSelectMode(false);
  }

  const isBusy = sending || uploadingImages;

  return (
    <div className="flex flex-col h-[calc(100vh-10rem)]">
      <div className="flex items-center gap-2 pb-4 border-b">
        <Link href="/tin-nhan">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h2 className="font-semibold flex-1">Trò chuyện</h2>
        {selectMode ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">{selectedMsgIds.size} đã chọn</span>
            <Button variant="outline" size="sm" className="gap-1" onClick={handleBatchRecall} disabled={selectedMsgIds.size === 0}>
              <RotateCcw className="h-3.5 w-3.5" />
              Thu hồi
            </Button>
            <Button variant="destructive" size="sm" className="gap-1" onClick={handleBatchDelete} disabled={selectedMsgIds.size === 0}>
              <Trash2 className="h-3.5 w-3.5" />
              Xóa
            </Button>
            <Button variant="ghost" size="sm" onClick={() => { setSelectMode(false); setSelectedMsgIds(new Set()); }}>
              Hủy
            </Button>
          </div>
        ) : (
          <Button variant="ghost" size="sm" className="gap-1" onClick={() => setSelectMode(true)}>
            <CheckSquare className="h-4 w-4" />
            Chọn
          </Button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto py-4 space-y-3">
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => <Skeleton key={i} className="h-10 w-48" />)}
          </div>
        ) : messages.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">Chưa có tin nhắn</p>
        ) : (
          messages.map((msg) => {
            const isMine = msg.sender_id === user?.id;
            return (
              <div
                key={msg.id}
                className={cn("flex items-start gap-2", isMine ? "justify-end" : "justify-start")}
                onContextMenu={(e) => handleContextMenu(e, msg)}
              >
                {selectMode && (
                  <button
                    type="button"
                    onClick={() => toggleSelectMsg(msg.id)}
                    className="mt-2 text-muted-foreground hover:text-foreground"
                  >
                    {selectedMsgIds.has(msg.id) ? (
                      <CheckSquare className="h-4 w-4 text-primary" />
                    ) : (
                      <Square className="h-4 w-4" />
                    )}
                  </button>
                )}
                <div
                  className={cn(
                    "max-w-[70%] rounded-2xl px-4 py-2 text-sm group relative",
                    isMine
                      ? "bg-primary text-primary-foreground rounded-br-md"
                      : "bg-muted rounded-bl-md"
                  )}
                >
                  {msg.type === "recalled" ? (
                    <p className="italic text-xs opacity-70">{msg.content}</p>
                  ) : msg.type === "image" ? (
                    <img src={msg.content} alt="" className="max-w-full rounded-lg cursor-pointer" onClick={() => window.open(msg.content, "_blank")} />
                  ) : msg.type === "audio" ? (
                    <audio controls src={msg.content} className="max-w-full" />
                  ) : (
                    <p className="whitespace-pre-wrap">{msg.content}</p>
                  )}
                  <p
                    className={cn(
                      "text-[10px] mt-1",
                      isMine ? "text-primary-foreground/70" : "text-muted-foreground"
                    )}
                  >
                    {timeAgo(msg.created_at)}
                  </p>
                </div>
              </div>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Context menu */}
      {contextMenu && (
        <div
          className="fixed z-50 bg-white border rounded-lg shadow-lg py-1 min-w-[140px]"
          style={{ top: contextMenu.y, left: contextMenu.x }}
        >
          {contextMenu.msg.sender_id === user?.id && (
            <button
              className="w-full text-left px-4 py-2 text-sm hover:bg-muted flex items-center gap-2"
              onClick={() => handleRecallMsg(contextMenu.msg)}
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Thu hồi
            </button>
          )}
          <button
            className="w-full text-left px-4 py-2 text-sm hover:bg-muted flex items-center gap-2 text-destructive"
            onClick={() => handleDeleteMsg(contextMenu.msg)}
          >
            <Trash2 className="h-3.5 w-3.5" />
            Xóa
          </button>
        </div>
      )}

      {/* Image preview bar */}
      {selectedImages.length > 0 && (
        <div className="flex gap-2 pt-3 pb-1 overflow-x-auto">
          {selectedImages.map((img, i) => (
            <div key={i} className="relative w-16 h-16 rounded-lg overflow-hidden border flex-shrink-0">
              <img src={img.preview} alt="" className="w-full h-full object-cover" />
              <button
                type="button"
                onClick={() => removeImage(i)}
                className="absolute top-0.5 right-0.5 bg-black/60 text-white rounded-full p-0.5 hover:bg-black/80"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
          {selectedImages.length < MAX_CHAT_IMAGES && (
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="w-16 h-16 rounded-lg border-2 border-dashed border-muted-foreground/30 flex items-center justify-center text-muted-foreground hover:border-primary hover:text-primary flex-shrink-0"
            >
              <ImagePlus className="h-5 w-5" />
            </button>
          )}
        </div>
      )}

      <form onSubmit={handleSend} className="flex gap-2 pt-4 border-t">
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          onChange={handleImageSelect}
          className="hidden"
        />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => fileInputRef.current?.click()}
          disabled={isBusy || selectedImages.length >= MAX_CHAT_IMAGES}
          title={`Gửi ảnh (tối đa ${MAX_CHAT_IMAGES})`}
        >
          <ImagePlus className="h-5 w-5" />
        </Button>
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Nhập tin nhắn..."
          className="flex-1"
          disabled={isBusy}
        />
        <Button type="submit" size="icon" disabled={isBusy || (!input.trim() && selectedImages.length === 0)}>
          {uploadingImages ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
        </Button>
      </form>
    </div>
  );
}
