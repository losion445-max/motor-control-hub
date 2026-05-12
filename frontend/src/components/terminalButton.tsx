type ButtonVariant = 'default' | 'danger' | 'success' | 'warning';

interface TerminalButtonProps {
  label: string;
  onClick: () => void;
  variant?: ButtonVariant;
  disabled?: boolean;
  className?: string;
}

export const TerminalButton = ({ 
  label, 
  onClick, 
  variant = 'default', 
  disabled = false,
  className = "" 
}: TerminalButtonProps) => {
  
  const themes = {
    default: "border-[#30363d] text-[#e6edf3] hover:bg-[#e6edf3] hover:text-[#0d1117]",
    danger:  "border-[#f8514966] text-[#f85149] hover:bg-[#f85149] hover:text-white",
    success: "border-[#3fb95066] text-[#3fb950] hover:bg-[#3fb950] hover:text-white",
    warning: "border-[#d2992266] text-[#d29922] hover:bg-[#d29922] hover:text-white",
  };

  return (
    <button 
      onClick={onClick} 
      disabled={disabled}
      className={`
        px-4 py-3 sm:py-2 text-[11px] font-bold tracking-widest border transition-all duration-150 
        active:scale-[0.97] uppercase disabled:opacity-30 disabled:cursor-not-allowed
        flex items-center justify-center text-center
        ${themes[variant]}
        ${className} 
      `}
    >
      {label}
    </button>
  );
};