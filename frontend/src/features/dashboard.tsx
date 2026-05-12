import React, { useMemo } from 'react';
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
  const { motor_mapping } = config.global;

  // 1. Сортируем моторы строго по маппингу из конфига (TOP_L, TOP_R, BTM_L, BTM_R)
  const orderedMotors = useMemo(() => {
    return motor_mapping.map(id => 
      status.motors.find(m => m.motor_id === id)
    ).filter((m): m is MotorStatus => !!m);
  }, [status.motors, motor_mapping]);

  // 2. Расчет позиции маркера с защитой от вылета за границы (Clamping)
  const markerPos = {
    x: Math.min(Math.max((status.position.x / width) * 100, 0), 100),
    y: Math.min(Math.max((status.position.y / height) * 100, 0), 100)
  };

  const handleCanvasClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    
    // Переводим пиксели клика в миллиметры рамы
    const clickX = ((e.clientX - rect.left) / rect.width) * width;
    const clickY = ((e.clientY - rect.top) / rect.height) * height;
    
    // Ограничиваем физическими размерами
    const clampedX = parseFloat(Math.min(Math.max(clickX, 0), width).toFixed(2));
    const clampedY = parseFloat(Math.min(Math.max(clickY, 0), height).toFixed(2));
    
    hubApi.motion.moveTo({ x: clampedX, y: clampedY, speed: 1 });
  };

  return (
    <div className="flex flex-col lg:grid lg:grid-cols-12 gap-4 font-mono pb-10">
      
      {/* ЛЕВАЯ КОЛОНКА: Визуализатор и Статус приводов */}
      <div className="flex flex-col gap-4 lg:col-span-8">
        
        <Panel title="WORKSPACE_VISUALIZER">
          {/* Контейнер с динамическим соотношением сторон и обработкой клика */}
          <div 
            className="relative overflow-hidden bg-[#010409] touch-none cursor-crosshair border border-[#30363d]"
            style={{ aspectRatio: `${width} / ${height}` }}
            onClick={handleCanvasClick}
          >
            {/* Сетка координат */}
            <div className="absolute inset-0 opacity-10 pointer-events-none" 
                 style={{ 
                   backgroundImage: 'linear-gradient(#58a6ff 1px, transparent 1px), linear-gradient(90deg, #58a6ff 1px, transparent 1px)', 
                   backgroundSize: '10% 10%' 
                 }} />
            
            {/* Направляющие оси */}
            <div className="absolute inset-0 pointer-events-none">
               <div className="absolute h-full w-px bg-[#58a6ff] opacity-10" style={{ left: `${markerPos.x}%` }} />
               <div className="absolute w-full h-px bg-[#58a6ff] opacity-10" style={{ top: `${markerPos.y}%` }} />
            </div>

            {/* Маркер текущего положения (Курсор) */}
            <div className="absolute w-6 h-6 -ml-3 -mt-3 transition-all duration-300 ease-out pointer-events-none"
                 style={{ left: `${markerPos.x}%`, top: `${markerPos.y}%` }}>
              <div className="absolute inset-0 border border-[#58a6ff] rounded-full animate-ping opacity-20" />
              <div className="absolute inset-1 border-2 border-[#58a6ff] rounded-full shadow-[0_0_15px_rgba(88,166,255,0.4)]" />
              <div className="absolute top-1/2 left-1/2 w-8 h-px bg-[#58a6ff] -translate-x-1/2 opacity-50" />
              <div className="absolute top-1/2 left-1/2 w-px h-8 bg-[#58a6ff] -translate-y-1/2 opacity-50" />
            </div>
          </div>
        </Panel>

        {/* Сетка моторов: 2 в ряд на мобиле, 4 на десктопе */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {orderedMotors.map((m, idx) => (
            <div key={m.motor_id} className="border border-[#30363d] p-3 bg-[#161b22] flex flex-col gap-1 shadow-sm">
              <div className="flex justify-between items-center">
                <span className="text-[8px] text-[#6e7681] font-bold uppercase tracking-tighter">NODE_{idx}</span>
                <span className="text-[8px] text-[#58a6ff]">ID:{m.motor_id}</span>
              </div>
              <div className="flex justify-between items-baseline">
                <span className={`text-[10px] font-black ${m.online ? 'text-[#3fb950]' : 'text-[#f85149]'}`}>
                  {m.online ? 'ONLINE' : 'OFFLINE'}
                </span>
                <span className="text-xs font-mono">{m.current_steps}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ПРАВАЯ КОЛОНКА: Телеметрия и Управление */}
      <div className="flex flex-col gap-4 lg:col-span-4">
        
        <Panel title="TELEMETRY_FEED">
          <div className="space-y-6 py-2">
            <div className="flex justify-between items-center">
              <ParameterDisplay label="X_AXIS" value={status.position.x.toFixed(1)} unit="MM" />
              <ParameterDisplay label="Y_AXIS" value={status.position.y.toFixed(1)} unit="MM" />
            </div>
            <div className="flex flex-col gap-2">
              <StatusBadge label="KINEMATICS_SYNC" active={status.is_calibrated} pulse />
              <StatusBadge label="ALL_MOTORS_LIVE" active={status.motors.every(m => m.online)} />
            </div>
          </div>
        </Panel>

        <Panel title="PRIMARY_CONTROLS">
          <div className="flex flex-col gap-2">
            <div className="grid grid-cols-2 gap-2">
              <TerminalButton label="HOME_ALL" onClick={() => hubApi.motion.home(20)} />
              <TerminalButton label="CALIBRATE" variant="warning" onClick={() => hubApi.motion.calibrate(10)} />
            </div>
            <TerminalButton 
              label="EMERGENCY_STOP" 
              variant="danger" 
              className="py-4 shadow-[0_0_20px_rgba(248,81,73,0.2)]"
              onClick={() => hubApi.motion.stop()} 
            />
          </div>
        </Panel>

      </div>
    </div>
  );
};