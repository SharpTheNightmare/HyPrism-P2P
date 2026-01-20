import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Check } from 'lucide-react';
import { Language } from '../constants/enums';
import { LANGUAGE_CONFIG } from '../constants/languages';

export const LanguageSelector: React.FC = () => {
    const { i18n } = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    const handleLanguageSelect = (langCode: Language) => {
        i18n.changeLanguage(langCode);
        setIsOpen(false);
    };

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        document.addEventListener('keydown', handleEscape);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
            document.removeEventListener('keydown', handleEscape);
        };
    }, []);

    return (
        <div ref={dropdownRef} className="relative">
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="w-12 h-12 rounded-xl glass border border-white/5 flex items-center justify-center text-white/60 hover:text-[#FFA845] hover:bg-[#FFA845]/10 active:scale-95 transition-all duration-150"
                title={i18n.t('Change Language')}
            >
                <span className="font-bold text-sm">{i18n.language.toUpperCase()}</span>
            </button>

            {/* Language Dropdown */}
            {isOpen && (
                <div className="absolute bottom-full left-0 mb-2 z-[100] min-w-[150px] bg-[#1a1a1a] backdrop-blur-xl border border-white/10 rounded-xl shadow-xl shadow-black/50 overflow-hidden">
                    {Object.values(LANGUAGE_CONFIG).map((lang) => (
                        <button
                            key={lang.code}
                            onClick={() => handleLanguageSelect(lang.code)}
                            className={`w-full px-3 py-2 flex items-center gap-2 text-sm ${i18n.language === lang.code
                                ? 'bg-white/20 text-white'
                                : 'text-white/70 hover:bg-white/10 hover:text-white'
                                }`}
                        >
                            {i18n.language === lang.code && <Check size={14} className="text-white" strokeWidth={3} />}
                            <div className={`flex flex-col items-start ${i18n.language === lang.code ? '' : 'ml-[22px]'}`}>
                                <span className="font-medium">{lang.nativeName}</span>
                                <span className="text-xs opacity-50">{lang.name}</span>
                            </div>
                        </button>
                    ))}
                </div>
            )}
        </div>
    );
};
