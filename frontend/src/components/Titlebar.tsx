import React, { useEffect, useState } from 'react';
import { Minus, X } from 'lucide-react';
import { Quit, WindowMinimise } from '../../wailsjs/runtime/runtime';
import { GetLauncherVersion } from '../../wailsjs/go/app/App';
import appIcon from '../assets/appicon.png';

export const Titlebar: React.FC = () => {
  const [version, setVersion] = useState<string>('');

  useEffect(() => {
    GetLauncherVersion().then(setVersion).catch(() => setVersion(''));
  }, []);

  return (
    <div 
      className="absolute top-0 left-0 right-0 h-10 flex items-center justify-between px-4 z-50"
      style={{ '--wails-draggable': 'drag' } as React.CSSProperties}
    >
      {/* Logo/Title */}
      <div className="flex items-center gap-3" style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}>
        <div className="w-6 h-6 rounded-md bg-gradient-to-br from-[#4A90E2] to-[#2E5C9A] flex items-center justify-center p-0.5">
          <img src={appIcon} alt="HyPrism" className="w-full h-full object-contain" />
        </div>
        <div className="flex items-center gap-2">
          <span className="text-white/80 text-sm font-semibold tracking-tight">HyPrism</span>
          {version && <span className="text-white/40 text-xs">v{version}</span>}
        </div>
      </div>

      {/* Window Controls */}
      <div className="flex items-center gap-1" style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}>
        <button
          onClick={() => WindowMinimise()}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-white/10 active:scale-95 transition-all duration-150"
        >
          <Minus size={14} className="text-white/60" />
        </button>
        
        <button
          onClick={() => Quit()}
          className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-red-500/80 active:scale-95 transition-all duration-150"
        >
          <X size={14} className="text-white/60 hover:text-white" />
        </button>
      </div>
    </div>
  );
};
