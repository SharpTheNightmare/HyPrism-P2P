import React from 'react';
import { Play, Package, Square, Github, Bug, Coffee, Gamepad2 } from 'lucide-react';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface ControlSectionProps {
  onPlay: () => void;
  onExit?: () => void;
  isGameRunning: boolean;
  isLoggedIn: boolean;
  onGameLocationChange?: () => void;
  actions: {
    showModManager: () => void;
  };
}

const NavBtn: React.FC<{ onClick?: () => void; icon: React.ReactNode; tooltip?: string }> = ({ onClick, icon, tooltip }) => (
  <button
    onClick={onClick}
    className="w-12 h-12 rounded-xl glass border border-white/5 flex items-center justify-center text-white/60 hover:text-[#FFA845] hover:bg-[#FFA845]/10 active:scale-95 transition-all duration-150 relative group"
    title={tooltip}
  >
    {icon}
    {tooltip && (
      <span className="absolute -top-10 left-1/2 -translate-x-1/2 px-2 py-1 text-xs bg-black/90 text-white rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-50">
        {tooltip}
      </span>
    )}
  </button>
);


export const ControlSection: React.FC<ControlSectionProps> = ({
  onPlay,
  onExit,
  isGameRunning,
  isLoggedIn,
  onGameLocationChange,
  actions
}) => {

  const openGitHub = () => BrowserOpenURL('https://github.com/yyyumeniku/HyPrism');
  const openBugReport = () => BrowserOpenURL('https://github.com/yyyumeniku/HyPrism/issues/new');
  const openCoffee = () => BrowserOpenURL('https://buymeacoffee.com/yyyumeniku');

  return (
    <div className="flex flex-col gap-3">
      {/* Nav buttons */}
      <div className="flex gap-3 items-center">
        <NavBtn onClick={actions.showModManager} icon={<Package size={20} />} tooltip="Mod Manager" />
        <NavBtn 
          onClick={() => {
            if (onGameLocationChange) {
              onGameLocationChange();
            }
          }} 
          icon={<Gamepad2 size={20} />} 
          tooltip="Set Game Location" 
        />
        <NavBtn onClick={openGitHub} icon={<Github size={20} />} tooltip="GitHub" />
        <NavBtn onClick={openBugReport} icon={<Bug size={20} />} tooltip="Report Bug" />
        <button
          onClick={openCoffee}
          className="h-12 px-4 rounded-xl glass border border-white/5 flex items-center justify-center gap-2 text-white/60 hover:text-[#FFA845] hover:bg-[#FFA845]/10 active:scale-95 transition-all duration-150"
          title="Buy Me a Coffee"
        >
          <span className="text-sm font-medium">Buy me a</span>
          <Coffee size={20} />
        </button>
        
        {/* Spacer */}
        <div className="flex-1"></div>

        {/* Play/Exit button on right - Fixed width container */}
        <div className="w-[200px] flex justify-end">
          {isGameRunning ? (
            <button
              onClick={onExit}
              className="h-12 px-8 rounded-xl font-black text-xl tracking-tight flex items-center justify-center gap-2 bg-gradient-to-r from-red-600 to-red-500 text-white hover:shadow-lg hover:shadow-red-500/25 hover:scale-[1.02] active:scale-[0.98] transition-all duration-150 cursor-pointer"
            >
              <Square size={20} fill="currentColor" />
              <span>EXIT</span>
            </button>
          ) : isLoggedIn ? (
            <button
              onClick={onPlay}
              className="h-12 px-8 rounded-xl font-black text-xl tracking-tight flex items-center justify-center gap-2 bg-gradient-to-r from-[#FFA845] to-[#FF6B35] text-white hover:shadow-lg hover:shadow-[#FFA845]/25 hover:scale-[1.02] active:scale-[0.98] transition-all duration-150 cursor-pointer"
            >
              <Play size={20} fill="currentColor" />
              <span>PLAY</span>
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
};
