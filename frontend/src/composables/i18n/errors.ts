// Error and toast translations
export const errorsTranslations = {
    ru: {
        // Toasts
        'toast.selectFiles': 'Выберите файлы для построения контекста',
        'toast.contextBuilt': 'Контекст успешно построен',
        'toast.contextError': 'Ошибка при построении контекста',
        'toast.contextCopied': 'Контекст скопирован в буфер обмена',
        'toast.copyError': 'Ошибка при копировании контекста',
        'toast.buildFirst': 'Сначала постройте контекст',
        'toast.contextCleared': 'Контекст очищен',
        'toast.contextEmpty': 'Контекст уже пуст',
        'toast.solutionSoon': 'Генерация решения - скоро будет доступна',
        'toast.noFiles': 'Нет выбранных файлов',
        'toast.refreshed': 'Дерево файлов обновлено',
        'toast.refreshError': 'Ошибка при обновлении дерева файлов',

        // Errors
        'error.generic': 'Произошла ошибка',
        'error.loadFailed': 'Не удалось загрузить данные',
        'error.saveFailed': 'Не удалось сохранить',
        'error.networkError': 'Ошибка сети',
        'error.notFound': 'Не найдено',
        'error.invalidData': 'Некорректные данные',
        'error.suggestions': 'Рекомендации',
        'error.checkFiles': 'Проверьте, что файлы находятся в директории проекта',
        'error.checkPaths': 'Убедитесь, что пути не содержат ".." или абсолютных путей',
        'error.tryRefresh': 'Попробуйте обновить дерево файлов',
        'error.emptyContext': 'Контекст построен, но пуст',
        'error.tokenLimitExceeded': 'Контекст превышает лимит токенов: {actual}K (лимит: {limit}K). Уменьшите выбор файлов или увеличьте лимит в настройках.',
        'error.tokenLimitGeneric': 'Контекст превышает лимит токенов. Уменьшите выбор файлов или увеличьте лимит в настройках.',
    },
    en: {
        'toast.selectFiles': 'Select files to build context',
        'toast.contextBuilt': 'Context built successfully',
        'toast.contextError': 'Error building context',
        'toast.contextCopied': 'Context copied to clipboard',
        'toast.copyError': 'Error copying context',
        'toast.buildFirst': 'Build context first',
        'toast.contextCleared': 'Context cleared',
        'toast.contextEmpty': 'Context is already empty',
        'toast.solutionSoon': 'Solution generation - coming soon',
        'toast.noFiles': 'No files selected',
        'toast.refreshed': 'File tree refreshed',
        'toast.refreshError': 'Failed to refresh file tree',
        'error.generic': 'An error occurred',
        'error.loadFailed': 'Failed to load data',
        'error.saveFailed': 'Failed to save',
        'error.networkError': 'Network error',
        'error.notFound': 'Not found',
    }
}

// Extend en translations
Object.assign(errorsTranslations.en, {
    'error.invalidData': 'Invalid data',
    'error.suggestions': 'Suggestions',
    'error.checkFiles': 'Check that files are within the project directory',
    'error.checkPaths': 'Verify file paths don\'t contain ".." or absolute paths',
    'error.tryRefresh': 'Try refreshing the file tree',
    'error.emptyContext': 'Context built but empty',
    'error.tokenLimitExceeded': 'Context exceeds token limit: {actual}K (limit: {limit}K). Reduce file selection or increase limit in settings.',
    'error.tokenLimitGeneric': 'Context exceeds token limit. Reduce file selection or increase limit in settings.',
})
