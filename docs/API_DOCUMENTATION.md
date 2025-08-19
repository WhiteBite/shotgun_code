# 📚 API Documentation

## Обзор

Shotgun Code использует Wails v2 для связи между frontend (Vue.js) и backend (Go). Все взаимодействие происходит через RPC методы, экспортированные из `backend/app.go`.

## 🔗 Архитектура

```
Frontend (Vue.js) ←→ Wails Bridge ←→ Backend (Go)
```

Все методы автоматически доступны в frontend через `wailsjs/go/main/App.js`.

## 📋 RPC Методы

### Проекты и файлы

#### `SelectDirectory()`
Открывает диалог выбора директории.

**Параметры:** нет

**Возвращает:** `string` - путь к выбранной директории

**Пример:**
```typescript
import { SelectDirectory } from 'wailsjs/go/main/App';

const projectPath = await SelectDirectory();
console.log('Выбранная директория:', projectPath);
```

#### `ListFiles(dirPath: string, useGitignore: boolean, useCustomIgnore: boolean)`
Получает список файлов в директории.

**Параметры:**
- `dirPath` - путь к директории
- `useGitignore` - использовать .gitignore
- `useCustomIgnore` - использовать пользовательские правила игнорирования

**Возвращает:** `FileNode[]` - массив файлов

**Пример:**
```typescript
import { ListFiles } from 'wailsjs/go/main/App';

const files = await ListFiles('/path/to/project', true, true);
console.log('Найдено файлов:', files.length);
```

#### `RequestShotgunContextGeneration(rootDir: string, includedPaths: string[])`
Запускает генерацию контекста для выбранных файлов.

**Параметры:**
- `rootDir` - корневая директория проекта
- `includedPaths` - массив путей к файлам для включения

**Возвращает:** `void`

**Пример:**
```typescript
import { RequestShotgunContextGeneration } from 'wailsjs/go/main/App';

await RequestShotgunContextGeneration('/path/to/project', [
  'src/main.go',
  'src/utils.go'
]);
```

### ИИ и генерация кода

#### `GenerateCode(systemPrompt: string, userPrompt: string)`
Генерирует код с помощью ИИ.

**Параметры:**
- `systemPrompt` - системный промпт
- `userPrompt` - пользовательский промпт

**Возвращает:** `string` - сгенерированный код

**Пример:**
```typescript
import { GenerateCode } from 'wailsjs/go/main/App';

const code = await GenerateCode(
  'Ты опытный Go разработчик',
  'Создай функцию для парсинга JSON'
);
```

#### `GenerateIntelligentCode(task: string, context: string, optionsJson: string)`
Интеллектуальная генерация кода с расширенными опциями.

**Параметры:**
- `task` - описание задачи
- `context` - контекст проекта
- `optionsJson` - JSON с опциями генерации

**Возвращает:** `string` - JSON с результатом

**Пример:**
```typescript
import { GenerateIntelligentCode } from 'wailsjs/go/main/App';

const options = {
  provider: 'openai',
  model: 'gpt-4',
  temperature: 0.7,
  maxTokens: 2000,
  strategy: 'balanced'
};

const result = await GenerateIntelligentCode(
  'Создай REST API для пользователей',
  context,
  JSON.stringify(options)
);

const parsedResult = JSON.parse(result);
console.log('Сгенерированный код:', parsedResult.code);
```

#### `GenerateCodeWithOptions(systemPrompt: string, userPrompt: string, optionsJson: string)`
Генерация кода с дополнительными опциями.

**Параметры:**
- `systemPrompt` - системный промпт
- `userPrompt` - пользовательский промпт
- `optionsJson` - JSON с опциями

**Возвращает:** `string` - сгенерированный код

#### `GetProviderInfo()`
Получает информацию о текущем ИИ провайдере.

**Параметры:** нет

**Возвращает:** `string` - JSON с информацией о провайдере

**Пример:**
```typescript
import { GetProviderInfo } from 'wailsjs/go/main/App';

const providerInfo = await GetProviderInfo();
const info = JSON.parse(providerInfo);
console.log('Провайдер:', info.name);
console.log('Модели:', info.models);
```

#### `ListAvailableModels()`
Получает список доступных моделей ИИ.

**Параметры:** нет

**Возвращает:** `string[]` - массив названий моделей

### Анализ контекста

#### `SuggestContextFiles(task: string, allFiles: FileNode[])`
Предлагает файлы для контекста на основе задачи.

**Параметры:**
- `task` - описание задачи
- `allFiles` - все файлы проекта

**Возвращает:** `string[]` - массив путей к рекомендуемым файлам

#### `AnalyzeTaskAndCollectContext(task: string, allFilesJson: string, rootDir: string)`
Интеллектуально анализирует задачу и автоматически собирает контекст.

**Параметры:**
- `task` - описание задачи
- `allFilesJson` - JSON с файлами проекта
- `rootDir` - корневая директория

**Возвращает:** `string` - JSON с результатом анализа

**Пример:**
```typescript
import { AnalyzeTaskAndCollectContext } from 'wailsjs/go/main/App';

const result = await AnalyzeTaskAndCollectContext(
  'Исправь баг в функции парсинга',
  JSON.stringify(allFiles),
  '/path/to/project'
);

const analysis = JSON.parse(result);
console.log('Рекомендуемые файлы:', analysis.suggestedFiles);
console.log('Уверенность:', analysis.confidence);
```

### Настройки

#### `GetSettings()`
Получает текущие настройки приложения.

**Параметры:** нет

**Возвращает:** `SettingsDTO` - объект с настройками

#### `SaveSettings(settingsJson: string)`
Сохраняет настройки приложения.

**Параметры:**
- `settingsJson` - JSON с настройками

**Возвращает:** `void`

**Пример:**
```typescript
import { GetSettings, SaveSettings } from 'wailsjs/go/main/App';

// Получение настроек
const settings = await GetSettings();

// Изменение настроек
settings.ai.defaultProvider = 'openai';
settings.ai.maxTokens = 4000;

// Сохранение
await SaveSettings(JSON.stringify(settings));
```

#### `RefreshAIModels(provider: string, apiKey: string)`
Обновляет список моделей для ИИ провайдера.

**Параметры:**
- `provider` - название провайдера
- `apiKey` - API ключ

**Возвращает:** `void`

### Git интеграция

#### `IsGitAvailable()`
Проверяет доступность Git.

**Параметры:** нет

**Возвращает:** `boolean` - доступен ли Git

#### `GetUncommittedFiles(projectRoot: string)`
Получает список незакоммиченных файлов.

**Параметры:**
- `projectRoot` - корневая директория проекта

**Возвращает:** `FileStatus[]` - массив статусов файлов

#### `GetRichCommitHistory(projectRoot: string, branchName: string, limit: number)`
Получает историю коммитов с деталями.

**Параметры:**
- `projectRoot` - корневая директория проекта
- `branchName` - название ветки
- `limit` - максимальное количество коммитов

**Возвращает:** `CommitWithFiles[]` - массив коммитов с файлами

#### `GetFileContentAtCommit(projectRoot: string, filePath: string, commitHash: string)`
Получает содержимое файла на определенном коммите.

**Параметры:**
- `projectRoot` - корневая директория проекта
- `filePath` - путь к файлу
- `commitHash` - хеш коммита

**Возвращает:** `string` - содержимое файла

#### `GetGitignoreContent(projectRoot: string)`
Получает содержимое .gitignore файла.

**Параметры:**
- `projectRoot` - корневая директория проекта

**Возвращает:** `string` - содержимое .gitignore

### Файловые операции

#### `ReadFileContent(rootDir: string, relPath: string)`
Читает содержимое файла.

**Параметры:**
- `rootDir` - корневая директория
- `relPath` - относительный путь к файлу

**Возвращает:** `string` - содержимое файла

#### `StartFileWatcher(rootDirPath: string)`
Запускает отслеживание изменений файлов.

**Параметры:**
- `rootDirPath` - путь к корневой директории

**Возвращает:** `void`

#### `StopFileWatcher()`
Останавливает отслеживание изменений файлов.

**Параметры:** нет

**Возвращает:** `void`

### Экспорт

#### `ExportContext(settingsJson: string)`
Экспортирует контекст в различных форматах.

**Параметры:**
- `settingsJson` - JSON с настройками экспорта

**Возвращает:** `ExportResult` - результат экспорта

**Пример:**
```typescript
import { ExportContext } from 'wailsjs/go/main/App';

const exportSettings = {
  format: 'markdown',
  includeMetadata: true,
  includeFileStructure: true,
  files: ['src/main.go', 'src/utils.go']
};

const result = await ExportContext(JSON.stringify(exportSettings));

if (result.filePath) {
  console.log('Файл сохранен:', result.filePath);
} else {
  console.log('Контент:', result.content);
}
```

#### `CleanupTempFiles(filePath: string)`
Очищает временные файлы экспорта.

**Параметры:**
- `filePath` - путь к временному файлу

**Возвращает:** `void`

### Тестирование

#### `TestBackend(allFilesJson: string, rootDir: string)`
Простой тест для проверки работы backend.

**Параметры:**
- `allFilesJson` - JSON с файлами
- `rootDir` - корневая директория

**Возвращает:** `string` - JSON с результатом теста

## 📊 Типы данных

### FileNode
```typescript
interface FileNode {
  name: string;
  relPath: string;
  absPath: string;
  isDir: boolean;
  size: number;
  modTime: string;
  children?: FileNode[];
}
```

### SettingsDTO
```typescript
interface SettingsDTO {
  ai: {
    defaultProvider: string;
    defaultModel: string;
    maxTokens: number;
    temperature: number;
  };
  context: {
    compressionEnabled: boolean;
    maxFileSize: number;
    excludedPatterns: string[];
  };
  export: {
    defaultFormat: string;
    includeMetadata: boolean;
  };
}
```

### FileStatus
```typescript
interface FileStatus {
  path: string;
  status: 'modified' | 'added' | 'deleted' | 'untracked';
  staged: boolean;
}
```

### CommitWithFiles
```typescript
interface CommitWithFiles {
  hash: string;
  author: string;
  date: string;
  message: string;
  files: string[];
}
```

### ExportResult
```typescript
interface ExportResult {
  content?: string;
  filePath?: string;
  format: string;
  size: number;
  error?: string;
}
```

## 🔧 Примеры использования

### Полный цикл работы с проектом

```typescript
import { 
  SelectDirectory, 
  ListFiles, 
  RequestShotgunContextGeneration,
  GenerateCode 
} from 'wailsjs/go/main/App';

async function analyzeProject() {
  // 1. Выбор проекта
  const projectPath = await SelectDirectory();
  
  // 2. Получение списка файлов
  const files = await ListFiles(projectPath, true, true);
  
  // 3. Генерация контекста
  const relevantFiles = files
    .filter(f => f.relPath.endsWith('.go'))
    .map(f => f.relPath);
  
  await RequestShotgunContextGeneration(projectPath, relevantFiles);
  
  // 4. Генерация кода
  const code = await GenerateCode(
    'Ты опытный Go разработчик',
    'Добавь обработку ошибок в функцию main'
  );
  
  console.log('Сгенерированный код:', code);
}
```

### Работа с настройками

```typescript
import { GetSettings, SaveSettings } from 'wailsjs/go/main/App';

async function configureAI() {
  const settings = await GetSettings();
  
  // Настройка ИИ
  settings.ai.defaultProvider = 'openai';
  settings.ai.defaultModel = 'gpt-4';
  settings.ai.maxTokens = 4000;
  settings.ai.temperature = 0.7;
  
  // Настройка контекста
  settings.context.compressionEnabled = true;
  settings.context.maxFileSize = 1024 * 1024; // 1MB
  
  await SaveSettings(JSON.stringify(settings));
}
```

## 📝 Обработка ошибок

Все методы могут возвращать ошибки. Рекомендуется использовать try-catch:

```typescript
import { GenerateCode } from 'wailsjs/go/main/App';

try {
  const code = await GenerateCode('Системный промпт', 'Задача');
  console.log('Успех:', code);
} catch (error) {
  console.error('Ошибка:', error);
  // Обработка ошибки
}
```

## 🔄 События

Wails поддерживает события для real-time обновлений:

```typescript
import { EventsOn, EventsOff } from 'wailsjs/runtime/runtime';

// Подписка на события
EventsOn('file-changed', (data) => {
  console.log('Файл изменился:', data);
});

// Отписка
EventsOff('file-changed');
```

## 📚 Дополнительная информация

- [Wails Documentation](https://wails.io/docs/)
- [Vue.js Integration](https://wails.io/docs/guides/frontend/vue)
- [Runtime API](https://wails.io/docs/reference/runtime/intro)