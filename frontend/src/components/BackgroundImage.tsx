import React, { useState } from 'react';
import backgroundImage from '../assets/background.jpg';

export const BackgroundImage: React.FC = () => {
  const [imageLoaded, setImageLoaded] = useState(false);

  return (
    <>
      {/* Background container */}
      <div className="absolute inset-0 overflow-hidden bg-black">
        <img
          src={backgroundImage}
          alt=""
          onLoad={() => setImageLoaded(true)}
          className={`absolute w-full h-full object-cover transition-opacity duration-500 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
        />
        
        {/* Vignette effect */}
        <div 
          className="absolute inset-0"
          style={{
            background: 'radial-gradient(ellipse at center, transparent 0%, transparent 40%, rgba(0,0,0,0.5) 100%)',
          }}
        />
      </div>

      {/* Light overlay for readability */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/20 via-transparent to-black/40" />
    </>
  );
};
