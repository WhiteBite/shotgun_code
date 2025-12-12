# 🎨 АГЕНТ 2: Полировка UI и Accessibility

## Роль

Ты — UI/UX Designer-Developer с экспертизой в Vue 3, CSS, и Web Accessibility (WCAG). Твоя задача — улучшить пользовательский опыт, добавить polish и сделать приложение доступным для всех пользователей.

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
└── assets/styles/      # CSS модули (variables, buttons, forms, etc.)
```

## Твои задачи (9 штук)

---

### 🟡 СРЕДНИЕ (UI Polish)

#### Задача 1: UI-8 — AI Settings: маскировка API ключа

**Файлы:**

- `frontend/src/features/ai-chat/ui/AISettingsPanel.vue`

**Проблема:**
Поле API ключ показывает частично замаскированное значение "••" но реальные символы видны. Маскировка работает некорректно.

**Где смотреть:**
Правая панель → иконка шестерёнки (Настройки AI) → секция "API Ключ"

**Решение:**

```vue
<template>
  <div class="form-group">
    <label class="form-label">{{ t("settings.apiKey") }}</label>
    <div class="relative">
      <input
        :type="showApiKey ? 'text' : 'password'"
        v-model="apiKey"
        class="input pr-10"
        :placeholder="t('settings.apiKeyPlaceholder')"
        autocomplete="off"
      />
      <button
        type="button"
        @click="showApiKey = !showApiKey"
        class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-200 transition-colors"
        :title="showApiKey ? t('settings.hideKey') : t('settings.showKey')"
        :aria-label="showApiKey ? t('settings.hideKey') : t('settings.showKey')"
      >
        <EyeIcon v-if="!showApiKey" class="w-4 h-4" />
        <EyeOffIcon v-else class="w-4 h-4" />
      </button>
    </div>
    <p class="text-xs text-gray-500 mt-1">{{ t("settings.apiKeyHint") }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { EyeIcon, EyeOffIcon } from "lucide-vue-next";

const showApiKey = ref(false);
</script>
```

**Локализация (добавить в i18n/settings.ts):**

```typescript
settings: {
  apiKey: { ru: 'API Ключ', en: 'API Key' },
  apiKeyPlaceholder: { ru: 'Введите API ключ', en: 'Enter API key' },
  apiKeyHint: { ru: 'Ключ хранится локально и не передаётся на сервер', en: 'Key is stored locally and never sent to server' },
  showKey: { ru: 'Показать ключ', en: 'Show key' },
  hideKey: { ru: 'Скрыть ключ', en: 'Hide key' }
}
```

**DoD:**

- [ ] API ключ полностью скрыт по умолчанию (type="password")
- [ ] Кнопка toggle показывает/скрывает ключ
- [ ] Иконка меняется (Eye ↔ EyeOff)
- [ ] Tooltip и aria-label на кнопке
- [ ] Локализация ru/en

---

#### Задача 2: UI-9 — AI Chat: чекбокс "Использовать контекст"

**Файлы:**

- `frontend/src/features/ai-chat/ui/ChatPanel.vue`
- `frontend/src/features/ai-chat/ui/ChatInput.vue`

**Проблема:**
Чекбокс "Использовать контекст" всегда disabled и checked. Пользователь не понимает почему нельзя его изменить.

**Где смотреть:**
Правая панель → AI Чат → чекбокс над полем ввода сообщения.

**Решение (вариант 1 — сделать интерактивным):**

```vue
<template>
  <div class="flex items-center gap-2 mb-2">
    <input
      type="checkbox"
      id="use-context"
      v-model="useContext"
      :disabled="!hasContext"
      class="checkbox"
    />
    <label
      for="use-context"
      class="text-sm"
      :class="hasContext ? 'text-gray-200' : 'text-gray-500'"
    >
      {{ t("chat.useContext") }}
    </label>
    <span
      v-if="!hasContext"
      class="text-xs text-gray-500"
      :title="t('chat.buildContextFirst')"
    >
      ({{ t("chat.noContext") }})
    </span>
  </div>
</template>
```

**Решение (вариант 2 — убрать если всегда включён):**
Если контекст всегда используется автоматически, убрать чекбокс и показать информационный текст:

```vue
<div
  v-if="hasContext"
  class="flex items-center gap-2 text-xs text-green-400 mb-2"
>
  <CheckCircleIcon class="w-4 h-4" />
  {{ t('chat.contextAttached') }}
</div>
<div v-else class="flex items-center gap-2 text-xs text-gray-500 mb-2">
  <InfoIcon class="w-4 h-4" />
  {{ t('chat.noContextHint') }}
</div>
```

**Локализация:**

```typescript
chat: {
  useContext: { ru: 'Использовать контекст', en: 'Use context' },
  noContext: { ru: 'нет контекста', en: 'no context' },
  buildContextFirst: { ru: 'Сначала постройте контекст', en: 'Build context first' },
  contextAttached: { ru: 'Контекст будет использован', en: 'Context will be used' },
  noContextHint: { ru: 'Постройте контекст для улучшения ответов AI', en: 'Build context to improve AI responses' }
}
```

**DoD:**

- [ ] Чекбокс либо интерактивный, либо заменён на информационный текст
- [ ] Понятно почему disabled (если disabled)
- [ ] Локализация ru/en

---

### 🟢 UX УЛУЧШЕНИЯ

#### Задача 3: UX-1 — Onboarding tour для новых пользователей

**Файлы:**

- `frontend/src/components/onboarding/OnboardingTour.vue` (создать)
- `frontend/src/stores/ui.store.ts`
- `frontend/src/App.vue`

**Проблема:**
Новый пользователь не понимает flow приложения. Нет подсказок что делать первым.

**Решение:**
Использовать библиотеку driver.js (лёгкая, без зависимостей):

```bash
npm install driver.js
```

```typescript
// composables/useOnboarding.ts
import { driver } from "driver.js";
import "driver.js/dist/driver.css";
import { useI18n } from "./useI18n";

export function useOnboarding() {
  const { t } = useI18n();

  const ONBOARDING_KEY = "shotgun-onboarding-completed";

  const steps = [
    {
      element: '[data-tour="file-tree"]',
      popover: {
        title: t("onboarding.step1Title"),
        description: t("onboarding.step1Desc"),
        side: "right",
      },
    },
    {
      element: '[data-tour="build-button"]',
      popover: {
        title: t("onboarding.step2Title"),
        description: t("onboarding.step2Desc"),
        side: "top",
      },
    },
    {
      element: '[data-tour="context-preview"]',
      popover: {
        title: t("onboarding.step3Title"),
        description: t("onboarding.step3Desc"),
        side: "left",
      },
    },
    {
      element: '[data-tour="ai-chat"]',
      popover: {
        title: t("onboarding.step4Title"),
        description: t("onboarding.step4Desc"),
        side: "left",
      },
    },
  ];

  function startTour() {
    const driverObj = driver({
      showProgress: true,
      steps,
      onDestroyStarted: () => {
        localStorage.setItem(ONBOARDING_KEY, "true");
        driverObj.destroy();
      },
    });
    driverObj.drive();
  }

  function shouldShowTour(): boolean {
    return !localStorage.getItem(ONBOARDING_KEY);
  }

  function resetTour() {
    localStorage.removeItem(ONBOARDING_KEY);
  }

  return { startTour, shouldShowTour, resetTour };
}
```

**Добавить data-tour атрибуты в компоненты:**

```vue
<!-- FileExplorer.vue -->
<div data-tour="file-tree" class="file-tree">...</div>
<button data-tour="build-button">{{ t('context.build') }}</button>

<!-- ContextPreview.vue -->
<div data-tour="context-preview">...</div>

<!-- ChatPanel.vue -->
<div data-tour="ai-chat">...</div>
```

**Локализация:**

```typescript
onboarding: {
  step1Title: { ru: 'Выберите файлы', en: 'Select files' },
  step1Desc: { ru: 'Отметьте файлы и папки которые хотите включить в контекст', en: 'Check files and folders to include in context' },
  step2Title: { ru: 'Постройте контекст', en: 'Build context' },
  step2Desc: { ru: 'Нажмите эту кнопку чтобы собрать контекст из выбранных файлов', en: 'Click this button to build context from selected files' },
  step3Title: { ru: 'Просмотрите результат', en: 'Review result' },
  step3Desc: { ru: 'Здесь отображается собранный контекст', en: 'Built context is displayed here' },
  step4Title: { ru: 'Используйте AI', en: 'Use AI' },
  step4Desc: { ru: 'Задавайте вопросы AI с учётом контекста вашего проекта', en: 'Ask AI questions with your project context' }
}
```

**DoD:**

- [ ] Tour запускается при первом открытии приложения
- [ ] 4 шага с подсветкой элементов
- [ ] Можно пропустить или закрыть
- [ ] Не показывается повторно (localStorage)
- [ ] Кнопка "Показать тур" в Help/Settings для повторного запуска

---

#### Задача 4: UX-3 — Undo/Redo для выбора файлов

**Файлы:**

- `frontend/src/composables/useUndoRedo.ts` (создать)
- `frontend/src/features/files/model/file.store.ts`

**Проблема:**
Если пользователь случайно снял выбор с файлов, нет способа отменить.

**Решение:**

```typescript
// composables/useUndoRedo.ts
import { ref, computed } from "vue";

export function useUndoRedo<T>(initialState: T, maxHistory = 20) {
  const history = ref<T[]>([initialState]);
  const currentIndex = ref(0);

  const current = computed(() => history.value[currentIndex.value]);
  const canUndo = computed(() => currentIndex.value > 0);
  const canRedo = computed(() => currentIndex.value < history.value.length - 1);

  function push(state: T) {
    // Удаляем "будущее" если мы в середине истории
    history.value = history.value.slice(0, currentIndex.value + 1);
    history.value.push(state);

    // Ограничиваем размер истории
    if (history.value.length > maxHistory) {
      history.value.shift();
    } else {
      currentIndex.value++;
    }
  }

  function undo(): T | undefined {
    if (canUndo.value) {
      currentIndex.value--;
      return current.value;
    }
  }

  function redo(): T | undefined {
    if (canRedo.value) {
      currentIndex.value++;
      return current.value;
    }
  }

  function clear() {
    history.value = [history.value[currentIndex.value]];
    currentIndex.value = 0;
  }

  return { current, canUndo, canRedo, push, undo, redo, clear };
}
```

**Интеграция в file.store.ts:**

```typescript
// file.store.ts
import { useUndoRedo } from "@/composables/useUndoRedo";

export const useFileStore = defineStore("file", () => {
  const selectionHistory = useUndoRedo<Set<string>>(new Set());

  const selectedFiles = computed(() => selectionHistory.current.value);

  function toggleSelection(path: string) {
    const newSelection = new Set(selectedFiles.value);
    if (newSelection.has(path)) {
      newSelection.delete(path);
    } else {
      newSelection.add(path);
    }
    selectionHistory.push(newSelection);
  }

  function undoSelection() {
    selectionHistory.undo();
  }

  function redoSelection() {
    selectionHistory.redo();
  }

  return {
    selectedFiles,
    canUndo: selectionHistory.canUndo,
    canRedo: selectionHistory.canRedo,
    toggleSelection,
    undoSelection,
    redoSelection,
  };
});
```

**UI кнопки (в FileExplorer.vue):**

```vue
<div class="flex gap-1">
  <button 
    @click="undoSelection"
    :disabled="!canUndo"
    :title="t('files.undo')"
    class="action-btn"
  >
    <UndoIcon class="w-4 h-4" />
  </button>
  <button 
    @click="redoSelection"
    :disabled="!canRedo"
    :title="t('files.redo')"
    class="action-btn"
  >
    <RedoIcon class="w-4 h-4" />
  </button>
</div>
```

**DoD:**

- [ ] Ctrl+Z отменяет последний выбор
- [ ] Ctrl+Y/Ctrl+Shift+Z повторяет
- [ ] Кнопки Undo/Redo в UI
- [ ] История ограничена 20 состояниями

---

#### Задача 5: UX-4 — Confirmation при очистке выбора

**Файлы:**

- `frontend/src/features/files/ui/FileExplorer.vue`
- `frontend/src/components/ConfirmDialog.vue` (создать если нет)

**Проблема:**
Кнопка "Очистить" сбрасывает весь выбор без подтверждения.

**Решение:**

```vue
<script setup lang="ts">
import { useConfirm } from "@/composables/useConfirm";

const { confirm } = useConfirm();

async function handleClearSelection() {
  const count = selectedFiles.value.size;

  // Показываем confirmation только если выбрано много файлов
  if (count > 5) {
    const confirmed = await confirm({
      title: t("files.clearConfirmTitle"),
      message: t("files.clearConfirmMessage", { count }),
      confirmText: t("common.clear"),
      cancelText: t("common.cancel"),
      variant: "warning",
    });

    if (!confirmed) return;
  }

  clearSelection();
}
</script>
```

**Composable useConfirm:**

```typescript
// composables/useConfirm.ts
import { ref } from "vue";

interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "info" | "warning" | "danger";
}

const isOpen = ref(false);
const options = ref<ConfirmOptions | null>(null);
let resolvePromise: ((value: boolean) => void) | null = null;

export function useConfirm() {
  function confirm(opts: ConfirmOptions): Promise<boolean> {
    options.value = opts;
    isOpen.value = true;

    return new Promise((resolve) => {
      resolvePromise = resolve;
    });
  }

  function handleConfirm() {
    isOpen.value = false;
    resolvePromise?.(true);
  }

  function handleCancel() {
    isOpen.value = false;
    resolvePromise?.(false);
  }

  return { isOpen, options, confirm, handleConfirm, handleCancel };
}
```

**Локализация:**

```typescript
files: {
  clearConfirmTitle: { ru: 'Очистить выбор?', en: 'Clear selection?' },
  clearConfirmMessage: { ru: 'Вы уверены что хотите снять выбор с {count} файлов?', en: 'Are you sure you want to deselect {count} files?' }
}
```

**DoD:**

- [ ] Confirmation показывается если выбрано > 5 файлов
- [ ] Диалог с кнопками "Очистить" / "Отмена"
- [ ] Локализация ru/en

---

### 🟢 ДИЗАЙН И ACCESSIBILITY

#### Задача 6: DES-1 — Tooltips для иконок в header

**Файлы:**

- `frontend/src/components/workspace/RightSidebar.vue`

**Проблема:**
В правом верхнем углу 6 иконок без подписей. Непонятно что они делают.

**Где смотреть:**
Правый верхний угол приложения, ряд иконок-кнопок.

**Решение:**

```vue
<template>
  <div class="flex items-center gap-1">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      @click="setActiveTab(tab.id)"
      :class="['tab-icon-btn', { active: activeTab === tab.id }]"
      :title="t(`toolbar.${tab.id}`)"
      :aria-label="t(`toolbar.${tab.id}`)"
      :aria-pressed="activeTab === tab.id"
    >
      <component :is="tab.icon" class="w-5 h-5" />
    </button>
  </div>
</template>

<script setup lang="ts">
const tabs = [
  { id: "statistics", icon: ChartBarIcon },
  { id: "export", icon: ArrowDownTrayIcon },
  { id: "prompts", icon: DocumentTextIcon },
  { id: "chat", icon: ChatBubbleIcon },
  { id: "memory", icon: CircleStackIcon },
  { id: "settings", icon: Cog6ToothIcon },
];
</script>
```

**Локализация:**

```typescript
toolbar: {
  statistics: { ru: 'Статистика', en: 'Statistics' },
  export: { ru: 'Экспорт', en: 'Export' },
  prompts: { ru: 'Промпты', en: 'Prompts' },
  chat: { ru: 'AI Чат', en: 'AI Chat' },
  memory: { ru: 'Память', en: 'Memory' },
  settings: { ru: 'Настройки AI', en: 'AI Settings' }
}
```

**DoD:**

- [ ] Каждая иконка имеет tooltip
- [ ] aria-label для screen readers
- [ ] aria-pressed для активной вкладки
- [ ] Локализация ru/en

---

#### Задача 7: DES-2 — Улучшить контраст текста

**Файлы:**

- `frontend/src/assets/styles/variables.css`
- Компоненты с `text-gray-500`, `text-gray-600`

**Проблема:**
Серый текст на тёмном фоне плохо читается. Не соответствует WCAG AA (контраст < 4.5:1).

**Решение:**

```css
/* variables.css */
:root {
  /* Текст - улучшенный контраст */
  --text-primary: #f3f4f6; /* gray-100 */
  --text-secondary: #d1d5db; /* gray-300 - было gray-500 */
  --text-muted: #9ca3af; /* gray-400 - было gray-600 */
  --text-disabled: #6b7280; /* gray-500 */
}
```

**Замены в компонентах:**

```
text-gray-500 → text-gray-400 (или var(--text-muted))
text-gray-600 → text-gray-500 (или var(--text-disabled))
```

**Проверка контраста:**

- Фон: #1f2937 (gray-800)
- Текст gray-400 (#9ca3af): контраст 4.64:1 ✅ WCAG AA
- Текст gray-500 (#6b7280): контраст 3.03:1 ❌ Не проходит

**DoD:**

- [ ] Весь текст проходит WCAG AA (контраст ≥ 4.5:1)
- [ ] CSS переменные для цветов текста
- [ ] Визуально текст читается легко

---

#### Задача 8: ACC-1 — ARIA labels для интерактивных элементов

**Файлы:**

- Все компоненты с кнопками-иконками

**Проблема:**
Многие кнопки без текста не имеют aria-label. Screen readers не могут их озвучить.

**Решение:**
Пройтись по всем компонентам и добавить aria-label:

```vue
<!-- Кнопки с иконками -->
<button @click="refresh" aria-label="Обновить">
  <RefreshIcon />
</button>

<!-- Кнопки toggle -->
<button
  @click="toggleExpand"
  :aria-expanded="isExpanded"
  :aria-label="isExpanded ? 'Свернуть' : 'Развернуть'"
>
  <ChevronIcon :class="{ 'rotate-90': isExpanded }" />
</button>

<!-- Чекбоксы -->
<input
  type="checkbox"
  :id="`file-${node.id}`"
  :aria-label="`Выбрать ${node.name}`"
/>

<!-- Поиск -->
<input type="search" aria-label="Поиск файлов" placeholder="Поиск..." />
```

**Чеклист компонентов для проверки:**

- [ ] `VirtualTreeRow.vue` - чекбокс, кнопка expand
- [ ] `FileExplorer.vue` - кнопки действий
- [ ] `BottomBar.vue` - кнопки копировать/экспорт
- [ ] `RightSidebar.vue` - иконки вкладок
- [ ] `ChatInput.vue` - кнопка отправки
- [ ] `AISettingsPanel.vue` - все кнопки

**DoD:**

- [ ] Все кнопки-иконки имеют aria-label
- [ ] Все чекбоксы имеют связанный label или aria-label
- [ ] Все input имеют aria-label или связанный label
- [ ] Screen reader может озвучить все интерактивные элементы

---

#### Задача 9: ACC-2 — Focus indicators

**Файлы:**

- `frontend/src/assets/styles/buttons.css`
- `frontend/src/assets/styles/forms.css`
- `frontend/src/assets/main.css`

**Проблема:**
При навигации клавиатурой не видно какой элемент в фокусе.

**Решение:**

```css
/* main.css или accessibility.css */

/* Убираем outline только для мыши, оставляем для клавиатуры */
:focus {
  outline: none;
}

:focus-visible {
  outline: 2px solid var(--color-primary, #6366f1);
  outline-offset: 2px;
}

/* Кнопки */
.btn:focus-visible,
.action-btn:focus-visible,
.tab-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.2);
}

/* Инпуты */
.input:focus-visible,
input:focus-visible,
textarea:focus-visible {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

/* Чекбоксы */
input[type="checkbox"]:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* Элементы дерева */
.tree-row:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
  background-color: rgba(99, 102, 241, 0.1);
}

/* Skip link для accessibility */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: var(--color-primary);
  color: white;
  padding: 8px 16px;
  z-index: 100;
  transition: top 0.3s;
}

.skip-link:focus {
  top: 0;
}
```

**Добавить skip link в App.vue:**

```vue
<template>
  <a href="#main-content" class="skip-link">
    {{ t("accessibility.skipToContent") }}
  </a>
  <main id="main-content">
    <!-- content -->
  </main>
</template>
```

**DoD:**

- [ ] Все интерактивные элементы имеют видимый focus indicator
- [ ] Focus indicator контрастный (видно на тёмном фоне)
- [ ] Tab навигация работает логично
- [ ] Skip link для быстрого перехода к контенту

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

### Accessibility

- Все интерактивные элементы доступны с клавиатуры
- aria-label на кнопках без текста
- Контраст текста ≥ 4.5:1 (WCAG AA)
- Focus indicators видимые

### Тестирование

После каждого изменения:

1. `npm run build` — должен проходить
2. Проверить в браузере (http://localhost:34115)
3. Проверить Tab навигацию
4. Проверить с screen reader (опционально)

### Git

Коммиты по задачам:

```
fix(settings): add proper API key masking [UI-8]
feat(onboarding): add tour for new users [UX-1]
fix(a11y): add ARIA labels to icon buttons [ACC-1]
fix(a11y): improve text contrast [DES-2]
```

---

## Definition of Done (общий)

- [ ] Все 9 задач выполнены
- [ ] `npm run build` проходит без ошибок
- [ ] Все тексты локализованы (ru + en)
- [ ] UI выглядит красиво и консистентно
- [ ] Accessibility: Tab навигация работает, focus видимый
- [ ] Код соответствует конвенциям проекта
