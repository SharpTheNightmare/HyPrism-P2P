import React, { useState } from 'react';

// Try to import background image
let backgroundImage: string | null = null;
try {
  backgroundImage = new URL('../assets/background.jpg', import.meta.url).href;
} catch {
  // Image not available, will use gradient fallback
}

export const BackgroundImage: React.FC = () => {
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageError, setImageError] = useState(false);

  const showImage = backgroundImage && !imageError;

  return (
    <>
      {/* Background container */}
      <div className="absolute inset-0 overflow-hidden">
        {/* Static image background if available */}
        {showImage && (
          <img
            src={backgroundImage}
            alt=""
            onLoad={() => setImageLoaded(true)}
            onError={() => setImageError(true)}
            className={`absolute w-full h-full object-cover transition-opacity duration-500 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
          />
        )}
        
        {/* Gradient fallback (shown if no image or while loading) */}
        {(!showImage || !imageLoaded) && (
          <>
            {/* Base layer - Deep blue/teal gradient */}
            <div 
              className="absolute inset-0"
              style={{
                background: 'linear-gradient(135deg, #0a1628 0%, #0d2137 25%, #0f2847 50%, #0d2137 75%, #0a1628 100%)',
              }}
            />
            
            {/* Animated aurora layer 1 - Teal */}
            <div 
              className="absolute inset-0 opacity-40"
              style={{
                background: 'radial-gradient(ellipse 80% 50% at 20% 40%, rgba(45, 212, 191, 0.4) 0%, transparent 60%)',
                animation: 'aurora1 15s ease-in-out infinite',
              }}
            />
            
            {/* Animated aurora layer 2 - Cyan */}
            <div 
              className="absolute inset-0 opacity-30"
              style={{
                background: 'radial-gradient(ellipse 60% 40% at 70% 60%, rgba(34, 211, 238, 0.3) 0%, transparent 50%)',
                animation: 'aurora2 20s ease-in-out infinite',
              }}
            />
            
            {/* Animated aurora layer 3 - Orange accent */}
            <div 
              className="absolute inset-0 opacity-20"
              style={{
                background: 'radial-gradient(ellipse 40% 30% at 80% 20%, rgba(251, 146, 60, 0.3) 0%, transparent 50%)',
                animation: 'aurora3 12s ease-in-out infinite',
              }}
            />
          </>
        )}
        
        {/* Vignette effect - always show */}
        <div 
          className="absolute inset-0"
          style={{
            background: 'radial-gradient(ellipse at center, transparent 0%, transparent 40%, rgba(0,0,0,0.5) 100%)',
          }}
        />
      </div>

      {/* Light overlay for readability */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/20 via-transparent to-black/40" />
      
      {/* CSS for fallback animations */}
      <style>{`
        @keyframes aurora1 {
          0%, 100% { transform: translateX(0%) translateY(0%) scale(1); opacity: 0.4; }
          50% { transform: translateX(10%) translateY(5%) scale(1.1); opacity: 0.5; }
        }
        @keyframes aurora2 {
          0%, 100% { transform: translateX(0%) translateY(0%) scale(1); opacity: 0.3; }
          50% { transform: translateX(-10%) translateY(10%) scale(1.15); opacity: 0.4; }
        }
        @keyframes aurora3 {
          0%, 100% { transform: translateX(0%) translateY(0%) scale(1); opacity: 0.2; }
          50% { transform: translateX(-15%) translateY(10%) scale(1.2); opacity: 0.35; }
        }
      `}</style>
    </>
  );
};
