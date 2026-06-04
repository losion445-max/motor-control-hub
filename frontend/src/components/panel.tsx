
export const Panel = ({
  title,
  children,
  className = "",
  onClick
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
  onClick?: (e: React.MouseEvent<HTMLDivElement>) => void;
}) => (
  <div
    onClick={onClick}
    className={`bg-[#161b22] border border-[#30363d] flex flex-col shadow-[0_4px_20px_rgba(0,0,0,0.5)] ${className} ${onClick ? 'cursor-pointer' : ''}`}>
    <div className="px-3 py-1.5 bg-[#30363d]/20 border-b border-[#30363d] flex justify-between items-center">
      <span className="text-[10px] text-[#8b949e] font-bold tracking-[0.2em] font-mono">
        // {title.toUpperCase()}
      </span>
      <div className="flex gap-1">
        <div className="w-1 h-1 bg-[#30363d]" />
        <div className="w-1 h-1 bg-[#30363d]" />
      </div>
    </div>
    <div className="p-4 flex-1">{children}</div>
  </div>
);