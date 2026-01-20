import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { Language } from './constants/enums';
import ru from './locales/ru.json';
import en from './locales/en.json';
import fr from './locales/fr.json';

const getSavedLanguage = (): string => {
    const saved = localStorage.getItem('i18nextLng');
    const supportedLanguages = Object.values(Language) as string[];

    if (saved && supportedLanguages.includes(saved)) {
        return saved;
    }
    return Language.ENGLISH;
};

i18n
    .use(initReactI18next)
    .init({
        resources: {
            [Language.ENGLISH]: {
                translation: en,
            },
            [Language.RUSSIAN]: {
                translation: ru,
            },
            [Language.FRENCH]: {
                translation: fr,
            },
        },
        lng: getSavedLanguage(),
        fallbackLng: Language.ENGLISH,
        interpolation: {
            escapeValue: false,
        },
    });

i18n.on('languageChanged', (lng) => {
    localStorage.setItem('i18nextLng', lng);
});

export default i18n;
