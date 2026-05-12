import React, { useState, useEffect } from 'react';
import { hubApi } from '../infrastructure/api';
import { Panel } from '../components/panel';
import { TerminalButton } from '../components/terminalButton';
import type { FullConfig, KinematicsConfig } from '../domain/types';

interface ConfigProps {
  config: FullConfig;
  onRefresh: () => void;
}

export const ConfigView: React.FC<ConfigProps> = ({ config, onRefresh }) => {
  const [formData, setFormData] = useState({
    ...config.global.kinematics,
    motor_mapping: config.global.motor_mapping
  });

  useEffect(() => {
    setFormData({
      ...config.global.kinematics,
      motor_mapping: config.global.motor_mapping
    });
  }, [config]);

  const handleSave = async () => {
    try {
      await hubApi.config.update(formData);
      onRefresh();
      alert("CONFIG_UPDATED_SUCCESSFULLY");
    } catch (err) {
      console.error("FAILED_TO_UPDATE_CONFIG", err);
    }
  };

  const updateMapping = (index: number, value: string) => {
    const newMapping = [...formData.motor_mapping] as [number, number, number, number];
    newMapping[index] = parseInt(value) || 0;
    setFormData({ ...formData, motor_mapping: newMapping });
  };

  return (
    <div className="max-w-3xl space-y-6 font-mono uppercase pb-10">
      
      <Panel title="KINEMATICS_PHYSICAL_MODEL">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-6">
          {(Object.keys(config.global.kinematics) as Array<keyof KinematicsConfig>).map((key) => (
            <div key={key} className="space-y-2">
              <label className="text-[10px] text-[#6e7681] tracking-tighter italic">
                {key.replace(/_/g, ' ')} (MM)
              </label>
              <input title={key.replace(/_/g, ' ')}
                type="number" 
                value={formData[key]} 
                onChange={(e) => setFormData({ ...formData, [key]: parseFloat(e.target.value) || 0 })}
                className="w-full bg-[#010409] border border-[#30363d] p-3 lg:p-2 text-[#58a6ff] font-bold outline-none focus:border-[#58a6ff] transition-colors"
              />
            </div>
          ))}
        </div>
      </Panel>

      <Panel title="HARDWARE_ID_MAPPING">
        <div className="text-[10px] text-[#8b949e] mb-6 leading-relaxed">
          ASSIGN MOTOR_ID TO PHYSICAL CORNERS: <br />
          <span className="text-[#58a6ff]">[0:TOP_L, 1:TOP_R, 2:BTM_L, 3:BTM_R]</span>
        </div>

        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {formData.motor_mapping.map((id, index) => (
            <div key={index} className="bg-[#0d1117] border border-[#30363d] p-4 flex flex-col gap-2">
              <div className="text-[8px] text-[#6e7681] uppercase tracking-widest">NODE_{index}</div>
              <input title={String(index)}
                type="number"
                value={String(id)}
                onChange={(e) => updateMapping(index, e.target.value)}
                className="bg-transparent border-b border-[#30363d] text-center text-xl font-black text-[#e6edf3] outline-none focus:border-[#58a6ff] transition-colors"
              />
            </div>
          ))}
        </div>

        <div className="mt-10 border-t border-[#30363d] pt-6 flex justify-end">
          <div className="w-full lg:w-max">
            <TerminalButton 
              label="COMMIT_AND_SYNC_ORCHESTRATOR" 
              variant="success" 
              onClick={handleSave} 
            />
          </div>
        </div>
      </Panel>
      
    </div>
  );
};