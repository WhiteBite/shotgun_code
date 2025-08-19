# 🚀 Setup Guide

## Требования

### Системные требования
- **ОС**: Windows 10/11, macOS 10.15+, Ubuntu 18.04+
- **RAM**: Минимум 4GB, рекомендуется 8GB+
- **Диск**: 2GB свободного места
- **Сеть**: Интернет для загрузки зависимостей

### Необходимое ПО
- **Go**: версия 1.21+ ([скачать](https://golang.org/dl/))
- **Node.js**: версия 18+ ([скачать](https://nodejs.org/))
- **Git**: любая версия ([скачать](https://git-scm.com/))

## 📦 Установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-org/shotgun-code.git
cd shotgun-code
```

### 2. Установка зависимостей

#### Backend (Go)
```bash
cd backend
go mod download
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### Frontend (Node.js)
```bash
cd frontend
npm install
```

### 3. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
# ИИ провайдеры
OPENAI_API_KEY=your_openai_key_here
GEMINI_API_KEY=your_gemini_key_here

# Настройки приложения
APP_PORT=34115
APP_ENV=development

# База данных (опционально)
DB_PATH=./data/shotgun.db
```

### 4. Первоначальная настройка

```bash
# Сборка проекта
cd backend
wails build

# Или для разработки
wails dev
```

## 🔧 Конфигурация

### Настройки приложения

Файл конфигурации: `~/.shotgun/config.yaml`

```yaml
# Основные настройки
app:
  port: 34115
  host: localhost
  debug: false

# ИИ провайдеры
ai:
  default_provider: openai
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
      models:
        - gpt-4
        - gpt-3.5-turbo
      rate_limit:
        requests_per_minute: 60
        tokens_per_minute: 90000
    
    gemini:
      api_key: ${GEMINI_API_KEY}
      models:
        - gemini-pro
      rate_limit:
        requests_per_minute: 30
        tokens_per_minute: 60000

# Настройки контекста
context:
  max_file_size: 1048576  # 1MB
  max_tokens: 4000
  compression_enabled: true
  excluded_patterns:
    - node_modules
    - dist
    - .git
    - *.log

# Настройки экспорта
export:
  default_format: markdown
  include_metadata: true
  include_file_structure: true
```

### Настройка LSP серверов

#### Go
```bash
go install golang.org/x/tools/gopls@latest
```

#### TypeScript/JavaScript
```bash
npm install -g typescript-language-server
```

#### Python
```bash
pip install python-language-server[all]
```

## 🏃‍♂️ Запуск

### Режим разработки

```bash
# Backend
cd backend
wails dev

# Frontend (в отдельном терминале)
cd frontend
npm run dev
```

### Продакшн сборка

```bash
cd backend
wails build -o shotgun-code
```

### Запуск приложения

```bash
./shotgun-code
```

Приложение будет доступно по адресу: `http://localhost:34115`

## 🔍 Проверка установки

### 1. Проверка зависимостей

```bash
# Проверка Go
go version

# Проверка Node.js
node --version
npm --version

# Проверка Wails
wails doctor
```

### 2. Тестовый запуск

```bash
# Запуск тестов
cd backend
go test ./...

cd frontend
npm test
```

### 3. Проверка API

```bash
# Проверка доступности API
curl http://localhost:34115/api/v1/health
```

Ожидаемый ответ:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 🛠️ Устранение неполадок

### Частые проблемы

#### 1. Ошибка "wails: command not found"
```bash
# Решение: переустановка Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 2. Ошибка "port already in use"
```bash
# Измените порт в конфигурации
# Или найдите и остановите процесс
lsof -ti:34115 | xargs kill -9
```

#### 3. Ошибка "API key not found"
```bash
# Проверьте переменные окружения
echo $OPENAI_API_KEY
# Или добавьте в .env файл
```

#### 4. Проблемы с LSP
```bash
# Переустановите LSP серверы
go install golang.org/x/tools/gopls@latest
npm install -g typescript-language-server
```

### Логи и отладка

#### Включение отладки
```bash
# В конфигурации
debug: true

# Или через переменную окружения
export SHOTGUN_DEBUG=true
```

#### Просмотр логов
```bash
# Логи приложения
tail -f ~/.shotgun/logs/app.log

# Логи ошибок
tail -f ~/.shotgun/logs/error.log
```

## 📚 Следующие шаги

1. **Изучите документацию**:
   - [API Documentation](./API_DOCUMENTATION.md)
   - [Architecture](./Architecture.md)
   - [Examples](./EXAMPLES.md)

2. **Настройте IDE**:
   - Установите расширения для Go, TypeScript, Vue
   - Настройте LSP серверы

3. **Изучите примеры**:
   - Запустите демо проекты
   - Попробуйте различные сценарии использования

4. **Присоединитесь к сообществу**:
   - GitHub Issues для багов
   - Discussions для вопросов
   - Discord для общения

## 🆘 Поддержка

Если у вас возникли проблемы:

1. Проверьте [Troubleshooting](./TROUBLESHOOTING.md)
2. Поищите в [Issues](https://github.com/your-org/shotgun-code/issues)
3. Создайте новый Issue с подробным описанием проблемы
4. Обратитесь в Discord канал поддержки

## 📝 Лицензия

Проект распространяется под лицензией MIT. См. [LICENSE](../LICENSE) для подробностей.