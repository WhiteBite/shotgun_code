# Исправление Синтаксической Ошибки Go и Стратегия Тестирования Playwright MCP

## Обзор

Данный дизайн решает две критически важные задачи:
1. **Немедленное исправление синтаксической ошибки Go** в `backend/application/ai_service.go`, препятствующей компиляции
2. **Комплексная стратегия тестирования Playwright MCP** для валидации полнофункционального приложения

Решение включает исправление дублированных блоков кода и реализацию автоматизированного end-to-end тестирования через Model Control Protocol (MCP) с интеграцией Playwright.

## Архитектура

### Архитектура Исправления Ошибок

```mermaid
flowchart TD
    A[Ошибка Синтаксиса Go] --> B[Анализ Структуры Кода]
    B --> C[Определение Дублированного Кода]
    C --> D[Удаление Избыточных Блоков]
    D --> E[Проверка Границ Функций]
    E --> F[Тест Компиляции]
    F --> G[Запуск dev.ps1]
```

### Архитектура Тестирования Playwright MCP

```mermaid
flowchart TD
    A[Тест-раннер Playwright] --> B[Уровень Интеграции MCP]
    B --> C[Тесты Панели Автономного Управления]
    B --> D[Тесты Frontend UI]
    B --> E[Тесты Backend API]
    B --> F[Тесты Full-Stack Интеграции]
    
    C --> G[Валидация Выполнения Задач]
    D --> H[Тесты Взаимодействия Компонентов]
    E --> I[Валидация Уровня Сервисов]
    F --> J[Тесты End-to-End Процессов]
    
    G --> K[Агрегация Результатов Тестов]
    H --> K
    I --> K
    J --> K
    
    K --> L[Генерация Отчётов]
    L --> M[Метрики Производительности]
    L --> N[Анализ Покрытия]
```

## Исправление Синтаксической Ошибки Go

### Анализ Проблемы

Ошибка возникает на строке 228 в `ai_service.go`:
```
syntax error: non-declaration statement outside function body
```

**Корневая Причина**: Дублированный блок кода существует вне границ функций, содержащий:
- Объявления переменных
- Условную логику
- Инструкции return

### Стратегия Решения

| Компонент | Действие | Обоснование |
|-----------|--------|--------|
| **Удаление Блока Кода** | Удалить строки 225-239 | Дублированная логика уже реализована в функции |
| **Проверка Границ Функций** | Обеспечить правильные закрывающие скобки | Предотвращение утечки области видимости |
| **Проверка Импортов** | Проверить импорт пакета fmt | Поддержка форматирования ошибок |

### Шаги Реализации

1. **Определить Дублированный Блок**: Строки 225-239 содержат логику, уже реализованную в функции
2. **Удалить Избыточный Код**: Удалить дублированную логику выбора провайдера/модели
3. **Проверить Структуру Функции**: Убедиться, что функция `GetIntelligentService()` правильно закрыта
4. **Протестировать Компиляцию**: Запустить `go mod tidy` и `go build` для проверки исправления

## Подробная Стратегия Тестирования через MCP

Поскольку MCP уже подключён, необходимо создать комплексную систему тестирования, которая полностью проверяет все аспекты взаимодействия с MCP инструментами.

### Архитектура MCP Тестирования

```mermaid
flowchart TD
    A[MCP Клиент] --> B[Менеджер Инструментов]
    B --> C[Файловая Система MCP]
    B --> D[Git MCP]
    B --> E[Процесс MCP]
    B --> F[Поиск MCP]
    B --> G[Память MCP]
    B --> H[Browser MCP]
    
    C --> I[Операции с Файлами]
    D --> J[Git Операции]
    E --> K[Выполнение Команд]
    F --> L[Поиск Кода]
    G --> M[Управление Контекстом]
    H --> N[Веб-Интеграция]
    
    I --> O[Playwright Тесты]
    J --> O
    K --> O
    L --> O
    M --> O
    N --> O
    
    O --> P[Результаты Тестирования]
```

### Детальное Тестирование MCP Инструментов

#### 1. Тестирование Filesystem MCP

```mermaid
graph TB
    A[Файловая Система MCP] --> B[Чтение Файлов]
    A --> C[Запись Файлов]
    A --> D[Создание Папок]
    A --> E[Удаление Файлов]
    A --> F[Навигация по Папкам]
    A --> G[Проверка Прав]
    
    B --> H[Тест Чтения Go Файлов]
    B --> I[Тест Чтения Vue Файлов]
    B --> J[Тест Чтения Конфигураций]
    
    C --> K[Тест Изменения Кода]
    C --> L[Тест Создания Новых Файлов]
    
    G --> M[Тест Безопасности Доступа]
    G --> N[Тест Ограничений Workspace]
```

**Ключевые тестовые сценарии:**

| Сценарий | Описание | Ожидаемый Результат |
|-----------|-------------|-------------------|
| **Чтение Go файлов** | Чтение ai_service.go через MCP | Полное содержимое файла |
| **Запись конфигураций** | Изменение task_protocol.yaml | Успешное сохранение |
| **Навигация по проекту** | Обход структуры backend/ | Полный список файлов |
| **Безопасность** | Попытка доступа за пределы workspace | Отклонение доступа |

#### 2. Тестирование Git MCP

```mermaid
sequenceDiagram
    participant T as Тест
    participant G as Git MCP
    participant R as Репозиторий
    
    T->>G: git status
    G->>R: Проверка состояния
    R-->>G: Список изменений
    G-->>T: Статус репозитория
    
    T->>G: git add .
    G->>R: Добавление файлов
    R-->>G: Подтверждение
    G-->>T: Успешное добавление
    
    T->>G: git commit -m "Тестовый коммит"
    G->>R: Создание коммита
    R-->>G: Коммит создан
    G-->>T: Hash коммита
    
    T->>G: git log --oneline -5
    G->>R: Получение истории
    R-->>G: Последние коммиты
**Специальные Git тесты:**

| Команда | Цель Теста | Критерии Успеха |
|---------|-------------|----------------|
| `git status` | Проверка состояния репозитория | Корректный вывод статуса |
| `git diff` | Просмотр изменений | Отображение diff между версиями |
| `git branch` | Управление ветками | Список всех веток |
| `git log` | Просмотр истории | Полная история коммитов |

#### 3. Тестирование Process MCP

```mermaid
graph TD
    A[Процесс MCP] --> B[Выполнение Go Команд]
    A --> C[Выполнение NPM Команд]
    A --> D[Запуск Wails]
    A --> E[Тестирование Компиляции]
    
    B --> F[go mod tidy]
    B --> G[go build]
    B --> H[go test]
    
    C --> I[npm install]
    C --> J[npm run build]
    C --> K[npm run test]
    
    D --> L[wails dev]
    D --> M[wails build]
    
    E --> N[Компиляция Backend]
    E --> O[Компиляция Frontend]
```

**Критические команды для тестирования:**

```bash
# Проверка синтаксиса Go
go mod tidy
go build ./backend/...
go vet ./backend/...

# Проверка Frontend
npm run lint
npm run type-check
npm run build

# Запуск разработки
wails dev

# Тестирование
go test -v ./backend/...
npm run test:unit
npm run test:e2e
```

#### 4. Тестирование Search MCP

```mermaid
flowchart LR
    A[Поиск MCP] --> B[Поиск по Коду]
    A --> C[Поиск по Файлам]
    A --> D[Поиск по Именам]
    A --> E[Регулярные Выражения]
    
    B --> F[Поиск AIService]
    B --> G[Поиск Методов]
    B --> H[Поиск Импортов]
    
    C --> I[Поиск *.go файлов]
    C --> J[Поиск *.vue файлов]
    C --> K[Поиск *.yaml файлов]
```

**Примеры поисковых запросов:**

| Тип Поиска | Пример Запроса | Ожидаемые Результаты |
|-------------|----------------|--------------------|
| **Классы/Структуры** | `AIService` | Файл ai_service.go |
| **Методы** | `GetIntelligentService` | Определение метода |
| **Конфигурации** | `task_protocol` | Файлы .yaml |
| **Импорты** | `import.*wails` | Все Wails импорты |

#### 5. Тестирование Memory MCP

```mermaid
sequenceDiagram
    participant T as Тест
    participant M as Memory MCP
    participant S as Стораж Контекста
    participant A as ИИ Сервис
    
    T->>M: Сохранить контекст
    M->>S: Запись в память
    S-->>M: Подтверждение
    M-->>T: Контекст сохранён
    
    T->>M: Получить контекст
    M->>S: Чтение из памяти
    S-->>M: Данные контекста
    M-->>T: Контекст получен
    
    T->>M: Обновить контекст
    M->>A: Обработка через ИИ
    A-->>M: Обновлённые данные
    M->>S: Сохранение обновлений
    S-->>M: Подтверждение
    M-->>T: Контекст обновлён
```

**Операции с контекстом:**

| Операция | Параметры | Проверка |
|----------|------------|----------|
| **Сохранение** | Ключ, данные, TTL | Успешное сохранение |
| **Получение** | Ключ | Корректные данные |
| **Обновление** | Ключ, новые данные | Синхронизация |
| **Удаление** | Ключ | Очистка памяти |

#### 6. Тестирование Browser MCP

```mermaid
graph LR
    A[Browser MCP] --> B[Веб-Запросы]
    A --> C[Парсинг HTML]
    A --> D[Интеграция API]
    A --> E[Документация]
    
    B --> F[GET Запросы]
    B --> G[POST Запросы]
    B --> H[Обработка Ответов]
    
    C --> I[Извлечение Текста]
    C --> J[Поиск Элементов]
    C --> K[Обработка Форм]
```

**Сценарии веб-интеграции:**

| Задача | URL | Ожидаемый Результат |
|-------|-----|--------------------|
| **Документация Go** | https://golang.org/doc | Получение справки |
| **Документация Vue** | https://vuejs.org/guide | Получение гайдов |
| **GitHub Issues** | https://github.com/user/repo/issues | Список проблем |
| **Новости технологий** | https://news.ycombinator.com | Последние обновления |

### Комплексные Сценарии MCP Тестирования

#### Интеграционные Тесты с Несколькими MCP Инструментами

```mermaid
sequenceDiagram
    participant P as Playwright
    participant FS as Filesystem MCP
    participant Git as Git MCP
    participant Proc as Process MCP
    participant Mem as Memory MCP
    
    P->>FS: Чтение ai_service.go
    FS-->>P: Код с ошибкой
    
    P->>FS: Исправление кода
    FS-->>P: Файл обновлён
    
    P->>Proc: go build ./backend/...
    Proc-->>P: Компиляция успешна
    
    P->>Git: git add ai_service.go
    Git-->>P: Файл добавлен
    
    P->>Git: git commit -m "Исправлена синтаксическая ошибка"
    Git-->>P: Коммит создан
    
    P->>Mem: Сохранить контекст исправления
    Mem-->>P: Контекст сохранён
```

#### Автономные Тестовые Потоки

```mermaid
flowchart TD
    A[Начало Автономного Теста] --> B[Запуск Приложения]
    B --> C[Проверка MCP Подключения]
    C --> D[Отправка Тестовой Задачи]
    D --> E[Мониторинг Выполнения]
    E --> F[Проверка Результатов]
    F --> G[Валидация Изменений]
    G --> H[Отчёт о Тестировании]
    
    E --> I{Ошибка?}
    I --> |Yes| J[Обработка Ошибки]
    I --> |No| F
    J --> K[Повторная Попытка]
    K --> E
```

### Test Suite Architecture

#### 1. Frontend Component Testing

```mermaid
graph LR
    A[UI Components] --> B[Autonomous Control Panel]
    A --> C[File Panel]
    A --> D[Context Panel]
    A --> E[Reports Panel]
    
    B --> F[Task Submission]
    B --> G[Status Monitoring]
    B --> H[Results Display]
    
    C --> I[File Selection]
    C --> J[Project Loading]
    
    D --> K[Context Building]
    D --> L[Context Validation]
    
    E --> M[Report Generation]
    E --> N[Export Functionality]
```

#### 2. Backend Service Testing

```mermaid
graph TD
    A[Backend Services] --> B[AI Service]
    A --> C[Task Protocol Service]
    A --> D[Autonomous Service]
    
    B --> E[Provider Management]
    B --> F[Model Selection]
    B --> G[Intelligent Service]
    
    C --> H[Protocol Validation]
    C --> I[Stage Execution]
    C --> J[Correction Guidance]
    
    D --> K[Task Planning]
    D --> L[Execution Engine]
    D --> M[Monitoring]
```

### Test Scenarios

#### Core Functionality Tests

| Test Category | Test Cases | Expected Behavior |
|--------------|------------|------------------|
| **Autonomous Mode** | Task submission, execution monitoring, result validation | Complete task lifecycle without intervention |
| **AI Integration** | Provider switching, model selection, response streaming | Proper AI service integration |
| **Context Management** | Context building, validation, token management | Efficient context handling |
| **File Operations** | Project loading, file selection, modification tracking | Reliable file system interaction |

#### MCP-Specific Tests

| MCP Tool | Test Focus | Validation Criteria |
|----------|------------|-------------------|
| **filesystem** | File operations, directory traversal | Secure file access within workspace |
| **git** | Repository operations, commit tracking | Proper version control integration |
| **process** | Command execution, output capture | Safe process management |
| **search** | Code search, symbol lookup | Accurate search results |
| **memory** | Knowledge retrieval, context caching | Efficient memory utilization |
| **browsermcp** | Web content fetching, analysis | External resource integration |

### Реализация Тестов

#### Улучшенная Конфигурация Playwright

```typescript
// Улучшенная конфигурация для MCP тестирования
export default defineConfig({
  testDir: './tests-e2e',
  timeout: 60000, // Расширено для автономных операций
  fullyParallel: false,
  retries: 3,
  workers: 1,
  
  use: {
    baseURL: 'http://localhost:34115',
    trace: 'on',
    screenshot: 'on',
    video: 'on',
    // MCP-специфичные заголовки
    extraHTTPHeaders: {
      'X-MCP-Client': 'playwright-test'
    }
  },
  
  projects: [
    {
      name: 'mcp-integration',
      testDir: './tests-e2e/mcp',
      use: { ...devices['Desktop Chrome'] }
    },
    {
      name: 'autonomous-workflow',
      testDir: './tests-e2e/autonomous',
      use: { ...devices['Desktop Chrome'] }
    }
  ]
});
```

#### Структура Тестовых Случаев

```mermaid
sequenceDiagram
    participant T as Тест-раннер
    participant UI as Frontend UI
    participant MCP as MCP Уровень
    participant BE as Backend
    participant AI as AI Сервис
    
    T->>UI: Перейти к Панели Автономного Управления
    UI->>T: Панель загружена
    
    T->>UI: Отправить тестовую задачу
    UI->>MCP: Передать запрос задачи
    MCP->>BE: Обработать автономную задачу
    BE->>AI: Выполнить AI операции
    AI-->>BE: Вернуть результаты
    BE-->>MCP: Завершение задачи
    MCP-->>UI: Обновить статус
    UI-->>T: Отобразить результаты
    
    T->>T: Проверить выполнение задачи
    T->>T: Проверить обновления UI
    T->>T: Проверить логи backend
```

### Тестирование Производительности

#### Сбор Метрик

| Тип Метрики | Измерение | Целевой Порог |
|-------------|-------------|------------------|
| **Время Отклика** | От инициации до завершения задачи | < 30 секунд |
| **Использование Памяти** | Пиковое потребление при выполнении | < 100MB frontend |
| **Загрузка CPU** | Нагрузка обработки backend | < 80% постоянно |
| **Сетевая Задержка** | Коммуникация frontend-backend | < 100ms |

#### Сценарии Нагрузочного Тестирования

```mermaid
graph TD
    A[Сценарии Нагрузочного Тестирования] --> B[Однопользовательский Поток]
    A --> C[Параллельное Выполнение Задач]
    A --> D[Обработка Большого Контекста]
    A --> E[Расширенное Тестирование Сессий]
    
    B --> F[Полный цикл разработки]
    C --> G[Несколько автономных задач]
    D --> H[Анализ большой кодовой базы]
    E --> I[8-часовое непрерывное использование]
```

## Testing Execution Plan

### Phase 1: Syntax Error Resolution
1. **Fix Go Code**: Remove duplicate code block
2. **Verify Compilation**: Run `go mod tidy` and build
3. **Launch Application**: Execute `dev.ps1`
4. **Basic Smoke Test**: Verify application starts

### Phase 2: MCP Tool Validation
1. **Tool Availability Check**: Verify all required MCP tools
2. **Connection Testing**: Validate MCP communication
3. **Permission Verification**: Test security restrictions
4. **Command Whitelisting**: Verify allowed commands

### Phase 3: Comprehensive Testing
1. **Frontend Component Tests**: UI interaction validation
2. **Backend Service Tests**: API endpoint verification
3. **Integration Tests**: Full workflow validation
4. **Performance Tests**: Load and stress testing

### Phase 4: Autonomous Workflow Testing
1. **Task Submission**: Various task types
2. **Execution Monitoring**: Progress tracking
3. **Result Validation**: Output verification
4. **Error Handling**: Failure recovery testing

## Error Handling and Recovery

### Test Failure Scenarios

```mermaid
graph TD
    A[Test Failure] --> B{Failure Type}
    
    B --> C[Compilation Error]
    B --> D[Runtime Error]
    B --> E[UI Interaction Failure]
    B --> F[Backend Service Error]
    
    C --> G[Code syntax validation]
    D --> H[Environment validation]
    E --> I[Component state verification]
    F --> J[Service dependency check]
    
    G --> K[Auto-retry with fix]
    H --> K
    I --> K
    J --> K
    
    K --> L[Report generation]
```

### Recovery Strategies

| Error Type | Recovery Action | Escalation Path |
|------------|----------------|-----------------|
| **Syntax Error** | Auto-fix and retry | Manual code review |
| **Timeout** | Retry with extended timeout | Performance analysis |
| **UI Element Not Found** | Wait and retry | Component architecture review |
| **API Failure** | Service restart | Backend diagnostics |

## Quality Assurance

### Test Coverage Requirements

| Component | Coverage Target | Validation Method |
|-----------|----------------|------------------|
| **Frontend Components** | 90% | Component interaction tests |
| **Backend Services** | 85% | Integration tests |
| **API Endpoints** | 95% | End-to-end tests |
| **MCP Integration** | 80% | Tool-specific tests |

### Continuous Integration

```mermaid
graph LR
    A[Code Commit] --> B[Syntax Validation]
    B --> C[Unit Tests]
    C --> D[Integration Tests]
    D --> E[E2E Tests]
    E --> F[Performance Tests]
    F --> G[Report Generation]
    G --> H[Deployment Ready]
```

## Тестирование

### Стратегия Модульного Тестирования
- **Frontend**: Vitest с JSDOM для тестирования компонентов
- **Backend**: Фреймворк Go testing для проверки сервисов
- **Интеграция**: Playwright для межслойного тестирования

### Управление Тестовыми Данными
- **Мок-Данные**: Структурированные тестовые наборы данных
- **Настройка Окружения**: Изолированные тестовые окружения
- **Управление Состоянием**: Очистка состояния между тестами

### Отчётность и Аналитика
- **Результаты Тестов**: HTML-отчёты с детальными метриками
- **Метрики Производительности**: Автоматизированное отслеживание производительности
- **Анализ Покрытия**: Отчёты о покрытии кода
- **Анализ Сбоев**: Автоматизированная категоризация ошибок