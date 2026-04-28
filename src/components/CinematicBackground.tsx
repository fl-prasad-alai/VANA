import React from 'react';
import { motion } from 'framer-motion';
import { useTheme } from '../contexts/ThemeContext';

const CinematicBackground: React.FC = React.memo(() => {
  const { theme } = useTheme();
  const isGreening = theme === 'greening';
  const isLight = theme === 'light';

  return (
    <div className={`fixed inset-0 -z-10 overflow-hidden transition-colors duration-1000 ${
      isGreening ? 'bg-[#022c22]' : isLight ? 'bg-slate-50' : 'bg-[#0a0a0c]'
    }`}>
      {/* Liquid Mesh Blobs */}
      <div className={`absolute inset-0 transition-opacity duration-1000 ${isLight ? 'opacity-40' : 'opacity-100'}`}>
        <motion.div
          animate={{
            x: [0, 40, -40, 0],
            y: [0, -60, 60, 0],
            scale: [1, 1.2, 0.9, 1],
            rotate: [0, 90, 180, 270, 360],
          }}
          transition={{ duration: 25, repeat: Infinity, ease: 'linear' }}
          className={`
            absolute -top-[20%] -left-[10%] h-[80%] w-[80%] rounded-full blur-[120px] will-change-transform transition-colors duration-1000
            ${isGreening ? 'bg-[#10b981]/20' : isLight ? 'bg-violet-400/20' : 'bg-[#8b5cf6]/20'}
          `}
        />
        <motion.div
          animate={{
            x: [0, -50, 50, 0],
            y: [0, 70, -70, 0],
            scale: [1, 0.8, 1.1, 1],
            rotate: [360, 270, 180, 90, 0],
          }}
          transition={{ duration: 30, repeat: Infinity, ease: 'linear' }}
          className={`
            absolute -bottom-[20%] -right-[10%] h-[70%] w-[70%] rounded-full blur-[100px] will-change-transform transition-colors duration-1000
            ${isGreening ? 'bg-[#064e3b]/40' : isLight ? 'bg-slate-200/40' : 'bg-[#1e293b]/40'}
          `}
        />
        <motion.div
          animate={{
            x: [0, 30, -30, 0],
            y: [0, 30, -30, 0],
            scale: [0.8, 1, 0.8],
          }}
          transition={{ duration: 20, repeat: Infinity, ease: 'easeInOut' }}
          className={`
            absolute top-1/4 left-1/4 h-[50%] w-[50%] rounded-full blur-[140px] will-change-transform transition-colors duration-1000
            ${isGreening ? 'bg-[#022c22]/60' : isLight ? 'bg-slate-300/30' : 'bg-[#0f172a]/60'}
          `}
        />
      </div>

      {/* Noise Texture Overlay */}
      <div className="absolute inset-0 opacity-[0.03] pointer-events-none mix-blend-overlay">
        <svg viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg" className="h-full w-full">
          <filter id="noiseFilter">
            <feTurbulence type="fractalNoise" baseFrequency="0.65" numOctaves="3" stitchTiles="stitch" />
          </filter>
          <rect width="100%" height="100%" filter="url(#noiseFilter)" />
        </svg>
      </div>

      {/* Vignettes */}
      <div className={`absolute inset-0 bg-gradient-to-t via-transparent pointer-events-none transition-colors duration-1000 ${
        isGreening ? 'from-[#022c22] to-[#022c22]/80'
        : isLight ? 'from-white/30 to-white/10'
        : 'from-[#0a0a0c] to-[#0a0a0c]/80'
      }`} />
      <div className={`absolute inset-0 bg-gradient-to-r via-transparent pointer-events-none transition-colors duration-1000 ${
        isGreening ? 'from-[#022c22] to-[#022c22]/80'
        : isLight ? 'from-white/20 to-white/10'
        : 'from-[#0a0a0c] to-[#0a0a0c]/80'
      }`} />
    </div>
  );
});

CinematicBackground.displayName = 'CinematicBackground';

export default CinematicBackground;
