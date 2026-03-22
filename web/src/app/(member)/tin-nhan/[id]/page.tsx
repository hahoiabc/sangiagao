"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Send, ImagePlus, X, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getMessages,
  sendMessage as apiSendMessage,
  markConversationRead,
  uploadImage,
  type Message,
  type PaginatedResponse,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { timeAgo } from "@/lib/utils";

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

    // Poll for new messages every 3 seconds
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

    // Send images first if any
    if (selectedImages.length > 0) {
      await handleSendImages();
    }

    // Send text message
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

  const isBusy = sending || uploadingImages;

  return (
    <div className="flex flex-col h-[calc(100vh-10rem)]">
      <div className="flex items-center gap-2 pb-4 border-b">
        <Link href="/tin-nhan">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h2 className="font-semibold">Trò chuyện</h2>
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
              <div key={msg.id} className={cn("flex", isMine ? "justify-end" : "justify-start")}>
                <div
                  className={cn(
                    "max-w-[70%] rounded-2xl px-4 py-2 text-sm",
                    isMine
                      ? "bg-primary text-primary-foreground rounded-br-md"
                      : "bg-muted rounded-bl-md"
                  )}
                >
                  {msg.type === "recalled" ? (
                    <p className="italic text-xs opacity-70">{msg.content}</p>
                  ) : msg.type === "image" ? (
                    <img src={msg.content} alt="" className="max-w-full rounded-lg cursor-pointer" onClick={() => window.open(msg.content, "_blank")} />
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
