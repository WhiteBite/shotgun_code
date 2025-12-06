# Структура UI/UX Shotgun Code

## Общая архитектура

Проект Shotgun Code представляет собой Wails + Go + Vue.js приложение для построения контекста проекта с возможностью AI анализа. Архитектура следует принципам Clean Architecture с разделением на Domain, Application и Infrastructure слои.

## Структура UI

### Основные панели:
1. **Left Sidebar** - File Explorer с выбором файлов
2. **Center Workspace** - Context Building (построение контекста)
3. **Right Sidebar** - Tools & AI Chat (инструменты и чат)
4. **Action Bar** - кнопки действий (Build, Export, Generate)

### Функциональность Context Building:

#### Параметры построения контекста:
1. **Max Tokens** (работает ✓)
   - Максимальное количество токенов в контексте
   - По умолчанию: 100K

2. **Strip Comments** (работает ✓)
   - Удаляет комментарии из кода
   - Поддерживаемые языки: Go, JS, TS, Java, C, C++, C#, Python, Shell, HTML, XML
   - Реализация: `backend/infrastructure/textutils/comment_stripper.go`

3. **Include Tests** (будет работать после обновления)
   - Включает/исключает тестовые файлы
   - Паттерны: *_test.go, *.test.ts, *.spec.ts, test_*.py
   - Требует обновления: `backend/domain/models.go` и `backend/application/context_builder.go`

4. **Split Strategy** (будет работать после обновления)
   - semantic: разбиение по файлам (рекомендуется)
   - fixed: фиксированные блоки токенов
   - adaptive: умное разбиение с учетом структуры
   - Реализация: `backend/infrastructure/textutils/context_splitter.go`
   - Требует обновления: `backend/domain/models.go` и `backend/application/context_builder.go`

### AI Chat:
- Расположение: Right Sidebar, вкладка "AI Chat"
- Статус: Coming Soon (не реализовано)
- Планируемые функции:
  - Real-time AI conversation
  - Context-aware responses
  - Code generation and explanation
  - Streaming responses
  - Chat history

### Удаленные дубликаты:
- AI Chat из Center Workspace (был дубликат)

## Технический стек:
- Frontend: Vue.js 3 + TypeScript + Tailwind CSS
- Backend: Go + Wails
- Архитектура: Clean Architecture (Domain, Application, Infrastructure)

## Следующие шаги:
1. Обновить backend модели для поддержки всех параметров
2. Реализовать фильтрацию тестовых файлов
3. Интегрировать context splitter с параметрами построения
4. Реализовать AI Chat функциональность
5. Добавить unit тесты для новых функций