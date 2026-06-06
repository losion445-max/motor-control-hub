import type { SystemStatus, FullConfig, KinematicsConfig } from './types';

export type WsCommand =
    | { type: 'MOVE'; payload: { x: number; y: number; speed: number } }
    | { type: 'STOP' }
    | { type: 'HOME'; payload: { speed: number } }
    | { type: 'CALIBRATE'; payload: { speed: number } }

    | { type: 'MOTORS_ENABLE'; payload: { enabled: boolean } }
    | { type: 'MOVE_SINGLE'; payload: { motor_id: number; steps: number; speed: number } }
    | { type: 'SYNC_POSITION'; payload: { x: number; y: number } }

    | { type: 'GET_CONFIG' }
    | { type: 'UPDATE_CONFIG'; payload: Partial<KinematicsConfig & { motor_mapping: number[] }> };

export type WsEvent =
    | { type: 'STATUS_UPDATE'; payload: SystemStatus }
    | { type: 'CONFIG_DATA'; payload: FullConfig }
    | { type: 'LOG'; payload: { level: 'info' | 'warn' | 'error'; message: string; timestamp: number } }
    | { type: 'ERROR'; payload: { message: string } };