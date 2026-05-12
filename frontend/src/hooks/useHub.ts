import { useState, useEffect, useCallback } from 'react';
import { hubApi } from '../infrastructure/api';
import type { SystemStatus, FullConfig } from '../domain/types';

export const useHub = () => {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [config, setConfig] = useState<FullConfig | null>(null);
  const [isLive, setIsLive] = useState(false);

  const refreshConfig = useCallback(async () => {
    try {
      const data = await hubApi.config.get();
      setConfig(data);
    } catch (e) {
      console.error("CONFIG_FETCH_ERROR:", e);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;

    const loadInitialData = async () => {
      await refreshConfig();
    };

    loadInitialData();

    const timer = setInterval(async () => {
      try {
        const data = await hubApi.diag.getStatus();
        if (isMounted) {
          setStatus(data);
          setIsLive(true);
        }
      } catch (err) {
        if (isMounted) setIsLive(false);
        console.error("Times error:", err)
      }
    }, 250);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [refreshConfig]);

  return { status, config, isLive, refreshConfig };
};