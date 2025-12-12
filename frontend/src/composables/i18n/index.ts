// i18n modules index - merges all translation modules
import { chatTranslations } from './chat'
import { commonTranslations } from './common'
import { contextTranslations } from './context'
import { errorsTranslations } from './errors'
import { filesTranslations } from './files'
import { gitTranslations } from './git'
import { onboardingTranslations } from './onboarding'
import { settingsTranslations } from './settings'
import { welcomeTranslations } from './welcome'

export type Locale = 'ru' | 'en'

export interface TranslationModule {
    ru: Record<string, string>
    en: Record<string, string>
}

// Merge all translations
function mergeTranslations(): Record<Locale, Record<string, string>> {
    const modules = [
        welcomeTranslations,
        filesTranslations,
        contextTranslations,
        gitTranslations,
        chatTranslations,
        settingsTranslations,
        errorsTranslations,
        commonTranslations,
        onboardingTranslations,
    ]

    const merged: Record<Locale, Record<string, string>> = {
        ru: {},
        en: {},
    }

    for (const module of modules) {
        Object.assign(merged.ru, module.ru)
        Object.assign(merged.en, module.en)
    }

    return merged
}

export const translations = mergeTranslations()

// Re-export individual modules for direct access if needed
export {
    chatTranslations, commonTranslations, contextTranslations, errorsTranslations, filesTranslations, gitTranslations, onboardingTranslations, settingsTranslations, welcomeTranslations
}

