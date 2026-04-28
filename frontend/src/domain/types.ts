export interface MotorConfig {
  motor_id: number;
  steps_per_rev: number;
  pulley_mm: number;
  ip_address: string;
}

export interface HubConfig {
  frame_width: number;
  frame_height: number;
  motors: MotorConfig[];
}

export interface MotorStatus {
  motor_id: number;
  enabled: boolean;
  current_steps: number;
  target_steps: number;
  speed_rps: number;
}

export interface SystemStatus {
  timestamp: number;
  position: { x: number; y: number };
  is_calibrated: boolean;
  motors: MotorStatus[];
}