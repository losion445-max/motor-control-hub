import React from 'react';
import { hubApi } from '../infrastructure/api';
import { Panel } from '../components/panel';
import { TerminalButton } from '../components/terminalButton';
import { ParameterDisplay } from '../components/parameterDisplay';
import { StatusBadge } from '../components/statusBadge';
import type { SystemStatus, FullConfig, MotorStatus } from '../domain/types';

interface DashboardProps {
  status: SystemStatus;
  config: FullConfig;
}

export const DashboardView: React.FC<DashboardProps> = ({ status, config }) => {
  const { width, height } = config.global.kinematics;

  const handleCanvasClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = parseFloat((((e.clientX - rect.left) / rect.width) * width).toFixed(2));
    const y = parseFloat((((e.clientY - rect.top) / rect.height) * height).toFixed(2));
    
    hubApi.motion.moveTo({ x, y, speed: 1 });
  };

  return (
    <div className="flex flex-col lg:grid lg:grid-cols-12 gap-4 h-full font-mono pb-8 lg:pb-0">
      
      <div className="flex flex-col gap-4 lg:col-span-8 order-1">
        <Panel 
          title="WORKSPACE_VISUALIZER" 
          className="aspect-square lg:aspect-video relative overflow-hidden bg-[#010409] touch-none" 
          onClick={handleCanvasClick}
        >
          <div className="absolute inset-0 opacity-10" 
               style={{ 
                 backgroundImage: 'linear-gradient(#58a6ff 1px, transparent 1px), linear-gradient(90deg, #58a6ff 1px, transparent 1px)', 
                 backgroundSize: '10% 10%' 
               }} />
          
          <div className="absolute w-8 h-8 -ml-4 -mt-4 transition-all duration-500 cubic-bezier(0.4, 0, 0.2, 1)"
               style={{ 
                 left: `${(status.position.x / width) * 100}%`, 
                 top: `${(status.position.y / height) * 100}%` 
               }}>
            <div className="absolute inset-0 border-2 border-[#58a6ff] rounded-full animate-pulse shadow-[0_0_15px_rgba(88,166,255,0.5)]" />
            <div className="h-full w-px bg-[#58a6ff] mx-auto opacity-50" />
            <div className="w-full h-px bg-[#58a6ff] -mt-4 opacity-50" />
          </div>
        </Panel>

        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
          {status.motors.map((m: MotorStatus) => (
            <div key={m.motor_id} className="border border-[#30363d] p-2 bg-[#161b22] flex justify-between items-center shadow-sm">
              <span className="text-[9px] text-[#6e7681] font-bold">M0{m.motor_id}</span>
              <span className={`text-[10px] font-bold ${m.online ? 'text-[#3fb950]' : 'text-[#f85149]'}`}>
                {m.current_steps} <span className="font-normal opacity-50 text-[8px]">ST</span>
              </span>
            </div>
          ))}
        </div>
      </div>

      <div className="flex flex-col gap-4 lg:col-span-4 order-2">
        <Panel title="TELEMETRY_FEED">
          <div className="flex flex-col gap-6">
            <div className="flex justify-between items-center">
              <ParameterDisplay label="POS_X" value={status.position.x.toFixed(2)} unit="mm" />
              <ParameterDisplay label="POS_Y" value={status.position.y.toFixed(2)} unit="mm" />
            </div>
            <div className="grid grid-cols-2 lg:flex lg:flex-col gap-2">
              <StatusBadge label="CALIBRATED" active={status.is_calibrated} pulse />
              <StatusBadge label="LINK_LIVE" active={status.motors.some(m => m.online)} />
            </div>
          </div>
        </Panel>

        <Panel title="PRIMARY_CONTROLS">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <TerminalButton label="EMERGENCY_STOP" variant="danger" onClick={() => hubApi.motion.stop()} />
            <TerminalButton label="HOME_ALL" onClick={() => hubApi.motion.home(20)} />
            <TerminalButton label="CALIBRATE" variant="warning" onClick={() => hubApi.motion.calibrate(10)} />
            <TerminalButton label="REFRESH" onClick={() => {}} />
          </div>
        </Panel>
      </div>
    </div>
  );
};