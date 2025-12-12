# 🤖 АГЕНТ 1: Критические баги и Core UX

## Роль

Ты — Senior Frontend Developer с экспертизой в Vue 3, TypeScript и UX. Твоя задача — исправить критические баги, которые ломают функциональность приложения и создают плохой пользовательский опыт.

## Контекст проекта

**Shotgun Code** — десктопное приложение (Wails + Vue 3 + TypeScript + Pinia + Tailwind) для построения контекста проекта с AI анализом.

Структура frontend:

```
frontend/src/
├── features/           # Feature modules (files/, context/, ai-chat/, git/)
│   └── [name]/
│       ├── model/      # Pinia stores
│       ├── ui/         # Vue компоненты
│       └── composables/
├── components/         # Shared компоненты
├── composables/        # Глобальные composables
│   └── i18n/           # Локализация (ru/en)
├── stores/             # Глобальные stores
└── assets/styles/      # CSS модули
```

## Твои задачи (10 штук)

### 🔴 КРИТИЧЕСКИЕ (сделать первыми)

#### Задача 1: UI-1 — Vue injection error в VirtualFileTree

**Приоритет:** КРИТИЧЕСКИЙ  
**Файлы:**

- `frontend/src/features/files/ui/VirtualFileTree.vue`
- `frontend/src/features/files/ui/VirtualTreeRow.vue`

**Проблема:**
В консоли браузера массовые warnings при каждом рендере дерева:

```
[Vue warn]: provide() can only be used inside setup()
[Vue warn]: injection "Symbol(HoveredFile)" not found
```

**Причина:**
`provide()` вызывается в VirtualTreeRow, но при виртуализации (RecycleScroller) строки переиспользуются и provide() вызывается повторно вне setup().

**Решение:**

1. Найти где используется `provide('HoveredFile', ...)` или `provide(Symbol(...), ...)`
2. Перенести `provide()` в родительский VirtualFileTree.vue (в setup)
3. В VirtualTreeRow оставить только `inject()`
4. Альтернатива: использовать composable с reactive state вместо provide/inject

**Пример:**

```typescript
// VirtualFileTree.vue — setup()
const hoveredFileId = ref<string | null>(null);
provide("HoveredFile", {
  hoveredFileId,
  setHovered: (id: string | null) => (hoveredFileId.value = id),
});

// VirtualTreeRow.vue — setup()
const { hoveredFileId, setHovered } = inject("HoveredFile")!;
// НЕ вызывать provide() здесь!
```

**DoD:**

- [ ] Консоль браузера чистая от Vue warnings
- [ ] Hover на файлах работает корректно
- [ ] Виртуализация не сломана

---

#### Задача 2: UI-2 — Дерево файлов сбрасывается при переключении вкладок

**Приоритет:** КРИТИЧЕСКИЙ  
**Файлы:**

- `frontend/src/features/files/model/file.store.ts`
- `frontend/src/components/workspace/LeftSidebar.vue`

**Проблема:**
При переключении Файлы → Git → Файлы дерево сворачивается. Пользователь теряет контекст.

**Причина:**
Компонент FileExplorer пересоздаётся (v-if) вместо скрытия (v-show), или expandedNodes не персистится.

**Решение:**

1. В LeftSidebar.vue заменить `v-if` на `v-show` для вкладок
2. ИЛИ сохранять expandedNodes в localStorage:

```typescript
// file.store.ts
export const useFileStore = defineStore("file", () => {
  const expandedNodes = ref<Set<string>>(new Set());

  // Восстановить при инициализации
  onMounted(() => {
    const saved = localStorage.getItem("file-expanded-nodes");
    if (saved) expandedNodes.value = new Set(JSON.parse(saved));
  });

  // Сохранять при изменении
  watch(
    expandedNodes,
    (val) => {
      localStorage.setItem("file-expanded-nodes", JSON.stringify([...val]));
    },
    { deep: true }
  );

  return { expandedNodes };
});
```

**DoD:**

- [ ] Переключение вкладок не сбрасывает раскрытые папки
- [ ] После перезагрузки страницы состояние восстанавливается
- [ ] Нет лишних ререндеров

---

#### Задача 3: UI-3 — Отсутствует favicon

**Приоритет:** НИЗКИЙ (но быстрый фикс)  
**Файлы:**

- `frontend/public/` (создать favicon)
- `frontend/index.html`

**Проблема:**
Ошибка в консоли: `404 /favicon.ico`

**Решение:**

1. Создать или скопировать favicon.ico в `frontend/public/`
2. Или использовать SVG логотип:

```html
<!-- index.html -->
<link rel="icon" type="image/svg+xml" href="/logo.svg" />
```

**DoD:**

- [ ] Нет 404 ошибки в консоли
- [ ] Иконка отображается во вкладке браузера

---

### 🟠 ВЫСОКИЕ (Core UX)

#### Задача 4: UI-6 — Кнопка "Построить контекст" disabled без объяснения

**Файлы:**

- `frontend/src/features/files/ui/FileExplorer.vue`
- `frontend/src/composables/i18n/context.ts`

**Проблема:**
Кнопка disabled, но пользователь не понимает почему.

**Решение:**

```vue
<button
  :disabled="selectedCount === 0"
  :title="
    selectedCount === 0
      ? t('context.selectFilesFirst')
      : t('context.buildTooltip')
  "
  class="btn-primary"
>
  <BuildIcon />
  {{ t('context.build') }}
  <span v-if="selectedCount > 0" class="badge">{{ selectedCount }}</span>
</button>

<!-- Подсказка под кнопкой -->
<p v-if="selectedCount === 0" class="text-xs text-gray-400 mt-1 text-center">
  {{ t('context.selectFilesHint') }}
</p>
```

**Локализация (добавить):**

```typescript
// i18n/context.ts
context: {
  selectFilesFirst: { ru: 'Сначала выберите файлы', en: 'Select files first' },
  selectFilesHint: { ru: 'Выберите файлы в дереве выше', en: 'Select files in the tree above' },
  buildTooltip: { ru: 'Построить контекст из выбранных файлов', en: 'Build context from selected files' }
}
```

**DoD:**

- [ ] Tooltip объясняет почему кнопка disabled
- [ ] Текст-подсказка под кнопкой когда ничего не выбрано
- [ ] Локализация на ru и en

---

#### Задача 5: UI-10 — Кнопки "Копировать"/"Экспорт" disabled без объяснения

**Файлы:**

- `frontend/src/components/workspace/BottomBar.vue`
- `frontend/src/composables/i18n/export.ts`

**Проблема:**
Кнопки в нижней панели disabled без tooltip.

**Решение:**

```vue
<button
  :disabled="!hasContext"
  :title="!hasContext ? t('export.buildContextFirst') : t('export.copyTooltip')"
>
  <CopyIcon />
  {{ t('export.copy') }}
</button>
```

**DoD:**

- [ ] Tooltip на disabled кнопках
- [ ] Локализация

---

#### Задача 6: UI-7 — Центральная панель: пустое состояние не центрировано

**Файлы:**

- `frontend/src/features/context/ui/ContextPreview.vue`

**Проблема:**
Контент "Контекст ещё не построен" не центрирован, много пустого места.

**Решение:**

```vue
<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <div class="panel-header">...</div>

    <!-- Empty state - центрированный -->
    <div
      v-if="!hasContext"
      class="flex-1 flex flex-col items-center justify-center p-8"
    >
      <div class="text-center max-w-md">
        <DocumentIcon class="w-16 h-16 text-gray-600 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-200 mb-2">
          {{ t("context.notBuilt") }}
        </h3>
        <p class="text-gray-400 mb-6">{{ t("context.notBuiltDescription") }}</p>

        <ol class="text-left space-y-3 text-sm">
          <li class="flex items-start gap-3">
            <span
              class="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-400 flex items-center justify-center text-xs"
              >1</span
            >
            <span class="text-gray-300">{{ t("context.step1") }}</span>
          </li>
          <li class="flex items-start gap-3">
            <span
              class="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-400 flex items-center justify-center text-xs"
              >2</span
            >
            <span class="text-gray-300">{{ t("context.step2") }}</span>
          </li>
        </ol>
      </div>
    </div>

    <!-- Content when has context -->
    <div v-else class="flex-1 overflow-auto">...</div>
  </div>
</template>
```

**DoD:**

- [ ] Пустое состояние центрировано вертикально и горизонтально
- [ ] Красивый дизайн с иконкой и нумерованными шагами
- [ ] Адаптивность при изменении размера панели

---

#### Задача 7: UX-2 — Keyboard shortcuts

**Файлы:**

- `frontend/src/composables/useKeyboardShortcuts.ts` (создать)
- `frontend/src/App.vue`

**Решение:**

```typescript
// composables/useKeyboardShortcuts.ts
import { onMounted, onUnmounted } from "vue";

interface Shortcut {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  action: () => void;
  description: string;
}

export function useKeyboardShortcuts(shortcuts: Shortcut[]) {
  function handleKeydown(e: KeyboardEvent) {
    for (const s of shortcuts) {
      const ctrlMatch = s.ctrl ? e.ctrlKey || e.metaKey : !e.ctrlKey;
      const shiftMatch = s.shift ? e.shiftKey : !e.shiftKey;

      if (
        e.key.toLowerCase() === s.key.toLowerCase() &&
        ctrlMatch &&
        shiftMatch
      ) {
        e.preventDefault();
        s.action();
        return;
      }
    }
  }

  onMounted(() => document.addEventListener("keydown", handleKeydown));
  onUnmounted(() => document.removeEventListener("keydown", handleKeydown));
}

// Использование в App.vue
useKeyboardShortcuts([
  {
    key: "b",
    ctrl: true,
    action: buildContext,
    description: "Построить контекст",
  },
  {
    key: "c",
    ctrl: true,
    shift: true,
    action: copyContext,
    description: "Копировать контекст",
  },
  { key: "e", ctrl: true, action: exportContext, description: "Экспорт" },
  {
    key: "1",
    ctrl: true,
    action: () => setTab("files"),
    description: "Вкладка Файлы",
  },
  {
    key: "2",
    ctrl: true,
    action: () => setTab("git"),
    description: "Вкладка Git",
  },
  {
    key: "3",
    ctrl: true,
    action: () => setTab("contexts"),
    description: "Вкладка Контексты",
  },
]);
```

**DoD:**

- [ ] Ctrl+B строит контекст
- [ ] Ctrl+Shift+C копирует
- [ ] Ctrl+1/2/3 переключает вкладки
- [ ] Shortcuts не конфликтуют с браузерными

---

### 🟡 СРЕДНИЕ

#### Задача 8: UI-4 — Длинные пути обрезаются без tooltip

**Файлы:**

- `frontend/src/components/WelcomeScreen.vue`

**Решение:**

```vue
<div
  class="project-item"
  :title="project.path"
>
  <FolderIcon />
  <div class="flex-1 min-w-0">
    <div class="font-medium truncate">{{ project.name }}</div>
    <div class="text-xs text-gray-500 truncate">{{ project.path }}</div>
  </div>
</div>
```

**DoD:**

- [ ] Полный путь виден в tooltip при hover
- [ ] Текст корректно обрезается с ellipsis

---

#### Задача 9: UI-5 — Счётчики файлов без tooltip

**Файлы:**

- `frontend/src/features/files/ui/VirtualTreeRow.vue`
- `frontend/src/composables/i18n/files.ts`

**Решение:**

```vue
<span
  v-if="node.isDirectory && node.fileCount"
  class="file-count text-xs text-gray-500"
  :title="t('files.fileCountTooltip', { count: node.fileCount })"
>
  {{ node.fileCount }}
</span>
```

```typescript
// i18n/files.ts
files: {
  fileCountTooltip: {
    ru: 'Файлов в папке: {count}',
    en: 'Files in folder: {count}'
  }
}
```

**DoD:**

- [ ] Tooltip показывает "Файлов в папке: N"
- [ ] Локализация ru/en

---

#### Задача 10: UI-11 — Кнопка "vdev" без tooltip

**Файлы:**

- `frontend/src/components/workspace/BottomBar.vue`

**Решение:**

```vue
<button
  v-if="isDev"
  class="dev-button"
  title="Developer Tools (only in dev mode)"
  @click="openDevTools"
>
  <CodeIcon class="w-4 h-4" />
  <span>vdev</span>
</button>
```

Или скрыть в production:

```typescript
const isDev = import.meta.env.DEV;
```

**DoD:**

- [ ] Tooltip объясняет назначение
- [ ] Скрыта в production build (опционально)

---

## Правила работы

### Код

- TypeScript strict mode, никаких `any`
- Все тексты через `t()` из `useI18n()` — добавлять в ОБА языка (ru, en)
- CSS классы из `assets/styles/` — не дублировать inline
- Компоненты < 300 строк, выносить в composables

### Стиль

- Tailwind для layout (flex, grid, p-, m-)
- CSS классы для компонентов (.btn-primary, .panel-header)
- Цвета через CSS переменные (--color-primary)
- Тёмная тема по умолчанию

### Тестирование

После каждого изменения:

1. `npm run build` — должен проходить
2. Проверить в браузере (http://localhost:34115)
3. Консоль чистая от ошибок и warnings

### Git

Коммиты по задачам:

```
fix(files): resolve Vue injection error in VirtualFileTree [UI-1]
fix(files): persist expanded nodes state [UI-2]
feat(ux): add keyboard shortcuts [UX-2]
```

---

## Definition of Done (общий)

- [ ] Все 10 задач выполнены
- [ ] `npm run build` проходит без ошибок
- [ ] Консоль браузера чистая (нет errors/warnings)
- [ ] Все тексты локализованы (ru + en)
- [ ] UI выглядит красиво и консистентно
- [ ] Код соответствует конвенциям проекта
