import React, { useState } from 'react';
import { hubApi } from '../infrastructure/api';
import { Panel } from '../components/panel';
import { TerminalButton } from '../components/terminalButton';
import type { FullConfig, KinematicsConfig } from '../domain/types';

interface ConfigProps {
  config: FullConfig;
  onRefresh: () => void;
}

export const ConfigView: React.FC<ConfigProps> = ({ config, onRefresh }) => {
  const [formData, setFormData] = useState<KinematicsConfig>(config.global.kinematics);

  
  const handleSave = async () => {
    try {
      await hubApi.config.update(formData);
      onRefresh();
    } catch (err) {
      console.error("FAILED_TO_UPDATE_CONFIG", err);
    }
  };

  return (
    <div className="max-w-3xl space-y-6 font-mono uppercase pb-10">
      
      <Panel title="KINEMATICS_PHYSICAL_MODEL">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-6">
          {(Object.keys(formData) as Array<keyof KinematicsConfig>).map((key) => (
            <div key={key} className="space-y-2">
              <label className="text-[10px] text-[#6e7681] tracking-tighter italic">
                {key.replace(/_/g, ' ')}
              </label>
              <input
                title={key.replace(/_/g, ' ')}
                type="number" 
                inputMode="decimal"
                value={formData[key]} 
                onChange={(e) => setFormData({ ...formData, [key]: parseFloat(e.target.value) || 0 })}
                className="w-full bg-[#010409] border border-[#30363d] p-3 lg:p-2 text-[#58a6ff] font-bold outline-none focus:border-[#58a6ff] transition-colors text-base lg:text-sm"
              />
            </div>
          ))}
        </div>
        
        <div className="mt-8 border-t border-[#30363d] pt-6">
          <div className="w-full lg:w-max">
            <TerminalButton 
              label="COMMIT_CHANGES" 
              variant="success" 
              onClick={handleSave} 
            />
          </div>
        </div>
      </Panel>

      <Panel title="HARDWARE_ID_MAPPING">
        <div className="text-[10px] text-[#8b949e] mb-4 leading-relaxed">
          CORNER_SEQUENCE: <br className="sm:hidden" />
          <span className="text-[#58a6ff]">[TOP_L, TOP_R, BTM_L, BTM_R]</span>
        </div>

        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {config.global.motor_mapping.map((id, index) => (
            <div key={index} className="bg-[#0d1117] border border-[#30363d] p-4 lg:p-6 text-center shadow-inner">
              <div className="text-[8px] text-[#6e7681] mb-2 uppercase tracking-widest">NODE_{index}</div>
              <div className="text-lg lg:text-xl font-black text-[#e6edf3]">ID_{id}</div>
            </div>
          ))}
        </div>
      </Panel>
      
    </div>
  );
};