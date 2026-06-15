"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { motion } from "framer-motion";
import { Send, ArrowLeft, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { useRealtime } from "@/lib/realtime-context";

export default function MessagesContent() {
  const { user, loading: authLoading } = useAuth();
  const { sendMessage: sendRealtime, onMessage } = useRealtime();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [conversations, setConversations] = useState<any[]>([]);
  const [messages, setMessages] = useState<any[]>([]);
  const [selectedUser, setSelectedUser] = useState<string | null>(searchParams.get("user"));
  const [newMessage, setNewMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [typingUsers, setTypingUsers] = useState<Map<string, { name: string; timeout: NodeJS.Timeout }>>(new Map());
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (!authLoading && !user) { router.push("/auth/login"); return; }
    loadConversations();
  }, [user, authLoading]);

  useEffect(() => {
    if (selectedUser) loadMessages(selectedUser);
  }, [selectedUser]);

  useEffect(() => {
    const unsub1 = onMessage("online-status", (msg) => {
      const { userId, online } = msg.payload;
      setOnlineUsers((prev) => {
        const next = new Set(prev);
        if (online) next.add(userId); else next.delete(userId);
        return next;
      });
    });

    const unsub2 = onMessage("typing", (msg) => {
      const { userId, name, typing } = msg.payload;
      if (userId === user?.id) return;
      setTypingUsers((prev) => {
        const next = new Map(prev);
        if (typing) {
          if (next.has(userId)) {
            clearTimeout(next.get(userId)!.timeout);
          }
          const timeout = setTimeout(() => {
            setTypingUsers((p) => { const n = new Map(p); n.delete(userId); return n; });
          }, 3000);
          next.set(userId, { name, timeout });
        } else {
          if (next.has(userId)) clearTimeout(next.get(userId)!.timeout);
          next.delete(userId);
        }
        return next;
      });
    });

    const unsub3 = onMessage("connected", (msg) => {
      if (msg.payload.onlineUsers) {
        setOnlineUsers(new Set(msg.payload.onlineUsers));
      }
    });

    return () => { unsub1(); unsub2(); unsub3(); };
  }, [onMessage, user?.id]);

  const loadConversations = async () => {
    try {
      const convs = await api.getConversations();
      setConversations(convs);
    } catch {}
    setLoading(false);
  };

  const loadMessages = async (userId: string) => {
    try {
      const msgs = await api.getMessages(userId);
      setMessages(msgs);
      api.markConversationRead(userId).catch(() => {});
      setTimeout(() => messagesEndRef.current?.scrollIntoView({ behavior: "smooth" }), 100);
    } catch {}
  };

  const handleTyping = useCallback(() => {
    if (!selectedUser) return;
    sendRealtime("typing", { receiverId: selectedUser, typing: true });
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    typingTimeoutRef.current = setTimeout(() => {
      sendRealtime("typing", { receiverId: selectedUser, typing: false });
    }, 3000);
  }, [selectedUser, sendRealtime]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMessage.trim() || !selectedUser) return;
    setSending(true);
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    sendRealtime("typing", { receiverId: selectedUser, typing: false });
    try {
      await api.sendMessage({ receiverId: selectedUser, content: newMessage.trim() });
      setNewMessage("");
      await loadMessages(selectedUser);
      loadConversations();
    } catch {}
    setSending(false);
  };

  const isUserOnline = (userId: string) => onlineUsers.has(userId);
  const getTypingName = (userId: string) => typingUsers.get(userId)?.name;

  if (authLoading || loading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;

  return (
    <div className="min-h-screen pt-20 pb-4">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-4 h-[calc(100vh-6rem)]">
        <div className="glass-card rounded-2xl h-full flex overflow-hidden">
          <div className={`w-full md:w-80 border-r border-white/5 flex flex-col ${selectedUser ? "hidden md:flex" : "flex"}`}>
            <div className="p-4 border-b border-white/5">
              <h2 className="text-lg font-semibold mb-3">Messages</h2>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input placeholder="Search conversations..." className="pl-9 h-9 text-sm" />
              </div>
            </div>
            <div className="flex-1 overflow-y-auto">
              {conversations.length === 0 ? (
                <p className="text-center text-muted-foreground text-sm py-8">No conversations yet</p>
              ) : conversations.map((conv: any) => (
                <motion.button key={conv.id} whileHover={{ x: 2 }} onClick={() => setSelectedUser(conv.otherUser?.id)} className={`w-full text-left p-4 border-b border-white/5 hover:bg-white/5 transition-colors ${selectedUser === conv.otherUser?.id ? "bg-primary/10" : ""}`}>
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-sm font-bold shrink-0">{conv.otherUser?.name?.[0] || "?"}</div>
                      <div className={`absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-background ${isUserOnline(conv.otherUser?.id) ? "bg-green-500" : "bg-gray-500"}`} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex justify-between items-center">
                        <p className="font-medium text-sm truncate">{conv.otherUser?.name}</p>
                        <span className="text-[10px] text-muted-foreground shrink-0">{conv.lastMessageAt?.split("T")[1]?.slice(0, 5)}</span>
                      </div>
                      <p className="text-xs text-muted-foreground truncate mt-0.5">{conv.lastMessage || "No messages yet"}</p>
                    </div>
                    {conv.unreadCount > 0 && <span className="w-5 h-5 rounded-full bg-primary text-[10px] flex items-center justify-center font-bold shrink-0">{conv.unreadCount}</span>}
                  </div>
                </motion.button>
              ))}
            </div>
          </div>

          <div className={`flex-1 flex flex-col ${selectedUser ? "flex" : "hidden md:flex"}`}>
            {selectedUser ? (
              <>
                <div className="p-4 border-b border-white/5 flex items-center gap-3">
                  <button onClick={() => setSelectedUser(null)} className="md:hidden p-1 hover:bg-white/10 rounded-lg"><ArrowLeft className="w-5 h-5" /></button>
                  <div className="relative">
                    <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-xs font-bold">
                      {conversations.find(c => c.otherUser?.id === selectedUser)?.otherUser?.name?.[0] || "?"}
                    </div>
                    <div className={`absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-background ${isUserOnline(selectedUser) ? "bg-green-500" : "bg-gray-500"}`} />
                  </div>
                  <div>
                    <p className="font-medium text-sm">{conversations.find(c => c.otherUser?.id === selectedUser)?.otherUser?.name}</p>
                    <p className="text-[10px] text-muted-foreground">{isUserOnline(selectedUser) ? "Online" : "Offline"}</p>
                  </div>
                </div>
                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                  {messages.map((msg: any) => (
                    <motion.div key={msg.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className={`flex ${msg.senderId === user?.id ? "justify-end" : "justify-start"}`}>
                      <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${msg.senderId === user?.id ? "bg-primary text-primary-foreground rounded-br-md" : "glass-card rounded-bl-md"}`}>
                        <p>{msg.content}</p>
                        <p className={`text-[10px] mt-1 ${msg.senderId === user?.id ? "text-primary-foreground/60" : "text-muted-foreground"}`}>{msg.createdAt?.split("T")[1]?.slice(0, 5)}</p>
                      </div>
                    </motion.div>
                  ))}
                  {typingUsers.has(selectedUser) && (
                    <div className="flex justify-start">
                      <div className="glass-card px-4 py-2.5 rounded-2xl rounded-bl-md text-sm">
                        <div className="flex items-center gap-2">
                          <div className="flex gap-1">
                            <span className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce" style={{ animationDelay: "0ms" }} />
                            <span className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce" style={{ animationDelay: "150ms" }} />
                            <span className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce" style={{ animationDelay: "300ms" }} />
                          </div>
                          <span className="text-xs text-muted-foreground">{getTypingName(selectedUser)} is typing...</span>
                        </div>
                      </div>
                    </div>
                  )}
                  <div ref={messagesEndRef} />
                </div>
                <form onSubmit={handleSend} className="p-4 border-t border-white/5 flex gap-3">
                  <Input placeholder="Type a message..." value={newMessage} onChange={e => { setNewMessage(e.target.value); handleTyping(); }} className="flex-1" />
                  <Button type="submit" variant="glow" size="icon" disabled={sending || !newMessage.trim()}><Send className="w-4 h-4" /></Button>
                </form>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <div className="w-16 h-16 rounded-2xl glass-card flex items-center justify-center mx-auto mb-4"><Send className="w-8 h-8 text-muted-foreground" /></div>
                  <p className="text-lg font-medium">Select a conversation</p>
                  <p className="text-sm text-muted-foreground mt-1">Choose from your existing conversations or start a new one</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
