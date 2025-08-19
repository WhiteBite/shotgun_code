# 📖 Examples

## Обзор

Этот документ содержит практические примеры использования Shotgun Code для различных сценариев разработки.

## 🎯 Сценарии использования

### 1. Анализ бага в Go проекте

#### Задача
Найти и исправить баг в функции парсинга JSON в Go проекте.

#### Шаги

1. **Сканирование проекта**
```bash
# Запуск приложения
./shotgun-code

# Выбор проекта
# Выбор файлов: main.go, parser.go, parser_test.go
```

2. **Генерация контекста**
```json
{
  "contextId": "bug-analysis-001",
  "files": [
    "main.go",
    "parser.go", 
    "parser_test.go"
  ],
  "options": {
    "includeTests": true,
    "includeImports": true,
    "maxTokens": 3000
  }
}
```

3. **Анализ с ИИ**
```
Промпт: "Проанализируй код и найди потенциальные баги в функции ParseJSON. 
Обрати внимание на обработку ошибок и edge cases."
```

4. **Результат**
- Обнаружен баг: отсутствует проверка на nil в строке 45
- Предложено исправление с тестами
- Оценка уверенности: 0.85

### 2. Рефакторинг TypeScript компонента

#### Задача
Упростить сложный Vue компонент, разбив его на более мелкие части.

#### Шаги

1. **Выбор файлов**
```
src/components/ComplexUserProfile.vue
src/components/UserProfile.vue
src/stores/user.store.ts
```

2. **Генерация контекста**
```json
{
  "contextId": "refactor-vue-001",
  "files": [
    "src/components/ComplexUserProfile.vue",
    "src/components/UserProfile.vue",
    "src/stores/user.store.ts"
  ],
  "options": {
    "includeImports": true,
    "compression": "smart"
  }
}
```

3. **Запрос к ИИ**
```
Промпт: "Разбей ComplexUserProfile.vue на более мелкие компоненты. 
Создай отдельные компоненты для: UserAvatar, UserInfo, UserActions. 
Используй Composition API и TypeScript."
```

4. **Результат**
- Создано 3 новых компонента
- Обновлен основной компонент
- Добавлены типы TypeScript
- Улучшена читаемость кода

### 3. Миграция с JavaScript на TypeScript

#### Задача
Конвертировать JavaScript файл в TypeScript с добавлением типов.

#### Шаги

1. **Выбор файла**
```
src/utils/calculator.js
```

2. **Генерация контекста**
```json
{
  "contextId": "js-to-ts-001",
  "files": ["src/utils/calculator.js"],
  "options": {
    "includeImports": false,
    "maxTokens": 2000
  }
}
```

3. **Запрос к ИИ**
```
Промпт: "Конвертируй этот JavaScript файл в TypeScript. 
Добавь строгую типизацию для всех функций и переменных. 
Создай интерфейсы для объектов."
```

4. **Результат**
- Файл переименован в `calculator.ts`
- Добавлены типы для всех функций
- Созданы интерфейсы
- Обновлены импорты

### 4. Оптимизация производительности

#### Задача
Найти узкие места в производительности Python скрипта.

#### Шаги

1. **Выбор файлов**
```
src/performance_test.py
src/optimization_target.py
requirements.txt
```

2. **Генерация контекста**
```json
{
  "contextId": "perf-optimization-001",
  "files": [
    "src/performance_test.py",
    "src/optimization_target.py"
  ],
  "options": {
    "includeImports": true,
    "maxTokens": 4000
  }
}
```

3. **Запрос к ИИ**
```
Промпт: "Проанализируй код на предмет узких мест производительности. 
Предложи оптимизации для улучшения скорости выполнения. 
Учти профилирование и кэширование."
```

4. **Результат**
- Обнаружены медленные операции
- Предложены оптимизации
- Добавлено кэширование
- Улучшена структура данных

## 🔧 Практические примеры кода

### Пример 1: Создание нового API эндпоинта

#### Backend (Go)

```go
// domain/models.go
type User struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Created  string `json:"created"`
}

// application/user_service.go
type UserService struct {
    repo domain.UserRepository
    log  domain.Logger
}

func (s *UserService) CreateUser(ctx context.Context, user domain.User) error {
    // Валидация
    if user.Name == "" {
        return domain.ErrInvalidInput
    }
    
    // Создание пользователя
    return s.repo.Create(ctx, user)
}
```

#### Frontend (TypeScript)

```typescript
// services/api.service.ts
export class ApiService {
    private baseUrl = 'http://localhost:34115/api/v1';
    
    async createUser(user: User): Promise<User> {
        const response = await fetch(`${this.baseUrl}/users`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(user)
        });
        
        if (!response.ok) {
            throw new Error('Failed to create user');
        }
        
        return response.json();
    }
}

// stores/user.store.ts
export const useUserStore = defineStore('user', () => {
    const users = ref<User[]>([]);
    const loading = ref(false);
    
    const createUser = async (userData: Partial<User>) => {
        loading.value = true;
        try {
            const user = await apiService.createUser(userData as User);
            users.value.push(user);
            return user;
        } finally {
            loading.value = false;
        }
    };
    
    return { users, loading, createUser };
});
```

### Пример 2: Интеграция с ИИ провайдером

#### Backend (Go)

```go
// infrastructure/ai/openai.go
type OpenAIProvider struct {
    client *openai.Client
    config domain.AIConfig
}

func (p *OpenAIProvider) Generate(ctx context.Context, req domain.AIRequest) (domain.AIResponse, error) {
    resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: req.Model,
        Messages: []openai.ChatCompletionMessage{
            {Role: "system", Content: req.SystemPrompt},
            {Role: "user", Content: req.UserPrompt},
        },
        Temperature: req.Temperature,
        MaxTokens:   req.MaxTokens,
    })
    
    if err != nil {
        return domain.AIResponse{}, fmt.Errorf("openai error: %w", err)
    }
    
    return domain.AIResponse{
        Content: resp.Choices[0].Message.Content,
        Usage: domain.TokenUsage{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
    }, nil
}
```

#### Frontend (TypeScript)

```typescript
// stores/ai.store.ts
export const useAIStore = defineStore('ai', () => {
    const isGenerating = ref(false);
    const lastResult = ref<AIResponse | null>(null);
    const error = ref<string | null>(null);
    
    const generateCode = async (prompt: string, context: string) => {
        isGenerating.value = true;
        error.value = null;
        
        try {
            const response = await apiService.generateAI({
                prompt,
                context,
                provider: 'openai',
                model: 'gpt-4',
                options: {
                    temperature: 0.7,
                    maxTokens: 2000
                }
            });
            
            lastResult.value = response;
            return response;
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Unknown error';
            throw err;
        } finally {
            isGenerating.value = false;
        }
    };
    
    return { isGenerating, lastResult, error, generateCode };
});
```

### Пример 3: Экспорт контекста

#### Backend (Go)

```go
// application/export_service.go
type ExportService struct {
    contextService *ContextService
    pdfGenerator   domain.PDFGenerator
}

func (s *ExportService) ExportContext(ctx context.Context, req domain.ExportRequest) (domain.ExportResponse, error) {
    // Получение контекста
    context, err := s.contextService.GetContext(ctx, req.ContextID)
    if err != nil {
        return domain.ExportResponse{}, err
    }
    
    // Экспорт в зависимости от формата
    switch req.Format {
    case "markdown":
        return s.exportMarkdown(context, req.Options)
    case "pdf":
        return s.exportPDF(context, req.Options)
    case "json":
        return s.exportJSON(context, req.Options)
    default:
        return domain.ExportResponse{}, domain.ErrUnsupportedFormat
    }
}

func (s *ExportService) exportMarkdown(ctx domain.Context, options domain.ExportOptions) (domain.ExportResponse, error) {
    var content strings.Builder
    
    // Заголовок
    content.WriteString("# Project Context\n\n")
    
    // Метаданные
    if options.IncludeMetadata {
        content.WriteString("## Metadata\n\n")
        content.WriteString(fmt.Sprintf("- Files: %d\n", len(ctx.Files)))
        content.WriteString(fmt.Sprintf("- Total Lines: %d\n", ctx.TotalLines))
        content.WriteString(fmt.Sprintf("- Generated: %s\n\n", time.Now().Format(time.RFC3339)))
    }
    
    // Содержимое файлов
    for _, file := range ctx.Files {
        content.WriteString(fmt.Sprintf("## %s\n\n", file.Path))
        content.WriteString("```" + file.Language + "\n")
        content.WriteString(file.Content)
        content.WriteString("\n```\n\n")
    }
    
    return domain.ExportResponse{
        Content: content.String(),
        Format:  "markdown",
        Size:    len(content.String()),
    }, nil
}
```

## 🎨 Примеры UI компонентов

### Компонент выбора файлов

```vue
<template>
  <div class="file-selector">
    <div class="file-tree">
      <FileTreeItem
        v-for="item in fileTree"
        :key="item.path"
        :item="item"
        :selected="selectedFiles.has(item.path)"
        @toggle="toggleFile"
      />
    </div>
    
    <div class="file-actions">
      <button @click="selectAll" :disabled="isGenerating">
        Select All
      </button>
      <button @click="clearSelection" :disabled="isGenerating">
        Clear
      </button>
      <button @click="generateContext" :disabled="selectedFiles.size === 0 || isGenerating">
        {{ isGenerating ? 'Generating...' : 'Generate Context' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useFileTreeStore } from '@/stores/file-tree.store';
import { useContextStore } from '@/stores/context.store';
import FileTreeItem from './FileTreeItem.vue';

const fileTreeStore = useFileTreeStore();
const contextStore = useContextStore();

const selectedFiles = ref(new Set<string>());
const isGenerating = ref(false);

const fileTree = computed(() => fileTreeStore.tree);

const toggleFile = (path: string) => {
  if (selectedFiles.value.has(path)) {
    selectedFiles.value.delete(path);
  } else {
    selectedFiles.value.add(path);
  }
};

const selectAll = () => {
  const allFiles = fileTreeStore.getAllFiles();
  selectedFiles.value = new Set(allFiles.map(f => f.path));
};

const clearSelection = () => {
  selectedFiles.value.clear();
};

const generateContext = async () => {
  isGenerating.value = true;
  try {
    await contextStore.generateContext(Array.from(selectedFiles.value));
  } finally {
    isGenerating.value = false;
  }
};
</script>

<style scoped>
.file-selector {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.file-tree {
  flex: 1;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 0.5rem;
}

.file-actions {
  display: flex;
  gap: 0.5rem;
  padding: 1rem 0;
}

.file-actions button {
  padding: 0.5rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.file-actions button:hover:not(:disabled) {
  background: #f3f4f6;
}

.file-actions button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
```

## 📊 Примеры метрик и мониторинга

### Метрики производительности

```typescript
// services/metrics.service.ts
export class MetricsService {
    private metrics: Map<string, number> = new Map();
    
    recordMetric(name: string, value: number) {
        this.metrics.set(name, value);
        this.sendToAnalytics(name, value);
    }
    
    recordTiming(name: string, startTime: number) {
        const duration = Date.now() - startTime;
        this.recordMetric(`${name}_duration`, duration);
    }
    
    private sendToAnalytics(name: string, value: number) {
        // Отправка в систему аналитики
        console.log(`Metric: ${name} = ${value}`);
    }
}

// Использование в компонентах
const metrics = new MetricsService();

const generateContext = async () => {
    const startTime = Date.now();
    try {
        const result = await contextStore.generateContext(files);
        metrics.recordTiming('context_generation', startTime);
        metrics.recordMetric('context_size', result.content.length);
        return result;
    } catch (error) {
        metrics.recordMetric('context_generation_errors', 1);
        throw error;
    }
};
```

## 🔄 Примеры автоматизации

### GitHub Action для автоматического анализа

```yaml
# .github/workflows/code-analysis.yml
name: Code Analysis

on:
  pull_request:
    branches: [ main, develop ]

jobs:
  analyze:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
    
    - name: Install dependencies
      run: |
        cd backend && go mod download
        cd ../frontend && npm install
    
    - name: Run tests
      run: |
        cd backend && go test ./...
        cd ../frontend && npm test
    
    - name: Static analysis
      run: |
        cd backend && staticcheck ./...
        cd ../frontend && npm run lint
    
    - name: Generate context and analyze
      run: |
        # Запуск Shotgun Code для анализа
        ./shotgun-code analyze --pr ${{ github.event.pull_request.number }}
```

## 📝 Заключение

Эти примеры демонстрируют основные возможности Shotgun Code и показывают, как интегрировать его в различные рабочие процессы разработки.

Для получения дополнительной информации обратитесь к:
- [API Documentation](./API_DOCUMENTATION.md)
- [Architecture](./Architecture.md)
- [Setup Guide](./SETUP_GUIDE.md)