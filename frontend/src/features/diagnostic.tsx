import React, { useState } from 'react';
import { hubApi } from '../infrastructure/api';
import { Panel } from '../components/panel';
import { TerminalButton } from '../components/terminalButton';
import type { MotorStatus } from '../domain/types';

export const DiagnosticView: React.FC<{ motors: MotorStatus[] }> = ({ motors }) => {
  const [manualCoords, setManualCoords] = useState({ x: 0, y: 0 });

  const handleSync = () => {
    hubApi.diag.syncPosition(manualCoords.x, manualCoords.y);
  };

  return (
    <div className="space-y-6 pb-10">
      <Panel title="PER_MOTOR_HARDWARE_TESTS">
        <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
          {motors.map((motor) => (
            <div 
              key={motor.motor_id} 
              className={`border p-4 bg-[#0d1117] transition-all ${
                motor.online ? 'border-[#30363d]' : 'border-[#f8514933] opacity-80'
              }`}
            >
              <div className="flex justify-between items-center mb-4 border-b border-[#30363d] pb-2">
                <span className="font-bold text-[#58a6ff] tracking-tighter">MOT_0{motor.motor_id}</span>
                <span className={`text-[9px] font-mono px-2 py-0.5 border ${
                  motor.online ? 'border-[#3fb95033] text-[#3fb950]' : 'border-[#f8514933] text-[#f85149]'
                }`}>
                  {motor.online ? `${motor.wifi_rssi} DBM` : 'OFFLINE'}
                </span>
              </div>
              
              <div className="space-y-4">
                <div className="flex justify-between text-[10px] uppercase tracking-widest">
                  <span className="text-[#6e7681]">STATUS:</span>
                  <span className={motor.enabled ? 'text-[#3fb950]' : 'text-[#f85149]'}>
                    {motor.enabled ? 'ENABLED' : 'DISABLED'}
                  </span>
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  <TerminalButton 
                    label="+100 ST" 
                    disabled={!motor.online}
                    onClick={() => hubApi.diag.moveMotor({motor_id: motor.motor_id, steps: 100, speed: 5})} 
                  />
                  <TerminalButton 
                    label="-100 ST" 
                    disabled={!motor.online}
                    onClick={() => hubApi.diag.moveMotor({motor_id: motor.motor_id, steps: -100, speed: 5})} 
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </Panel>

      <Panel title="FORCE_POSITION_OVERRIDE">
        <div className="flex flex-col sm:flex-row gap-4 sm:items-end max-w-full lg:max-w-md">
          <div className="flex flex-row sm:flex-col gap-4 sm:gap-1 flex-1">
            <div className="flex-1 space-y-1">
              <span className="text-[9px] text-[#6e7681] font-bold">X_COORD</span>
              <input 
                type="number" 
                inputMode="decimal"
                value={manualCoords.x}
                onChange={(e) => setManualCoords(prev => ({ ...prev, x: parseFloat(e.target.value) || 0 }))}
                className="w-full bg-[#010409] border border-[#30363d] p-3 sm:p-2 text-sm text-[#58a6ff] outline-none focus:border-[#58a6ff] transition-colors" 
                placeholder="0.0" 
              />
            </div>
            
            <div className="flex-1 space-y-1">
              <span className="text-[9px] text-[#6e7681] font-bold">Y_COORD</span>
              <input 
                type="number" 
                inputMode="decimal"
                value={manualCoords.y}
                onChange={(e) => setManualCoords(prev => ({ ...prev, y: parseFloat(e.target.value) || 0 }))}
                className="w-full bg-[#010409] border border-[#30363d] p-3 sm:p-2 text-sm text-[#58a6ff] outline-none focus:border-[#58a6ff] transition-colors" 
                placeholder="0.0" 
              />
            </div>
          </div>
          
          <TerminalButton 
            label="SYNC_COORDS" 
            variant="warning" 
            className="w-full sm:w-auto mt-2 sm:mt-0"
            onClick={handleSync} 
          />
        </div>
      </Panel>
    </div>
  );
};