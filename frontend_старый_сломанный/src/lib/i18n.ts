import { ref, computed } from "vue";
import { APP_CONFIG } from '@/config/app-config';

export type Lang = "ru" | "en";

// Reactive language state
export const lang = ref<Lang>((localStorage.getItem(APP_CONFIG.theme.STORAGE_KEY) as Lang) || "ru");

export function setLang(l: Lang) {
  lang.value = l;
  localStorage.setItem(APP_CONFIG.theme.STORAGE_KEY, l);
}

export function getLang(): Lang {
  return lang.value;
}

const dict: Record<Lang, Record<string, string>> = {
  ru: {
    "settings.title": "Настройки",
    "settings.language": "Язык",
    "settings.saving": "Сохранение...",
    "settings.saveClose": "Сохранить и закрыть",
    "projects.title": "Проекты",
    "projects.open": "Выбрать папку проекта",
    export: "Экспорт",
    "commits.title": "История коммитов",
    "commits.loading": "Загрузка истории...",
    "commits.empty": "Коммиты не найдены.",
    "button.cancel": "Отмена",
    "button.addFiles": "Добавить файлы",
    "console.toggle": "Показать/скрыть консоль",
    "context.build": "Собрать контекст",
    generate: "Сгенерировать решение",
    generating: "Генерация...",
    "settings.language.label": "Язык интерфейса",
    "settings.language.ru": "Русский",
    "settings.language.en": "English",
  },
  en: {
    "settings.title": "Settings",
    "settings.language": "Language",
    "settings.saving": "Saving...",
    "settings.saveClose": "Save & Close",
    "projects.title": "Projects",
    "projects.open": "Open project folder",
    export: "Export",
    "commits.title": "Commit History",
    "commits.loading": "Loading history...",
    "commits.empty": "No commits found.",
    "button.cancel": "Cancel",
    "button.addFiles": "Add files",
    "console.toggle": "Toggle Console",
    "context.build": "Build Context",
    generate: "Generate Solution",
    generating: "Generating...",
    "settings.language.label": "Interface language",
    "settings.language.ru": "Russian",
    "settings.language.en": "English",
  },
};

// Simple function — reactive via lang.value
export function t(key: string): string {
  const d = dict[lang.value] || dict.ru;
  return d[key] || key;
}

// Composable with computed translator
export function useI18n() {
  function tr(key: string) {
    return computed(() => (dict[lang.value] || dict.ru)[key] || key);
  }
  return { tr, lang, setLang };
}