import React from 'react';
import { motion } from 'framer-motion';
import { Edit3, Check, Download } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface ProfileSectionProps {
  username: string;
  isEditing: boolean;
  onEditToggle: (editing: boolean) => void;
  onUserChange: (name: string) => void;
  updateAvailable: boolean;
  onUpdate: () => void;
  launcherVersion: string;
}

export const ProfileSection: React.FC<ProfileSectionProps> = ({
  username,
  isEditing,
  onEditToggle,
  onUserChange,
  updateAvailable,
  onUpdate,
  launcherVersion
}) => {
  const { t } = useTranslation();
  const [editValue, setEditValue] = React.useState(username);

  React.useEffect(() => {
    setEditValue(username);
  }, [username]);

  const handleSave = () => {
    if (editValue.trim() && editValue.length <= 16) {
      onUserChange(editValue.trim());
      onEditToggle(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSave();
    } else if (e.key === 'Escape') {
      setEditValue(username);
      onEditToggle(false);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      className="flex flex-col gap-2"
    >
      {/* Username */}
      <div className="flex items-center gap-2">
        {isEditing ? (
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onKeyDown={handleKeyDown}
              maxLength={16}
              autoFocus
              className="bg-[#151515] text-white text-xl font-bold px-3 py-1 rounded-lg border border-[#FFA845]/30 focus:border-[#FFA845] outline-none w-40"
            />
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={handleSave}
              className="p-2 rounded-lg bg-[#FFA845]/20 text-[#FFA845] hover:bg-[#FFA845]/30"
            >
              <Check size={16} />
            </motion.button>
          </div>
        ) : (
          <>
            <span className="text-2xl font-bold text-white">{username}</span>
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={() => onEditToggle(true)}
              className="p-1.5 rounded-lg text-white/40 hover:text-white/80 hover:bg-white/5"
              title={t('Edit Profile')}
            >
              <Edit3 size={14} />
            </motion.button>
          </>
        )}
      </div>

      {/* Update button only if available */}
      {updateAvailable && (
        <motion.button
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={onUpdate}
          className="flex items-center gap-1 text-xs text-[#FFA845] hover:text-[#FFB85F] transition-colors mt-1"
        >
          <Download size={12} />
          {t('Update Available')}
        </motion.button>
      )}

      {/* Launcher version */}
      <div className="text-xs text-white/30 mt-1">
        HyPrism {launcherVersion}
      </div>
    </motion.div>
  );
};
