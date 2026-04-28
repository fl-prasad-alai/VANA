import React from 'react';
import { motion } from 'framer-motion';
import { useTheme } from '../contexts/ThemeContext';

interface VitalityCoreProps {
  status?: 'normal' | 'critical';
}

export default function VitalityCore({ status = 'normal' }: VitalityCoreProps) {
  const { theme } = useTheme();
  const isGreening = theme === 'greening';
  const isCritical = status === 'critical';

  const coreColor = isCritical
    ? '#e11d48'
    : isGreening
      ? '#10b981'
      : '#8b5cf6';

  return (
    <div className="relative w-48 h-48 flex items-center justify-center">
      <motion.div
        animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
        transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
        className="absolute inset-0 rounded-full blur-[60px]"
        style={{ backgroundColor: coreColor }}
      />

      <div className="relative z-10">
        {isGreening ? (
          <SproutingSeedling color={coreColor} />
        ) : (
          <HolographicSphere color={coreColor} />
        )}
      </div>

      <div className="absolute inset-0 rounded-full border border-white/10 dark:border-white/5 pointer-events-none" />
    </div>
  );
}

function HolographicSphere({ color }: { color: string }) {
  return (
    <motion.div
      animate={{ rotateY: 360, rotateX: [0, 10, -10, 0] }}
      transition={{ duration: 12, repeat: Infinity, ease: 'linear' }}
      className="w-24 h-24 rounded-full relative overflow-hidden"
      style={{
        background: `radial-gradient(circle at 30% 30%, white, ${color}33, transparent 70%)`,
        boxShadow: `0 0 40px ${color}44, inset 0 0 20px rgba(255,255,255,0.2)`,
      }}
    />
  );
}

function SproutingSeedling({ color }: { color: string }) {
  return (
    <motion.div className="flex flex-col items-center">
      <div className="relative">
        <motion.div
          initial={{ height: 0 }}
          animate={{ height: 40 }}
          transition={{ duration: 1.2, ease: 'easeOut' }}
          className="w-1.5 rounded-full"
          style={{ backgroundColor: color }}
        />
        <motion.div
          initial={{ scale: 0, rotate: 0 }}
          animate={{ scale: 1, rotate: -45 }}
          transition={{ delay: 0.5, type: 'spring', stiffness: 200 }}
          className="absolute top-2 -left-6 w-8 h-4 rounded-[100%_0_100%_0]"
          style={{ backgroundColor: color }}
        />
        <motion.div
          initial={{ scale: 0, rotate: 0 }}
          animate={{ scale: 1, rotate: 45 }}
          transition={{ delay: 0.7, type: 'spring', stiffness: 200 }}
          className="absolute top-4 -right-6 w-10 h-5 rounded-[0_100%_0_100%]"
          style={{ backgroundColor: color }}
        />
      </div>
      <motion.div
        animate={{ scale: [1, 1.1, 1], opacity: [0.5, 0.8, 0.5] }}
        transition={{ duration: 2, repeat: Infinity }}
        className="w-12 h-2 rounded-full mt-2 filter blur-sm"
        style={{ backgroundColor: color }}
      />
    </motion.div>
  );
}
