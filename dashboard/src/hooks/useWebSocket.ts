import { useState, useEffect, useRef, useCallback } from 'react';

export interface WebSocketHook {
  sendMessage: (msg: string) => void;
  lastMessage: MessageEvent<string> | null;
  readyState: number;
}

export function useWebSocket(url: string): WebSocketHook {
  const [lastMessage, setLastMessage] = useState<MessageEvent<string> | null>(null);
  const [readyState, setReadyState] = useState<number>(WebSocket.CONNECTING);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setReadyState(WebSocket.OPEN);
      console.log('[WS] Connected to', url);
    };

    ws.onmessage = (event) => {
      setLastMessage(event);
    };

    ws.onclose = () => {
      setReadyState(WebSocket.CLOSED);
      console.log('[WS] Disconnected');
    };

    ws.onerror = (error) => {
      console.error('[WS] Error:', error);
      setReadyState(WebSocket.CLOSED);
    };

    return () => {
      ws.close();
    };
  }, [url]);

  const sendMessage = useCallback((msg: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(msg);
    }
  }, []);

  return { sendMessage, lastMessage, readyState };
}
