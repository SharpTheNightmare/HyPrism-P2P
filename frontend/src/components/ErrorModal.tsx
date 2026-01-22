import React from 'react';
import { motion } from 'framer-motion';
import { AlertTriangle, X, Copy, RefreshCw, Info, CheckCircle } from 'lucide-react';

interface ErrorModalProps {
  error: {
    type: string;
    message: string;
    technical?: string;
    timestamp?: string;
  };
  onClose: () => void;
}

export const ErrorModal: React.FC<ErrorModalProps> = ({ error, onClose }) => {
  const [copied, setCopied] = React.useState(false);

  const copyError = () => {
    const errorText = `Error Type: ${error.type}\nMessage: ${error.message}\nTechnical: ${error.technical || 'N/A'}\nTimestamp: ${error.timestamp || new Date().toISOString()}`;
    navigator.clipboard.writeText(errorText);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getErrorColor = (type: string) => {
    switch (type) {
      case 'INFO': return 'text-green-400';
      case 'NETWORK': return 'text-blue-400';
      case 'FILESYSTEM': return 'text-yellow-400';
      case 'VALIDATION': return 'text-orange-400';
      case 'GAME': return 'text-red-400';
      case 'UPDATE': return 'text-purple-400';
      default: return 'text-red-400';
    }
  };

  const isInfo = error.type === 'INFO';
  const isSuccess = error.type === 'SUCCESS';

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-8"
      onClick={onClose}
    >
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        className={`w-full max-w-lg bg-[#0d0d0d] rounded-2xl border overflow-hidden ${
          isInfo || isSuccess ? 'border-green-500/20' : 'border-red-500/20'
        }`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className={`flex items-center justify-between p-5 border-b border-white/10 ${
          isInfo || isSuccess ? 'bg-green-500/5' : 'bg-red-500/5'
        }`}>
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
              isInfo || isSuccess ? 'bg-green-500/20' : 'bg-red-500/20'
            }`}>
              {isInfo ? (
                <Info size={20} className="text-green-400" />
              ) : isSuccess ? (
                <CheckCircle size={20} className="text-green-400" />
              ) : (
                <AlertTriangle size={20} className="text-red-400" />
              )}
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">
                {isInfo ? 'Information' : isSuccess ? 'Success' : 'Error Occurred'}
              </h2>
              <span className={`text-xs font-medium ${getErrorColor(error.type)}`}>
                {error.type}
              </span>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-white/5 text-gray-400 hover:text-white transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-5 space-y-4">
          <div>
            <h3 className="text-white font-medium mb-1">{error.message}</h3>
            {error.technical && (
              <div className="mt-3 p-3 bg-black/50 rounded-lg border border-white/5">
                <p className="text-xs text-gray-400 font-mono break-all">
                  {error.technical}
                </p>
              </div>
            )}
          </div>

          {error.timestamp && (
            <p className="text-xs text-gray-500">
              Occurred at: {new Date(error.timestamp).toLocaleString()}
            </p>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-5 border-t border-white/10 bg-black/30">
          <button
            onClick={copyError}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-white/5 text-gray-300 hover:bg-white/10 transition-colors text-sm"
          >
            <Copy size={14} />
            {copied ? 'Copied!' : 'Copy Error'}
          </button>
          
          <div className="flex gap-3">
            <button
              onClick={() => window.location.reload()}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-white/5 text-gray-300 hover:bg-white/10 transition-colors text-sm"
            >
              <RefreshCw size={14} />
              Reload
            </button>
            <button
              onClick={onClose}
              className={`px-6 py-2 rounded-lg font-medium text-sm transition-colors ${
                isInfo || isSuccess 
                  ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30' 
                  : 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
              }`}
            >
              Dismiss
            </button>
          </div>
        </div>
      </motion.div>
    </motion.div>
  );
};
