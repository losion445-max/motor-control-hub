import axios from 'axios';
import type { 
  FullConfig, 
  SystemStatus, 
  MoveRequest, 
  SingleMotorMoveRequest,
  KinematicsConfig 
} from '../domain/types';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const hubApi = {
  
  motion: {
    moveTo: (data: MoveRequest) => 
      api.post('/move', data),

    home: (speed: number) => 
      api.post('/home', { speed }),

    stop: () => 
      api.post('/stop'),

    calibrate: (speed: number) => 
      api.post('/calibrate', { speed }),
  },

  config: {
    get: () => 
      api.get<FullConfig>('/config').then(res => res.data),

    update: (data: Partial<KinematicsConfig>) => 
      api.post('/config', data).then(res => res.data),
  },

  diag: {
    getStatus: () => 
      api.get<SystemStatus>('/status').then(res => res.data),

    moveMotor: (data: SingleMotorMoveRequest) => 
      api.post('/motors/move-single', data),

    setEnable: (enabled: boolean) => 
      api.post('/motors/enable', { enabled }),

    syncPosition: (x: number, y: number) => 
      api.post('/position/sync', { x, y }),
  }
};