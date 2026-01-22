import React from 'react';
import { motion } from 'framer-motion';
import { Download, LogIn, LogOut, User } from 'lucide-react';

interface ProfileSectionProps {
  username: string;
  updateAvailable: boolean;
  onUpdate: () => void;
  launcherVersion: string;
  isLoggedIn: boolean;
  onLogin: () => void;
  onLogout: () => void;
  uuid?: string;
}

export const ProfileSection: React.FC<ProfileSectionProps> = ({
  username,
  updateAvailable,
  onUpdate,
  launcherVersion,
  isLoggedIn,
  onLogin,
  onLogout,
  uuid
}) => {

  return (
    <motion.div 
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex flex-col gap-2"
    >
      {/* Username and Auth Status */}
      {isLoggedIn && (
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <User size={20} className="text-green-500" />
            <span className="text-2xl font-bold text-white">{username}</span>
          </div>
          
          {/* Auth Status Badge */}
          {uuid && (
            <div className="flex items-center gap-1 text-xs text-green-500/80">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span>Authenticated</span>
            </div>
          )}
        </div>
      )}

      {/* Auth Buttons */}
      <div className="flex gap-2 mt-2">
        {isLoggedIn ? (
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onLogout}
            className="flex items-center gap-2 text-xs text-red-400 hover:text-red-300 transition-colors px-3 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20"
          >
            <LogOut size={12} />
            Logout
          </motion.button>
        ) : (
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onLogin}
            className="flex items-center gap-2 text-xs text-[#FFA845] hover:text-[#FFB85F] transition-colors px-3 py-1.5 rounded-lg bg-[#FFA845]/10 hover:bg-[#FFA845]/20"
          >
            <LogIn size={12} />
            Login with Hytale
          </motion.button>
        )}
        
        {/* Update button */}
        {updateAvailable && (
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onUpdate}
            className="flex items-center gap-2 text-xs text-[#FFA845] hover:text-[#FFB85F] transition-colors px-3 py-1.5 rounded-lg bg-[#FFA845]/10 hover:bg-[#FFA845]/20"
          >
            <Download size={12} />
            Update
          </motion.button>
        )}
      </div>
      
      {/* Launcher version */}
      <div className="text-xs text-white/30 mt-1">
        HyPrism {launcherVersion}
      </div>
    </motion.div>
  );
};
