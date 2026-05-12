export const StatusBadge = ({ label, active, pulse = false }: { label: string; active: boolean; pulse?: boolean }) => (
  <div className={`flex items-center gap-2 px-2 py-0.5 border ${active ? 'border-[#3fb95044] bg-[#2ea04311]' : 'border-[#f8514944] bg-[#f8514911]'}`}>
    <div className={`w-1.5 h-1.5 rounded-full ${active ? 'bg-[#3fb950]' : 'bg-[#f85149]'} ${active && pulse ? 'animate-pulse shadow-[0_0_8px_#3fb950]' : ''}`} />
    <span className={`text-[10px] font-bold tracking-tighter ${active ? 'text-[#3fb950]' : 'text-[#f85149]'}`}>
      {label.toUpperCase()}
    </span>
  </div>
);