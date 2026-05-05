import React, { useState, useEffect } from 'react';
import { hubApi } from './infrastructure/api';
import type { SystemStatus, HubConfig } from './domain/types';

const Sidebar = ({ activeTab, onTabChange }: { activeTab: string; onTabChange: (tab: string) => void }) => (
  <aside className="w-[220px] shrink-0 flex flex-col gap-2">
    <div className="flex items-center gap-2 font-semibold text-sm text-[#e6edf3] px-2 pb-4 border-b border-[#30363d] mb-1">
      <div className="w-5 h-5 bg-[#1f6feb] rounded flex items-center justify-center">
        <svg width="12" height="12" viewBox="0 0 12 12" fill="white"><circle cx="6" cy="6" r="4"/></svg>
      </div>
      Motor Hub
    </div>

    {[
      { id: 'dashboard', label: 'Dashboard' },
      { id: 'config',    label: 'Configuration' },
    ].map(({ id, label }) => (
      <div
        key={id}
        onClick={() => onTabChange(id)}
        className={`flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer text-sm select-none transition-colors
          ${activeTab === id
            ? 'bg-[#1f6feb26] text-[#58a6ff] font-semibold'
            : 'text-[#8b949e] hover:bg-[#6e768126] hover:text-[#e6edf3]'
          }`}
      >
        {label}
      </div>
    ))}
  </aside>
);

const GhBox = ({ header, children, noPadding = false }: {
  header: React.ReactNode; children: React.ReactNode; noPadding?: boolean;
}) => (
  <div className="bg-[#161b22] border border-[#30363d] rounded-md overflow-hidden">
    <div className="px-4 py-2.5 border-b border-[#30363d] font-semibold text-sm text-[#e6edf3] flex items-center gap-2">
      {header}
    </div>
    <div className={noPadding ? '' : 'p-4'}>{children}</div>
  </div>
);

const MotorCard = ({ motor }: { motor: any }) => (
  <div className="bg-[#010409] border border-[#21262d] rounded-md px-3 py-2.5">
    <div className="text-[11px] text-[#6e7681] mb-1 font-mono">motor_{motor.motor_id}</div>
    <div className="font-mono text-base font-semibold text-[#e6edf3]">
      {motor.current_steps} <span className="text-[10px] text-[#6e7681] font-normal">steps</span>
    </div>
    <div className="flex justify-between items-center mt-2">
      <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded-full
        ${motor.enabled
          ? 'bg-[#2ea04326] text-[#3fb950]'
          : 'bg-[#f8514926] text-[#f85149]'
        }`}>
        {motor.enabled ? 'ACTIVE' : 'OFF'}
      </span>
      <span className="text-[11px] font-mono text-[#8b949e]">{motor.speed_rps} rps</span>
    </div>
  </div>
);

export default function Dashboard() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [config, setConfig]       = useState<HubConfig | null>(null);
  const [status, setStatus]       = useState<SystemStatus | null>(null);
  const [speed, setSpeed]         = useState(15.0);
  const [isLive, setIsLive]       = useState(false);
  const [newWidth, setNewWidth]   = useState('');
  const [newHeight, setNewHeight] = useState('');

  const refreshConfig = () => {
    hubApi.getConfig().then(data => {
      setConfig(data);
      setNewWidth(data.frame_width.toString());
      setNewHeight(data.frame_height.toString());
    }).catch(console.error);
  };


  useEffect(() => { refreshConfig(); }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      hubApi.getStatus()
        .then(data => { setStatus(data); setIsLive(true); })
        .catch(() => setIsLive(false));
    }, 200);
    return () => clearInterval(timer);
  }, []);

  const handleUpdateDimensions = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await hubApi.updateDimensions(parseFloat(newWidth), parseFloat(newHeight));
      refreshConfig();
      setActiveTab('dashboard');
      alert('Configuration updated! Remember to re-calibrate.');
    } catch { alert('Update failed'); }
  };

  const getPositionStyles = () => {
    if (!status || !config) return { left: '50%', top: '50%' };
    return {
      left: `${(status.position.x / config.frame_width)  * 100}%`,
      top:  `${(status.position.y / config.frame_height) * 100}%`,
    };
  };

  return (
    <div className="min-h-screen bg-[#0d1117] text-[#e6edf3] p-4 flex gap-4 font-sans text-sm">
      <Sidebar activeTab={activeTab} onTabChange={setActiveTab} />

      <main className="flex-1 flex flex-col gap-4 min-w-0">

        {/* ── Header ── */}
        <header className="flex justify-between items-start pb-4 border-b border-[#30363d]">
          <div>
            <div className="text-xl font-semibold text-[#e6edf3]">
              {activeTab === 'dashboard' ? 'System Control' : 'System Configuration'}
            </div>
            <div className="text-xs text-[#8b949e] mt-0.5 font-mono">
              {config?.frame_width}mm × {config?.frame_height}mm
            </div>
          </div>
          <div className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border
            ${isLive
              ? 'bg-[#2ea04326] border-[#3fb95066] text-[#3fb950]'
              : 'bg-[#f8514926] border-[#f8514966] text-[#f85149]'
            }`}>
            <span className={`w-2 h-2 rounded-full ${isLive ? 'bg-[#3fb950] animate-pulse' : 'bg-[#f85149]'}`} />
            {isLive ? 'ONLINE' : 'OFFLINE'}
          </div>
        </header>

        {/* ══════════════ DASHBOARD ══════════════ */}
        {activeTab === 'dashboard' && (
          <div className="grid grid-cols-[1fr_300px] gap-4">

            {/* Left column */}
            <div className="flex flex-col gap-4">
              <GhBox header="Workspace Canvas">
                <div
                  className="aspect-square bg-[#010409] rounded-md relative cursor-crosshair overflow-hidden"
                  onClick={(e) => {
                    if (!config) return;
                    const rect = e.currentTarget.getBoundingClientRect();
                    hubApi.moveTo(
                      parseFloat((((e.clientX - rect.left) / rect.width)  * config.frame_width).toFixed(2)),
                      parseFloat((((e.clientY - rect.top)  / rect.height) * config.frame_height).toFixed(2)),
                      speed,
                    );
                  }}
                >
                  {/* Grid overlay */}
                  <div className="absolute inset-0" style={{
                    backgroundImage: 'linear-gradient(rgba(88,166,255,0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(88,166,255,0.06) 1px, transparent 1px)',
                    backgroundSize: '10% 10%',
                  }} />
                  {/* Crosshair */}
                  <div
                    className="absolute w-5 h-5 -ml-2.5 -mt-2.5 transition-all duration-300 ease-out"
                    style={getPositionStyles()}
                  >
                    <div className="absolute top-1/2 left-0 right-0 h-px bg-[#58a6ff]" />
                    <div className="absolute left-1/2 top-0 bottom-0 w-px bg-[#58a6ff]" />
                    <div className="absolute top-1/2 left-1/2 w-1.5 h-1.5 -translate-x-1/2 -translate-y-1/2 bg-[#58a6ff] rounded-full shadow-[0_0_6px_#58a6ff]" />
                  </div>
                </div>
              </GhBox>

              <GhBox header="Speed Control">
                <div className="flex items-center gap-3">
                  <span className="text-xs text-[#8b949e] whitespace-nowrap">Speed</span>
                  <input
                    type="range" min="1" max="50" step="0.5" value={speed}
                    onChange={(e) => setSpeed(parseFloat(e.target.value))}
                    className="flex-1 accent-[#58a6ff]"
                  />
                  <span className="font-mono text-xs text-[#58a6ff] min-w-[52px] text-right">
                    {speed.toFixed(1)} rps
                  </span>
                </div>
              </GhBox>
            </div>

            {/* Right column */}
            <div className="flex flex-col gap-4">

              <GhBox header="Position">
                <div className="font-mono flex flex-col gap-1.5">
                  {(['X', 'Y'] as const).map((axis) => (
                    <div key={axis} className="flex items-baseline gap-2">
                      <span className="text-[#6e7681] text-[11px] w-4">{axis}</span>
                      <span className="text-[#58a6ff] text-[22px] font-semibold leading-none">
                        {(axis === 'X' ? status?.position.x : status?.position.y)?.toFixed(1) ?? '—'}
                      </span>
                      <span className="text-[#6e7681] text-[11px]">mm</span>
                    </div>
                  ))}
                </div>
              </GhBox>

              <GhBox header="Commands">
                <div className="grid grid-cols-3 gap-2">
                  <button
                    onClick={() => hubApi.stop()}
                    className="py-1.5 px-3 rounded-md text-[13px] font-medium cursor-pointer bg-[#f8514926] border border-[#f8514966] text-[#f85149] hover:bg-[#f8514940] transition-colors"
                  >
                    STOP
                  </button>
                  <button
                    onClick={() => hubApi.home(speed)}
                    className="py-1.5 px-3 rounded-md text-[13px] font-medium cursor-pointer bg-[#21262d] border border-[#30363d] text-[#e6edf3] hover:bg-[#30363d] hover:border-[#8b949e] transition-colors"
                  >
                    HOME
                  </button>
                  <button
                    onClick={() => hubApi.calibrate(speed)}
                    className="py-1.5 px-3 rounded-md text-[13px] font-medium cursor-pointer bg-[#21262d] border border-[#30363d] text-[#e6edf3] hover:bg-[#30363d] hover:border-[#8b949e] transition-colors"
                  >
                    CAL
                  </button>
                </div>
              </GhBox>

              <GhBox header="Motors" noPadding>
                <div className="grid grid-cols-2 gap-2 p-2">
                  {status?.motors.map(m => <MotorCard key={m.motor_id} motor={m} />)}
                </div>
              </GhBox>

            </div>
          </div>
        )}

        {/* ══════════════ CONFIG ══════════════ */}
        {activeTab === 'config' && (
          <div className="flex flex-col gap-4 max-w-[640px]">

            <GhBox header="Physical Frame Settings">
              <form onSubmit={handleUpdateDimensions} className="flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-4">
                  {[
                    { label: 'Frame Width (mm)',  val: newWidth,  set: setNewWidth  },
                    { label: 'Frame Height (mm)', val: newHeight, set: setNewHeight },
                  ].map(({ label, val, set }) => (
                    <div key={label}>
                      <label className="block text-xs font-semibold text-[#8b949e] mb-1.5">{label}</label>
                      <input
                        type="number"
                        value={val}
                        onChange={e => set(e.target.value)}
                        className="w-full bg-[#010409] border border-[#30363d] rounded-md px-3 py-1.5 text-sm text-[#e6edf3] font-mono outline-none focus:border-[#58a6ff] focus:ring-2 focus:ring-[#1f6feb40]"
                      />
                    </div>
                  ))}
                </div>

                <div className="flex items-center gap-2 px-3 py-2.5 rounded-md text-xs bg-[#d2992226] border border-[#d2992240] text-[#d29922]">
                  ⚠ Changes require re-calibration after applying.
                </div>

                <button
                  type="submit"
                  className="w-full py-2 px-4 bg-[#238636] hover:bg-[#2ea043] border border-[#f0f6fc1a] rounded-md text-white text-sm font-medium cursor-pointer transition-colors"
                >
                  Apply Changes
                </button>
              </form>
            </GhBox>

            <GhBox header="Active Motor Hardware" noPadding>
              <table className="w-full border-collapse text-xs">
                <thead>
                  <tr>
                    {['ID', 'IP Address', 'Resolution', 'Pulley'].map(h => (
                      <th key={h} className="text-left px-4 py-1.5 text-[11px] font-semibold text-[#8b949e] border-b border-[#30363d]">
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {config?.motors.map((m, i) => (
                    <tr key={m.motor_id} className={i < (config.motors.length - 1) ? 'border-b border-[#21262d]' : ''}>
                      <td className="px-4 py-2 font-mono text-[#58a6ff]">#{m.motor_id}</td>
                      <td className="px-4 py-2 font-mono text-[#e6edf3]">{m.ip_address}</td>
                      <td className="px-4 py-2 font-mono text-[#8b949e]">{m.steps_per_rev} steps</td>
                      <td className="px-4 py-2 font-mono text-[#8b949e]">{m.pulley_mm} mm</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </GhBox>

          </div>
        )}

      </main>
    </div>
  );
}