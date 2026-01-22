import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, ExternalLink, Loader2, CheckCircle } from 'lucide-react';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  authMessage: string;
  isAuthenticating: boolean;
  onOpenURL: (url: string) => void;
}

export const AuthModal: React.FC<AuthModalProps> = ({
  isOpen,
  onClose,
  authMessage,
  isAuthenticating,
  onOpenURL
}) => {
  const [deviceCode, setDeviceCode] = useState<string>('');
  const [verificationURL, setVerificationURL] = useState<string>('');
  const [completeURL, setCompleteURL] = useState<string>('');

  useEffect(() => {
    if (!authMessage) return;

    // Parse the auth message to extract URLs and code
    const lines = authMessage.split('\n');
    for (const line of lines) {
      if (line.startsWith('Visit:')) {
        setVerificationURL(line.replace('Visit:', '').trim());
      } else if (line.startsWith('Enter code:')) {
        setDeviceCode(line.replace('Enter code:', '').trim());
      } else if (line.startsWith('Or visit:')) {
        setCompleteURL(line.replace('Or visit:', '').trim());
      }
    }
  }, [authMessage]);

  const handleOpenBrowser = () => {
    const urlToOpen = completeURL || verificationURL;
    if (urlToOpen) {
      onOpenURL(urlToOpen);
    }
  };

  const handleCopyCode = () => {
    if (deviceCode) {
      navigator.clipboard.writeText(deviceCode);
    }
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
        onClick={onClose}
      >
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          onClick={(e) => e.stopPropagation()}
          className="bg-gradient-to-br from-[#1a1a1a] to-[#0d0d0d] border border-[#FFA845]/30 rounded-2xl p-8 max-w-md w-full mx-4 shadow-2xl"
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-white">Login to Hytale</h2>
            <motion.button
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
              onClick={onClose}
              className="p-2 rounded-lg text-white/60 hover:text-white hover:bg-white/10 transition-colors"
            >
              <X size={20} />
            </motion.button>
          </div>

          {/* Content */}
          <div className="space-y-6">
            {isAuthenticating && deviceCode ? (
              <>
                {/* Instructions */}
                <p className="text-white/70 text-sm">
                  To authenticate HyPrism with your Hytale account:
                </p>

                {/* Device Code */}
                <div className="bg-[#0d0d0d] border border-[#FFA845]/20 rounded-xl p-4">
                  <div className="text-xs text-white/50 mb-2">Device Code</div>
                  <div className="flex items-center justify-between">
                    <code className="text-2xl font-mono font-bold text-[#FFA845] tracking-wider">
                      {deviceCode}
                    </code>
                    <motion.button
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      onClick={handleCopyCode}
                      className="px-3 py-1.5 text-xs bg-[#FFA845]/20 text-[#FFA845] rounded-lg hover:bg-[#FFA845]/30 transition-colors"
                    >
                      Copy
                    </motion.button>
                  </div>
                </div>

                {/* Open Browser Button */}
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={handleOpenBrowser}
                  className="w-full flex items-center justify-center gap-2 bg-gradient-to-r from-[#FFA845] to-[#FF8C45] text-white font-semibold py-3 px-6 rounded-xl hover:shadow-lg hover:shadow-[#FFA845]/20 transition-all"
                >
                  <ExternalLink size={18} />
                  Open Authentication Page
                </motion.button>

                {/* URL Display */}
                <div className="text-xs text-white/40 text-center break-all">
                  {verificationURL}
                </div>

                {/* Waiting Status */}
                <div className="flex items-center justify-center gap-3 text-white/60 text-sm py-4">
                  <Loader2 size={16} className="animate-spin" />
                  <span>Waiting for authorization...</span>
                </div>

                <p className="text-xs text-white/40 text-center">
                  The code will expire in 15 minutes
                </p>
              </>
            ) : authMessage.includes('Logged in as') ? (
              <>
                {/* Success State */}
                <div className="flex flex-col items-center justify-center py-6 space-y-4">
                  <motion.div
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ type: 'spring', stiffness: 200 }}
                  >
                    <CheckCircle size={64} className="text-green-500" />
                  </motion.div>
                  <div className="text-center">
                    <h3 className="text-xl font-bold text-white mb-2">
                      Successfully Logged In!
                    </h3>
                    <p className="text-white/60">
                      {authMessage.replace('Logged in as: ', '')}
                    </p>
                  </div>
                </div>
              </>
            ) : (
              <>
                {/* Initial Loading State */}
                <div className="flex flex-col items-center justify-center py-8 space-y-4">
                  <Loader2 size={48} className="animate-spin text-[#FFA845]" />
                  <p className="text-white/60">Initializing authentication...</p>
                </div>
              </>
            )}
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
