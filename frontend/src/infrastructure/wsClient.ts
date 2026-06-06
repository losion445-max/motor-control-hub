// infrastructure/ws-client.ts
import type { WsCommand, WsEvent } from '../domain/wstypes';

type EventHandler = (event: WsEvent) => void;

class WebSocketClient {
    private ws: WebSocket | null = null;
    private url: string;
    private listeners: Set<EventHandler> = new Set();
    private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    public isConnected = false;

    constructor(url: string) {
        this.url = url;
    }

    public connect() {
        if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
            return;
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
            this.isConnected = true;
            console.log('[WS] Connected to Motor Control Hub');
            this.send({ type: 'GET_CONFIG' });
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data) as WsEvent;
                this.listeners.forEach(listener => listener(data));
            } catch (err) {
                console.error('[WS] Message parse error:', err);
            }
        };

        this.ws.onclose = () => {
            this.isConnected = false;
            console.log('[WS] Disconnected. Reconnecting in 2s...');
            this.ws = null;
            this.scheduleReconnect();
        };

        this.ws.onerror = (err) => {
            console.error('[WS] Connection error:', err);
            this.ws?.close();
        };
    }

    private scheduleReconnect() {
        if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
        this.reconnectTimer = setTimeout(() => this.connect(), 2000);
    }

    public send(command: WsCommand) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(command));
        } else {
            console.warn('[WS] Cannot send command, socket is closed:', command.type);
        }
    }

    public subscribe(listener: EventHandler) {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener); // Возвращает функцию отписки
    }
}

export const wsApi = new WebSocketClient('ws://localhost:8080/ws');