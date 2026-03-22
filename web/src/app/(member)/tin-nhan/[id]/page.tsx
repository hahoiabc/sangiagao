"use client";

import { useEffect, useState, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getMessages,
  sendMessage as apiSendMessage,
  markConversationRead,
  type Message,
  type PaginatedResponse,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { timeAgo } from "@/lib/utils";

export default function ChatRoomPage() {
  const { id: convId } = useParams<{ id: string }>();
  const { user, token } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined);

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

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !convId || !input.trim()) return;
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
                    <img src={msg.content} alt="" className="max-w-full rounded-lg" />
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

      <form onSubmit={handleSend} className="flex gap-2 pt-4 border-t">
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Nhập tin nhắn..."
          className="flex-1"
          disabled={sending}
        />
        <Button type="submit" size="icon" disabled={sending || !input.trim()}>
          <Send className="h-4 w-4" />
        </Button>
      </form>
    </div>
  );
}
