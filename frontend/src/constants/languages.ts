import { Language } from './enums';

export interface LanguageMetadata {
    name: string;
    nativeName: string;
    code: Language;
}

export const LANGUAGE_CONFIG: Record<Language, LanguageMetadata> = {
    [Language.ENGLISH]: {
        name: 'English',
        nativeName: 'English',
        code: Language.ENGLISH,
    },
    [Language.RUSSIAN]: {
        name: 'Russian',
        nativeName: 'Русский',
        code: Language.RUSSIAN,
    },
};
