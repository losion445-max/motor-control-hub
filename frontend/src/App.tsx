import { useState, useEffect } from 'react';
import { useHub } from './hooks/useHub';

import { DashboardView } from './features/dashboard';
import { ConfigView } from './features/config';
import { DiagnosticView } from './features/diagnostic'

import { StatusBadge } from './components/statusBadge';
import type { SystemStatus, FullConfig } from './domain/types';

type Tab = 'dash' | 'config' | 'diag';

export default function App() {
const [activeTab, setActiveTab] = useState<Tab>('dash');
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [uptime, setUptime] = useState(0);

  const { status: rawStatus, config: rawConfig, isLive, refreshConfig } = useHub();

  // --- STRICT MOCK DATA ---
  // Используем интерфейс вместо any. Теперь TS проверит каждое поле.
  const status: SystemStatus = rawStatus ?? {
    timestamp: Date.now(),
    position: { x: 125.5, y: 80.2 },
    is_calibrated: true,
    motors: [
      { motor_id: 1, enabled: true, infinite: false, current_steps: 1024, target_steps: 1024, speed_rps: 0, wifi_rssi: -52, online: true },
      { motor_id: 2, enabled: true, infinite: false, current_steps: 850, target_steps: 850, speed_rps: 0, wifi_rssi: -48, online: true },
      { motor_id: 3, enabled: true, infinite: false, current_steps: 1200, target_steps: 1200, speed_rps: 0, wifi_rssi: -55, online: true },
      { motor_id: 4, enabled: false, infinite: false, current_steps: 0, target_steps: 0, speed_rps: 0, wifi_rssi: 0, online: false },
    ],
  };

  const config: FullConfig = rawConfig ?? {
    global: {
      kinematics: { 
        width: 500, 
        height: 400, 
        diameter: 20.0, 
        steps_per_rev: 200 
      },
      motor_mapping: [2, 3, 1, 4]
    },
    motors_hardware: [
      { motor_id: 1, step_plus: 12, step_minus: 13, dir_plus: 14, dir_minus: 15, steps_per_rev: 200, pulley_mm: 20 },
      { motor_id: 2, step_plus: 16, step_minus: 17, dir_plus: 18, dir_minus: 19, steps_per_rev: 200, pulley_mm: 20 },
      { motor_id: 3, step_plus: 21, step_minus: 22, dir_plus: 23, dir_minus: 24, steps_per_rev: 200, pulley_mm: 20 },
      { motor_id: 4, step_plus: 25, step_minus: 26, dir_plus: 27, dir_minus: 28, steps_per_rev: 200, pulley_mm: 20 },
    ]
  };

  useEffect(() => {
    const timer = setInterval(() => {
      setUptime(prev => prev + 60000);
    }, 60000);
    return () => clearInterval(timer);
  }, []);

  // if (!rawConfig || !rawStatus) {
  //   return (
  //     <div className="flex min-h-screen flex-col items-center justify-center bg-[#0d1117] font-mono uppercase tracking-widest p-4 text-center">
  //       <div className="mb-4 animate-pulse text-lg md:text-xl text-[#58a6ff]">INITIALIZING_CORE_SYSTEMS...</div>
  //       <div className="w-full max-w-xs h-px bg-[#30363d] overflow-hidden relative">
  //         <div className="absolute inset-0 bg-[#58a6ff] animate-[progress_2s_infinite]" />
  //       </div>
  //     </div>
  //   );
  // }
  

  return (
    <div className="flex min-h-screen flex-col lg:flex-row bg-[#0d1117] font-mono text-[#e6edf3] selection:bg-[#58a6ff33]">
      
      <header className="flex lg:hidden items-center justify-between p-4 border-b border-[#30363d] bg-[#010409] z-50">
        <h1 className="text-sm font-black tracking-tighter text-[#58a6ff]">
          MOTOR_HUB <span className="opacity-50">v1.0.4</span>
        </h1>
        <button 
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className="px-3 py-1 border border-[#30363d] text-[10px] font-bold active:bg-[#58a6ff11]"
        >
          {isMenuOpen ? 'CLOSE_MENU' : 'OPEN_MENU'}
        </button>
      </header>

      <aside className={`
        ${isMenuOpen ? 'flex' : 'hidden'} 
        lg:flex fixed lg:relative inset-0 z-40 lg:z-auto
        w-full lg:w-72 flex-col border-r border-[#30363d] bg-[#010409] p-6
      `}>
        <div className="hidden lg:block mb-10">
          <h1 className="text-xl font-black tracking-tighter text-[#58a6ff]">
            MOTOR_HUB <span className="font-normal opacity-50 text-[10px]">v1.0.4</span>
          </h1>
          <div className="mt-1 tracking-[0.3em] text-[9px] text-[#8b949e]">ROSTOV_ON_DON // LAB</div>
        </div>

        <nav className="flex flex-1 flex-col gap-2 mt-12 lg:mt-0">
          {[
            { id: 'dash', label: 'Dashboard', icon: '01' },
            { id: 'config', label: 'Kinematics', icon: '02' },
            { id: 'diag', label: 'Diagnostics', icon: '03' },
          ].map((item) => (
            <button
              key={item.id}
              onClick={() => {
                setActiveTab(item.id as Tab);
                setIsMenuOpen(false);
              }}
              className={`
                group flex items-center gap-4 border p-4 lg:p-3 text-sm lg:text-xs font-bold uppercase transition-all duration-150
                ${
                  activeTab === item.id
                    ? 'border-[#58a6ff] bg-[#58a6ff11] text-[#58a6ff] shadow-[0_0_15px_rgba(88,166,255,0.1)]'
                    : 'border-transparent text-[#8b949e] hover:border-[#30363d] hover:text-[#e6edf3]'
                }
              `}
            >
              <span className="opacity-40 group-hover:opacity-100">[{item.icon}]</span>
              {item.label}
            </button>
          ))}
        </nav>

        <div className="space-y-3 border-t border-[#30363d] pt-6 mt-4">
          <StatusBadge label={isLive || !rawStatus ? 'LINK_STABLE' : 'LINK_LOST'} active={isLive || !rawStatus} pulse />
          <div className="leading-relaxed text-[9px] text-[#6e7681]">
            SYSTEM_UPTIME: {(uptime / 1000 / 60).toFixed(1)}M <br />
            API_VERSION: 2.4.0-GO
          </div>
        </div>
      </aside>

      <main className="flex h-screen flex-1 flex-col overflow-hidden bg-[#0d1117]">
        <header className="hidden sm:flex h-16 shrink-0 items-center justify-between border-b border-[#30363d] bg-[#161b22]/30 px-4 lg:px-8 backdrop-blur-md">
          <div className="flex items-center gap-4 lg:gap-8">
            <div className="font-bold tracking-widest text-[9px] lg:text-[10px] text-[#8b949e]">
              VIEW: <span className="text-[#e6edf3]">{activeTab.toUpperCase()}</span>
            </div>
            <div className="h-4 w-px bg-[#30363d]" />
            <div className="font-bold tracking-widest text-[9px] lg:text-[10px] text-[#8b949e]">
              CALIB: <span className={status.is_calibrated ? "text-[#3fb950]" : "text-[#f85149]"}>
                {status.is_calibrated ? "READY" : "REQUIRED"}
              </span>
            </div>
          </div>

          <div className="text-[9px] lg:text-[10px] text-[#6e7681]">
            TS: {new Date(status.timestamp).toLocaleTimeString()}
          </div>
        </header>

        <div className="custom-scrollbar flex-1 overflow-y-auto p-4 lg:p-8">
          <div className="mx-auto max-w-6xl">
            {activeTab === 'dash' && <DashboardView status={status} config={config} />}
            {activeTab === 'config' && <ConfigView config={config} onRefresh={refreshConfig} />}
            {activeTab === 'diag' && <DiagnosticView motors={status.motors} />}
          </div>
        </div>
      </main>

      <style>{`
        @keyframes progress {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
        .custom-scrollbar::-webkit-scrollbar { width: 4px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: #30363d; border-radius: 10px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #58a6ff; }
      `}</style>
    </div>
  );
}