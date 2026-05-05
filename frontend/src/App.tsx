import React, { useState, useEffect } from 'react';
import { hubApi } from './infrastructure/api';
import type { SystemStatus, HubConfig } from './domain/types';


  
const Sidebar = ({ activeTab, onTabChange }: { activeTab: string, onTabChange: (tab: string) => void }) => (
  <aside className="w-64 bg-slate-900 p-6 rounded-2xl border border-slate-800 flex flex-col gap-6">
    <div className="flex items-center gap-2 font-bold text-lg text-white">
      <div className="w-8 h-8 bg-indigo-600 rounded-lg shadow-[0_0_15px_rgba(79,70,229,0.4)]"></div>
      Motor Hub
    </div>
    <nav className="space-y-2">
      <div 
        onClick={() => onTabChange('dashboard')}
        className={`px-4 py-3 rounded-xl cursor-pointer font-medium transition-all ${
          activeTab === 'dashboard' ? 'bg-indigo-600/20 text-indigo-400 border border-indigo-500/30' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
        }`}
      >
        Dashboard
      </div>
      <div 
        onClick={() => onTabChange('config')}
        className={`px-4 py-3 rounded-xl cursor-pointer font-medium transition-all ${
          activeTab === 'config' ? 'bg-indigo-600/20 text-indigo-400 border border-indigo-500/30' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
        }`}
      >
        Configuration
      </div>
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
  const [activeTab, setActiveTab] = useState('dashboard');
  const [config, setConfig] = useState<HubConfig | null>(null);
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [speed, setSpeed] = useState(15.0);
  const [isLive, setIsLive] = useState(false);

  const [newWidth, setNewWidth] = useState('');
  const [newHeight, setNewHeight] = useState('');


  const refreshConfig = () => {
    hubApi.getConfig().then(data => {
        setConfig(data);
        setNewWidth(data.frame_width.toString());
        setNewHeight(data.frame_height.toString());
    }).catch(console.error);
  };

  useEffect(() => {
    refreshConfig();
  }, []);


  useEffect(() => {
    const timer = setInterval(() => {
      hubApi.getStatus().then(data => {
          setStatus(data);
          setIsLive(true);
      }).catch(() => setIsLive(false));
    }, 200);
    return () => clearInterval(timer);
  }, []);

  const handleUpdateDimensions = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
        await hubApi.updateDimensions(parseFloat(newWidth), parseFloat(newHeight));
        refreshConfig();
        setActiveTab('dashboard');
        alert("Configuration updated! Remember to re-calibrate.");
    } catch (err) { alert("Update failed"); }
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
      <Sidebar activeTab={activeTab} onTabChange={setActiveTab} />
      
      <main className="flex-1 space-y-6">
        <header className="flex justify-between items-center">
          <div>
            <h2 className="text-2xl font-bold text-white tracking-tight">
              {activeTab === 'dashboard' ? 'System Control' : 'System Configuration'}
            </h2>
            <p className="text-sm text-slate-500">{config?.frame_width}mm x {config?.frame_height}mm</p>
          </div>
          <div className={`px-4 py-1.5 rounded-full border text-xs font-bold flex items-center gap-2 ${
            isLive ? 'bg-emerald-500/10 border-emerald-500/50 text-emerald-500' : 'bg-red-500/10 border-red-500/50 text-red-500'
          }`}>
            <span className={`w-2 h-2 rounded-full ${isLive ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`}></span>
            {isLive ? 'ONLINE' : 'OFFLINE'}
          </div>
        </header>

        {activeTab === 'dashboard' ? (
          <div className="grid grid-cols-12 gap-6">
            <div className="col-span-7 space-y-4">
              <div className="aspect-square bg-slate-900/50 rounded-3xl border border-slate-800 relative cursor-crosshair overflow-hidden group shadow-2xl"
                onClick={(e) => {
                  if (!config) return;
                  const rect = e.currentTarget.getBoundingClientRect();
                  hubApi.moveTo(
                    parseFloat((((e.clientX - rect.left) / rect.width) * config.frame_width).toFixed(2)),
                    parseFloat((((e.clientY - rect.top) / rect.height) * config.frame_height).toFixed(2)),
                    speed
                  );
                }}>
                <div className="absolute inset-0 opacity-10" style={{ backgroundImage: 'linear-gradient(#4F46E5 1px, transparent 1px), linear-gradient(90deg, #4F46E5 1px, transparent 1px)', backgroundSize: '10% 10%' }}></div>
                <div className="absolute w-6 h-6 -ml-3 -mt-3 transition-all duration-300 ease-out" style={getPositionStyles()}>
                  <div className="absolute top-1/2 left-0 w-full h-0.5 bg-indigo-500"></div>
                  <div className="absolute left-1/2 top-0 w-0.5 h-full bg-indigo-500"></div>
                </div>
              </div>
              <div className="bg-slate-900 p-6 rounded-2xl border border-slate-800">
                <input type="range" min="1" max="50" step="0.5" value={speed} onChange={(e) => setSpeed(parseFloat(e.target.value))} className="w-full accent-indigo-500" />
              </div>
            </div>
            
            <div className="col-span-5 space-y-6">
              <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl">
                <div className="text-5xl text-white font-bold font-mono">
                  X {status?.position.x.toFixed(1)} Y {status?.position.y.toFixed(1)}
                </div>
              </div>
              <div className="grid grid-cols-3 gap-3">
                <button onClick={() => hubApi.stop()} className="py-4 bg-red-500/10 text-red-500 rounded-2xl font-bold">STOP</button>
                <button onClick={() => hubApi.home(speed)} className="py-4 bg-slate-800 text-white rounded-2xl font-bold">HOME</button>
                <button onClick={() => hubApi.calibrate(speed)} className="py-4 bg-slate-800 text-white rounded-2xl font-bold">CALIBRATE</button>
              </div>
              <div className="grid grid-cols-2 gap-4">
                {status?.motors.map(m => <MotorCard key={m.motor_id} motor={m} />)}
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-6 max-w-4xl">
            {/* Frame Settings Form */}
            <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl shadow-xl">
              <h3 className="text-lg font-bold text-white mb-6">Physical Frame Settings</h3>
              <form onSubmit={handleUpdateDimensions} className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <label className="text-sm text-slate-400 block mb-2">Width (mm)</label>
                    <input type="number" value={newWidth} onChange={(e) => setNewWidth(e.target.value)} className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-white" />
                  </div>
                  <div>
                    <label className="text-sm text-slate-400 block mb-2">Height (mm)</label>
                    <input type="number" value={newHeight} onChange={(e) => setNewHeight(e.target.value)} className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-white" />
                  </div>
                </div>
                <button type="submit" className="w-full py-4 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-2xl">Apply Changes</button>
              </form>
            </div>

            {/* Motor Inventory Info */}
            <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl shadow-xl">
              <h3 className="text-lg font-bold text-white mb-6">Active Motor Hardware</h3>
              <div className="space-y-4">
                {config?.motors.map((m) => (
                  <div key={m.motor_id} className="grid grid-cols-4 gap-4 p-4 bg-slate-950 border border-slate-800 rounded-2xl">
                    <div><div className="text-[10px] text-slate-500">ID</div><div className="text-white font-bold">#{m.motor_id}</div></div>
                    <div><div className="text-[10px] text-slate-500">IP ADDRESS</div><div className="text-indigo-400 text-sm font-mono">{m.ip_address}</div></div>
                    <div><div className="text-[10px] text-slate-500">RESOLUTION</div><div className="text-slate-300 text-sm">{m.steps_per_rev} steps</div></div>
                    <div><div className="text-[10px] text-slate-500">PULLEY</div><div className="text-slate-300 text-sm">{m.pulley_mm} mm</div></div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}