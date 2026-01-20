import React from 'react';
import { motion } from 'framer-motion';
import { DownloadCloud } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface UpdateOverlayProps {
  progress: number;
  downloaded: number;
  total: number;
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

export const UpdateOverlay: React.FC<UpdateOverlayProps> = ({ progress, downloaded, total }) => {
  const { t } = useTranslation();
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="absolute inset-0 z-[100] bg-[#090909]/95 backdrop-blur-2xl flex flex-col items-center justify-center p-20 text-center"
    >
      <motion.div
        animate={{
          y: [0, -10, 0],
          scale: [1, 1.05, 1]
        }}
        transition={{
          duration: 2,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      >
        <DownloadCloud size={80} className="text-[#FFA845] mb-8" />
      </motion.div>

      <h1 className="text-5xl font-black mb-4 tracking-tight text-white">
        {t('UPDATING LAUNCHER')}
      </h1>

      <p className="text-gray-400 mb-12 max-w-md text-lg font-medium">
        {t('Please wait while we download and install the latest version of HyPrism.')}
      </p>

      {/* Progress bar */}
      <div className="w-full max-w-md">
        <div className="relative h-3 bg-white/5 rounded-full overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: `${Math.min(progress, 100)}%` }}
            transition={{ duration: 0.3 }}
            className="absolute inset-y-0 left-0 bg-gradient-to-r from-[#FFA845] to-[#FF6B35] rounded-full"
          />
          <div className="absolute inset-0 animate-shimmer" />
        </div>

        <div className="flex justify-between items-center mt-4 text-sm">
          <span className="text-gray-400">
            {formatBytes(downloaded)} / {formatBytes(total)}
          </span>
          <span className="font-bold text-[#FFA845]">{Math.round(progress)}%</span>
        </div>
      </div>

      <p className="text-xs text-gray-500 mt-8">
        {t('The launcher will restart automatically when the update is complete.')}
      </p>
    </motion.div>
  );
};
