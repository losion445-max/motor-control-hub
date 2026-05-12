
export interface MotorHardwareConfig {
  motor_id: number;
  step_plus: number;
  step_minus: number;
  dir_plus: number;
  dir_minus: number;
  steps_per_rev: number;
  pulley_mm: number;
}


export interface KinematicsConfig {
  width: number;
  height: number;
  diameter: number;
  steps_per_rev: number;
}

export interface FullConfig {
  global: {
    kinematics: KinematicsConfig;
    motor_mapping: [number, number, number, number];
  };
  motors_hardware: MotorHardwareConfig[];
}


export interface MotorStatus {
  motor_id: number;
  enabled: boolean;
  infinite: boolean;
  current_steps: number;
  target_steps: number;
  speed_rps: number;
  wifi_rssi: number;
  online: boolean;
}

export interface SystemStatus {
  timestamp: number;
  position: { 
    x: number; 
    y: number; 
  };
  is_calibrated: boolean;
  motors: MotorStatus[];
}

export interface MoveRequest {
  x: number;
  y: number;
  speed: number;
}

export interface SingleMotorMoveRequest {
  motor_id: number;
  steps: number;
  speed: number;
}