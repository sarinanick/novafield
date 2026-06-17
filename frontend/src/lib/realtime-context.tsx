"use client";

import { createContext, useContext, useState, useEffect, useCallback, useRef, ReactNode } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";

interface RealtimeMessage {
  type: string;
  payload: any;
}

type MessageHandler = (msg: RealtimeMessage) => void;

interface RealtimeContextType {
  connected: boolean;
  sendMessage: (type: string, payload: any) => void;
  onMessage: (type: string, handler: MessageHandler) => () => void;
}

const RealtimeContext = createContext<RealtimeContextType>({
  connected: false,
  sendMessage: () => {},
  onMessage: () => () => {},
});

export function RealtimeProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const handlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map());
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const onMessage = useCallback((type: string, handler: MessageHandler) => {
    if (!handlersRef.current.has(type)) {
      handlersRef.current.set(type, new Set());
    }
    handlersRef.current.get(type)!.add(handler);
    return () => {
      handlersRef.current.get(type)?.delete(handler);
    };
  }, []);

  const sendMessage = useCallback((type: string, payload: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    }
  }, []);

  useEffect(() => {
    if (!user) {
      wsRef.current?.close();
      wsRef.current = null;
      setConnected(false);
      return;
    }

    const connect = () => {
      const ws = api.connectRealtime(user.id, (msg: RealtimeMessage) => {
        if (msg.type === "connected") {
          setConnected(true);
        }
        const handlers = handlersRef.current.get(msg.type);
        if (handlers) {
          handlers.forEach((h) => h(msg));
        }
      });

      if (!ws) return;

      ws.onopen = () => setConnected(true);
      ws.onclose = () => {
        setConnected(false);
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      };
      ws.onerror = () => ws.close();
      wsRef.current = ws;
    };

    connect();

    return () => {
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      wsRef.current?.close();
      wsRef.current = null;
      setConnected(false);
    };
  }, [user]);

  return (
    <RealtimeContext.Provider value={{ connected, sendMessage, onMessage }}>
      {children}
    </RealtimeContext.Provider>
  );
}

export const useRealtime = () => useContext(RealtimeContext);
