"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft, Send, ImagePlus, X, Loader2, Trash2, RotateCcw,
  CheckSquare, Square, Mic, StopCircle, Package,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getMessages,
  sendMessage as apiSendMessage,
  markConversationRead,
  uploadImage,
  uploadAudio,
  deleteMessage,
  recallMessage,
  batchDeleteMessages,
  batchRecallMessages,
  getListingDetail,
  type Message,
  type ListingDetail,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { timeAgo, formatPrice, formatQuantity } from "@/lib/utils";
import { toast } from "sonner";

const MAX_CHAT_IMAGES = 10;
const RECALL_LIMIT_MS = 24 * 60 * 60 * 1000; // 24 hours

// Phoenix WebSocket protocol helpers
let wsRef = 0;
function nextRef() {
  return String(++wsRef);
}

function formatDateHeader(dateStr: string) {
  const d = new Date(dateStr);
  const now = new Date();
  const isToday = d.toDateString() === now.toDateString();
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = d.toDateString() === yesterday.toDateString();
  if (isToday) return "Hôm nay";
  if (isYesterday) return "Hôm qua";
  return `${d.getDate().toString().padStart(2, "0")}/${(d.getMonth() + 1).toString().padStart(2, "0")}/${d.getFullYear()}`;
}

function groupMessagesByDay(messages: Message[]) {
  const groups: { date: string; messages: Message[] }[] = [];
  let currentDate = "";
  for (const msg of messages) {
    const d = new Date(msg.created_at).toDateString();
    if (d !== currentDate) {
      currentDate = d;
      groups.push({ date: msg.created_at, messages: [msg] });
    } else {
      groups[groups.length - 1].messages.push(msg);
    }
  }
  return groups;
}

function canRecall(msg: Message) {
  return Date.now() - new Date(msg.created_at).getTime() < RECALL_LIMIT_MS;
}

// Listing link cache
const listingCache: Record<string, ListingDetail | null> = {};

function ListingLinkBubble({ content, isMine }: { content: string; isMine: boolean }) {
  const [listing, setListing] = useState<ListingDetail | null | undefined>(undefined);
  const listingId = content.replace("listing://", "");

  useEffect(() => {
    if (listingCache[listingId] !== undefined) {
      setListing(listingCache[listingId]);
      return;
    }
    getListingDetail(listingId)
      .then((l) => {
        listingCache[listingId] = l;
        setListing(l);
      })
      .catch(() => {
        listingCache[listingId] = null;
        setListing(null);
      });
  }, [listingId]);

  if (listing === undefined) {
    return <Skeleton className="h-16 w-48" />;
  }
  if (!listing) {
    return <p className="text-xs opacity-70 italic">Tin đăng không tồn tại</p>;
  }

  return (
    <Link href={`/san-giao-dich/${listingId}`} className="block">
      <div className={cn(
        "rounded-lg border p-2 space-y-1 min-w-[200px]",
        isMine ? "border-primary-foreground/30" : "border-border"
      )}>
        <div className="flex items-center gap-2">
          <Package className="h-4 w-4 flex-shrink-0" />
          <span className="text-sm font-medium line-clamp-1">{listing.title}</span>
        </div>
        <div className="text-xs space-y-0.5">
          <p className="font-semibold">{formatPrice(listing.price_per_kg)}</p>
          <p>{formatQuantity(listing.quantity_kg)}</p>
          {listing.harvest_season && <p>Vụ: {listing.harvest_season}</p>}
        </div>
      </div>
    </Link>
  );
}

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
  const pollingRef = useRef<ReturnType<typeof setInterval>>(undefined);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Message actions
  const [selectMode, setSelectMode] = useState(false);
  const [selectedMsgIds, setSelectedMsgIds] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; msg: Message } | null>(null);

  // WebSocket
  const wsSocketRef = useRef<WebSocket | null>(null);
  const heartbeatRef = useRef<ReturnType<typeof setInterval>>(undefined);
  const [wsConnected, setWsConnected] = useState(false);

  // Typing indicator
  const [typingUser, setTypingUser] = useState<string | null>(null);
  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const lastTypingSentRef = useRef(0);

  // Audio recording
  const [recording, setRecording] = useState(false);
  const [recordingTime, setRecordingTime] = useState(0);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const recordingTimerRef = useRef<ReturnType<typeof setInterval>>(undefined);

  // --- WebSocket connection ---
  const connectWs = useCallback(() => {
    if (!token || !convId) return;

    const wsUrl = typeof window !== "undefined"
      ? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/socket/websocket?token=${token}`
      : "";

    if (!wsUrl) return;

    const ws = new WebSocket(wsUrl);
    wsSocketRef.current = ws;

    ws.onopen = () => {
      // Join channel
      ws.send(JSON.stringify({
        topic: `chat:${convId}`,
        event: "phx_join",
        payload: {},
        ref: nextRef(),
      }));
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.event === "phx_reply" && data.payload?.status === "ok" && data.topic === `chat:${convId}`) {
          setWsConnected(true);
          // Stop polling when WS connected
          if (pollingRef.current) {
            clearInterval(pollingRef.current);
            pollingRef.current = undefined;
          }
        }

        if (data.event === "new_message" && data.topic === `chat:${convId}`) {
          const msg: Message = {
            id: data.payload.id,
            conversation_id: data.payload.conversation_id,
            sender_id: data.payload.sender_id,
            content: data.payload.content,
            type: data.payload.type,
            created_at: data.payload.timestamp || data.payload.created_at,
          };
          setMessages((prev) => {
            if (prev.some((m) => m.id === msg.id)) return prev;
            return [...prev, msg];
          });
          // Mark as read if from other user
          if (msg.sender_id !== user?.id) {
            markConversationRead(token, convId).catch(() => {});
          }
          setTypingUser(null);
        }

        if (data.event === "typing" && data.topic === `chat:${convId}`) {
          if (data.payload.user_id !== user?.id) {
            setTypingUser(data.payload.user_id);
            if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
            typingTimeoutRef.current = setTimeout(() => setTypingUser(null), 3000);
          }
        }
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      setWsConnected(false);
      // Restart polling as fallback
      if (!pollingRef.current && token && convId) {
        startPolling();
      }
    };

    ws.onerror = () => {
      // will trigger onclose
    };

    // Heartbeat every 30s
    heartbeatRef.current = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          topic: "phoenix",
          event: "heartbeat",
          payload: {},
          ref: nextRef(),
        }));
      }
    }, 30000);
  }, [token, convId, user?.id]);

  function startPolling() {
    pollingRef.current = setInterval(async () => {
      try {
        const res = await getMessages(token!, convId!, 1, 100);
        setMessages((res.data ?? []).reverse());
      } catch {
        // ignore
      }
    }, 3000);
  }

  // Send typing event
  function sendTyping() {
    const now = Date.now();
    if (now - lastTypingSentRef.current < 2000) return;
    lastTypingSentRef.current = now;

    const ws = wsSocketRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        topic: `chat:${convId}`,
        event: "typing",
        payload: {},
        ref: nextRef(),
      }));
    }
  }

  // --- Init ---
  useEffect(() => {
    if (!token || !convId) return;

    async function fetchMessages() {
      try {
        const res = await getMessages(token!, convId!, 1, 100);
        setMessages((res.data ?? []).reverse());
        markConversationRead(token!, convId!).catch(() => {});
      } catch {
        // ignore
      } finally {
        setLoading(false);
      }
    }

    fetchMessages();
    connectWs();

    // Start polling as initial fallback (will be stopped if WS connects)
    startPolling();

    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current);
      if (heartbeatRef.current) clearInterval(heartbeatRef.current);
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
      const ws = wsSocketRef.current;
      if (ws) {
        ws.close();
        wsSocketRef.current = null;
      }
    };
  }, [token, convId, connectWs]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    return () => {
      selectedImages.forEach((img) => URL.revokeObjectURL(img.preview));
    };
  }, [selectedImages]);

  // Close context menu
  useEffect(() => {
    if (!contextMenu) return;
    function handleClick() { setContextMenu(null); }
    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, [contextMenu]);

  // --- Audio recording ---
  async function startRecording() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
          ? "audio/webm;codecs=opus"
          : "audio/webm",
      });
      audioChunksRef.current = [];

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) audioChunksRef.current.push(e.data);
      };

      mediaRecorder.onstop = async () => {
        stream.getTracks().forEach((t) => t.stop());
        const blob = new Blob(audioChunksRef.current, { type: "audio/webm" });
        if (blob.size > 0 && token && convId) {
          try {
            const { url } = await uploadAudio(token, blob);
            const msg = await apiSendMessage(token, convId, url, "audio");
            setMessages((prev) => [...prev, msg]);
            toast.success("Đã gửi tin nhắn thoại");
          } catch (err) {
            toast.error(err instanceof Error ? err.message : "Gửi audio thất bại");
          }
        }
      };

      mediaRecorder.start();
      mediaRecorderRef.current = mediaRecorder;
      setRecording(true);
      setRecordingTime(0);
      recordingTimerRef.current = setInterval(() => {
        setRecordingTime((t) => t + 1);
      }, 1000);
    } catch {
      toast.error("Không thể truy cập microphone");
    }
  }

  function stopRecording() {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
      mediaRecorderRef.current.stop();
    }
    mediaRecorderRef.current = null;
    setRecording(false);
    if (recordingTimerRef.current) {
      clearInterval(recordingTimerRef.current);
      recordingTimerRef.current = undefined;
    }
  }

  function cancelRecording() {
    if (mediaRecorderRef.current) {
      mediaRecorderRef.current.ondataavailable = null;
      mediaRecorderRef.current.onstop = null;
      if (mediaRecorderRef.current.state !== "inactive") {
        mediaRecorderRef.current.stop();
      }
      mediaRecorderRef.current.stream.getTracks().forEach((t) => t.stop());
    }
    mediaRecorderRef.current = null;
    audioChunksRef.current = [];
    setRecording(false);
    if (recordingTimerRef.current) {
      clearInterval(recordingTimerRef.current);
      recordingTimerRef.current = undefined;
    }
  }

  function formatRecordingTime(secs: number) {
    const m = Math.floor(secs / 60).toString().padStart(2, "0");
    const s = (secs % 60).toString().padStart(2, "0");
    return `${m}:${s}`;
  }

  // --- Image handling ---
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

  // --- Message actions ---
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
    if (!canRecall(msg)) {
      toast.error("Chỉ có thể thu hồi tin nhắn trong vòng 24 giờ");
      setContextMenu(null);
      return;
    }
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
      return msg && msg.sender_id === user?.id && canRecall(msg);
    });
    if (myMsgIds.length === 0) {
      toast.error("Chỉ có thể thu hồi tin nhắn của bạn trong vòng 24 giờ");
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
  const messageGroups = groupMessagesByDay(messages);

  return (
    <div className="flex flex-col h-[calc(100vh-10rem)]">
      {/* Header */}
      <div className="flex items-center gap-2 pb-4 border-b">
        <Link href="/tin-nhan">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div className="flex-1">
          <h2 className="font-semibold">Trò chuyện</h2>
          {wsConnected && (
            <span className="text-[10px] text-green-600">Trực tuyến</span>
          )}
        </div>
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

      {/* Messages */}
      <div className="flex-1 overflow-y-auto py-4 space-y-1">
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => <Skeleton key={i} className="h-10 w-48" />)}
          </div>
        ) : messages.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">Chưa có tin nhắn</p>
        ) : (
          messageGroups.map((group, gi) => (
            <div key={gi}>
              {/* Date header */}
              <div className="flex justify-center my-3">
                <span className="text-[11px] text-muted-foreground bg-muted px-3 py-1 rounded-full">
                  {formatDateHeader(group.date)}
                </span>
              </div>
              <div className="space-y-3">
                {group.messages.map((msg) => {
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
                        ) : msg.type === "audio" ? (
                          <audio controls src={msg.content} className="max-w-[250px]" />
                        ) : msg.type === "listing_link" ? (
                          <ListingLinkBubble content={msg.content} isMine={isMine} />
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
                })}
              </div>
            </div>
          ))
        )}

        {/* Typing indicator */}
        {typingUser && (
          <div className="flex justify-start">
            <div className="bg-muted rounded-2xl rounded-bl-md px-4 py-2 text-sm">
              <span className="text-xs text-muted-foreground italic">đang soạn tin...</span>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Context menu */}
      {contextMenu && (
        <div
          className="fixed z-50 bg-white border rounded-lg shadow-lg py-1 min-w-[140px]"
          style={{ top: contextMenu.y, left: contextMenu.x }}
        >
          {contextMenu.msg.sender_id === user?.id && canRecall(contextMenu.msg) && (
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

      {/* Recording bar */}
      {recording && (
        <div className="flex items-center gap-3 px-4 py-3 border-t bg-red-50">
          <div className="h-3 w-3 rounded-full bg-red-500 animate-pulse" />
          <span className="text-sm font-mono text-red-600">{formatRecordingTime(recordingTime)}</span>
          <span className="text-sm text-muted-foreground flex-1">Đang ghi âm...</span>
          <Button variant="ghost" size="sm" onClick={cancelRecording}>Hủy</Button>
          <Button variant="destructive" size="sm" className="gap-1" onClick={stopRecording}>
            <StopCircle className="h-4 w-4" />
            Gửi
          </Button>
        </div>
      )}

      {/* Image preview bar */}
      {selectedImages.length > 0 && !recording && (
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

      {/* Input bar */}
      {!recording && (
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
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={startRecording}
            disabled={isBusy}
            title="Ghi âm"
          >
            <Mic className="h-5 w-5" />
          </Button>
          <Input
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              sendTyping();
            }}
            placeholder="Nhập tin nhắn..."
            className="flex-1"
            disabled={isBusy}
          />
          <Button type="submit" size="icon" disabled={isBusy || (!input.trim() && selectedImages.length === 0)}>
            {uploadingImages ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          </Button>
        </form>
      )}
    </div>
  );
}
