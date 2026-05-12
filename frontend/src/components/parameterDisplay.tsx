export const ParameterDisplay = ({ label, value, unit, color = "#58a6ff" }: { label: string; value: string | number; unit?: string; color?: string }) => (
  <div className="flex flex-col gap-1 border-l-2 border-[#30363d] pl-3 py-1">
    <span className="text-[9px] text-[#6e7681] font-bold tracking-widest">{label.toUpperCase()}</span>
    <div className="flex items-baseline gap-1">
      <span className="text-xl font-bold leading-none" style={{ color }}>{value}</span>
      {unit && <span className="text-[10px] text-[#6e7681]">{unit.toLowerCase()}</span>}
    </div>
  </div>
);