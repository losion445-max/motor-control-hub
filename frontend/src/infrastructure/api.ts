import axios from 'axios';
import type { HubConfig, SystemStatus } from '../domain/types';

const api = axios.create({
  baseURL: '/api',
});

export const hubApi = {
  getConfig: () => api.get<HubConfig>('/config').then(res => res.data),
  
  getStatus: () => api.get<SystemStatus>('/status').then(res => res.data),

  moveTo: (x: number, y: number, speed: number) => 
    api.post('/move', { x, y, speed }),

  home: (speed: number) => 
    api.post('/home', { speed }),

  stop: () => 
    api.post('/stop'),

  calibrate: (speed: number) => 
    api.post('/calibrate', { speed })
};