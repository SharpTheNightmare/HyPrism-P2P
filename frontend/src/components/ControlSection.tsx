import React from 'react';
import { FolderOpen, Trash, Play, Package, Square } from 'lucide-react';

interface ControlSectionProps {
  onPlay: () => void;
  onExit?: () => void;
  isDownloading: boolean;
  isGameRunning: boolean;
  progress: number;
  downloaded: number;
  total: number;
  actions: {
    openFolder: () => void;
    showDelete: () => void;
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
      <span className="absolute -top-10 left-1/2 -translate-x-1/2 px-2 py-1 text-xs bg-black/90 text-white rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
        {tooltip}
      </span>
    )}
  </button>
);

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

export const ControlSection: React.FC<ControlSectionProps> = ({
  onPlay,
  onExit,
  isDownloading,
  isGameRunning,
  progress,
  downloaded,
  total,
  actions
}) => {
  return (
    <div className="flex gap-4">
      {/* Left side - Navigation buttons */}
      <div className="flex flex-col gap-3">
        <div className="flex gap-3">
          <NavBtn onClick={actions.showModManager} icon={<Package size={20} />} tooltip="Mod Manager" />
          <NavBtn onClick={actions.openFolder} icon={<FolderOpen size={20} />} tooltip="Open Folder" />
          <NavBtn onClick={actions.showDelete} icon={<Trash size={20} />} tooltip="Delete Game" />
        </div>
        
        {/* Play/Exit button */}
        <div className="flex gap-3">
          {isGameRunning ? (
            <button
              onClick={onExit}
              className="flex-1 h-24 rounded-2xl font-black text-4xl tracking-tight flex items-center justify-center gap-4 bg-gradient-to-r from-red-600 to-red-500 text-white hover:shadow-lg hover:shadow-red-500/25 hover:scale-[1.02] active:scale-[0.98] transition-all duration-150 cursor-pointer"
            >
              <Square size={32} fill="currentColor" />
              <span>EXIT</span>
            </button>
          ) : isDownloading ? (
            <div className="flex-1 h-24 rounded-2xl bg-[#151515] border border-white/10 flex flex-col items-center justify-center gap-2 px-6 relative overflow-hidden">
              {/* Progress bar background */}
              <div 
                className="absolute inset-0 bg-gradient-to-r from-[#FFA845]/30 to-[#FF6B35]/30 transition-all duration-300"
                style={{ width: `${Math.min(progress, 100)}%` }}
              />
              
              {/* Content */}
              <div className="relative z-10 flex flex-col items-center">
                <span className="text-2xl font-bold text-white">{Math.round(progress)}%</span>
                {total > 0 && (
                  <span className="text-sm text-gray-400">
                    {formatBytes(downloaded)} / {formatBytes(total)}
                  </span>
                )}
              </div>
            </div>
          ) : (
            <button
              onClick={onPlay}
              className="flex-1 h-24 rounded-2xl font-black text-4xl tracking-tight flex items-center justify-center gap-4 bg-gradient-to-r from-[#FFA845] to-[#FF6B35] text-white hover:shadow-lg hover:shadow-[#FFA845]/25 hover:scale-[1.02] active:scale-[0.98] transition-all duration-150 cursor-pointer"
            >
              <Play size={32} fill="currentColor" />
              <span>PLAY</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
