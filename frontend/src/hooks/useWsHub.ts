// presentation/useHub.ts
import { useState, useEffect, useCallback } from 'react';
import { wsApi } from '../infrastructure/wsClient';
import type { SystemStatus, FullConfig } from '../domain/types';
import type { WsCommand } from '../domain/wstypes';

export const useWsHub = () => {
    const [status, setStatus] = useState<SystemStatus | null>(null);
    const [config, setConfig] = useState<FullConfig | null>(null);
    const [isLive, setIsLive] = useState(false);

    useEffect(() => {
        wsApi.connect();

        const unsubscribe = wsApi.subscribe((event) => {
            switch (event.type) {
                case 'STATUS_UPDATE':
                    setStatus(event.payload);
                    setIsLive(true);
                    break;
                case 'CONFIG_DATA':
                    setConfig(event.payload);
                    break;
                case 'ERROR':
                    console.error("Hub Error:", event.payload.message);
                    break;
            }
        });

        const liveCheckInterval = setInterval(() => {
            setIsLive(wsApi.isConnected);
        }, 1000);

        return () => {
            unsubscribe();
            clearInterval(liveCheckInterval);
        };
    }, []);

    const sendCommand = useCallback((cmd: WsCommand) => {
        wsApi.send(cmd);
    }, []);

    return {
        status,
        config,
        isLive,
        sendCommand
    };
};