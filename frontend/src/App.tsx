import React, { useState, useEffect } from 'react';
import { hubApi } from './infrastructure/api';
import type { SystemStatus, HubConfig } from './domain/types';

const Sidebar = () => (
  <aside className="w-64 bg-slate-900 p-6 rounded-2xl border border-slate-800 flex flex-col gap-6">
    <div className="flex items-center gap-2 font-bold text-lg text-white">
      <div className="w-8 h-8 bg-indigo-600 rounded-lg shadow-[0_0_15px_rgba(79,70,229,0.4)]"></div>
      Motor Hub
    </div>
    <nav className="space-y-2">
      <div className="px-4 py-3 bg-indigo-600/20 text-indigo-400 border border-indigo-500/30 rounded-xl cursor-pointer font-medium">Dashboard</div>
      <div className="px-4 py-3 text-slate-400 hover:bg-slate-800 hover:text-white rounded-xl transition-all cursor-pointer">Configuration</div>
    </nav>
  </aside>
);

const MotorCard = ({ motor }: { motor: any }) => (
  <div className="bg-slate-900 p-4 rounded-xl border border-slate-800 hover:border-slate-700 transition-colors">
    <div className="text-[10px] uppercase tracking-wider text-slate-500 mb-1">Motor {motor.motor_id}</div>
    <div className="font-mono text-lg text-white font-bold leading-none">{motor.current_steps} <span className="text-[10px] text-slate-500 font-normal">steps</span></div>
    <div className="mt-3 flex items-center justify-between">
      <div className={`px-2 py-0.5 rounded text-[10px] font-bold ${motor.enabled ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'}`}>
        {motor.enabled ? 'ACTIVE' : 'OFF'}
      </div>
      <div className="text-[10px] text-slate-400">{motor.speed_rps} rps</div>
    </div>
  </div>
);


export default function Dashboard() {
  const [config, setConfig] = useState<HubConfig | null>(null);
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [speed, setSpeed] = useState(15.0); // Дефолтная скорость из API доков
  const [isLive, setIsLive] = useState(false);

  useEffect(() => {
    hubApi.getConfig().then(setConfig).catch(console.error);
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      hubApi.getStatus()
        .then(data => {
          setStatus(data);
          setIsLive(true);
        })
        .catch(() => setIsLive(false));
    }, 200);
    return () => clearInterval(timer);
  }, []);

  const handleCanvasClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!config) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const pctX = (e.clientX - rect.left) / rect.width;
    const pctY = (e.clientY - rect.top) / rect.height;

    const targetX = parseFloat((pctX * config.frame_width).toFixed(2));
    const targetY = parseFloat((pctY * config.frame_height).toFixed(2));

    hubApi.moveTo(targetX, targetY, speed);
  };

  const getPositionStyles = () => {
    if (!status || !config) return { left: '50%', top: '50%' };
    return {
      left: `${(status.position.x / config.frame_width) * 100}%`,
      top: `${(status.position.y / config.frame_height) * 100}%`,
    };
  };

  return (
    <div className="min-h-screen bg-[#0F172A] text-slate-200 p-6 flex gap-6 font-sans">
      <Sidebar />
      
      <main className="flex-1 space-y-6">
        <header className="flex justify-between items-center">
          <div>
            <h2 className="text-2xl font-bold text-white tracking-tight">System Control</h2>
            <p className="text-sm text-slate-500">Dimensions: {config?.frame_width}m x {config?.frame_height}m</p>
          </div>
          <div className={`px-4 py-1.5 rounded-full border text-xs font-bold flex items-center gap-2 transition-all ${
            isLive ? 'bg-emerald-500/10 border-emerald-500/50 text-emerald-500' : 'bg-red-500/10 border-red-500/50 text-red-500'
          }`}>
            <span className={`w-2 h-2 rounded-full ${isLive ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`}></span>
            {isLive ? 'HUB ONLINE' : 'HUB OFFLINE'}
          </div>
        </header>

        <div className="grid grid-cols-12 gap-6">
          <div className="col-span-7 space-y-4">
            <div 
              className="aspect-square bg-slate-900/50 rounded-3xl border border-slate-800 relative cursor-crosshair overflow-hidden group shadow-2xl"
              onClick={handleCanvasClick}
            >
              <div className="absolute inset-0 opacity-20" style={{ 
                backgroundImage: 'linear-gradient(#4F46E5 1px, transparent 1px), linear-gradient(90deg, #4F46E5 1px, transparent 1px)', 
                backgroundSize: '10% 10%' 
              }}></div>
              
              <div 
                className="absolute w-6 h-6 -ml-3 -mt-3 transition-all duration-300 ease-out"
                style={getPositionStyles()}
              >
                <div className="absolute top-1/2 left-0 w-full h-0.5 bg-indigo-500"></div>
                <div className="absolute left-1/2 top-0 w-0.5 h-full bg-indigo-500"></div>
                <div className="absolute inset-0 rounded-full border-2 border-indigo-500 animate-ping opacity-25"></div>
              </div>
            </div>
            
            <div className="bg-slate-900/80 backdrop-blur-md p-6 rounded-2xl border border-slate-800">
              <div className="flex justify-between mb-4">
                <label className="text-sm font-semibold text-slate-400">Movement Speed</label>
                <span className="font-mono text-indigo-400 font-bold">{speed} units/s</span>
              </div>
              <input title="input" 
                type="range" min="1" max="50" step="0.5" value={speed} 
                onChange={(e) => setSpeed(parseFloat(e.target.value))} 
                className="w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-indigo-500"
              />
            </div>
          </div>
          
          <div className="col-span-5 space-y-6">
            <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl shadow-xl relative overflow-hidden">
               <div className="absolute top-0 right-0 p-4 opacity-10">
                 <div className="w-20 h-20 bg-indigo-500 blur-[50px]"></div>
               </div>
              <div className="text-xs font-bold text-slate-500 uppercase tracking-[0.2em] mb-2">Live Telemetry</div>
              <div className="font-mono text-5xl text-white font-bold tracking-tighter flex gap-4">
                <span className="opacity-50 text-2xl mt-auto pb-1">X</span> {status?.position.x.toFixed(2) || '0.00'}
                <span className="opacity-50 text-2xl mt-auto pb-1 ml-4">Y</span> {status?.position.y.toFixed(2) || '0.00'}
              </div>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <button onClick={() => hubApi.stop()} className="group py-4 bg-red-500/10 hover:bg-red-500 border border-red-500/20 hover:border-red-500 text-red-500 hover:text-white font-black rounded-2xl transition-all duration-200 uppercase text-xs tracking-widest">
                Stop
              </button>
              <button onClick={() => hubApi.home(speed)} className="py-4 bg-slate-800 hover:bg-indigo-600 text-white font-bold rounded-2xl transition-all duration-200 uppercase text-xs tracking-widest">
                Home
              </button>
              <button onClick={() => hubApi.calibrate(speed)} className="py-4 bg-slate-800 hover:bg-amber-600 text-white font-bold rounded-2xl transition-all duration-200 uppercase text-xs tracking-widest">
                Calibrate
              </button>
            </div>

            <div className="grid grid-cols-2 gap-4">
              {status?.motors.map(m => <MotorCard key={m.motor_id} motor={m} />) || 
               [0,1,2,3].map(i => <div key={i} className="h-24 bg-slate-900/50 animate-pulse rounded-xl border border-slate-800"></div>)
              }
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}