# 🛠️ Troubleshooting

## Обзор

Этот документ содержит решения для типовых проблем, которые могут возникнуть при использовании Shotgun Code.

## 🔍 Диагностика проблем

### Проверка состояния системы

```bash
# Проверка версий
go version
node --version
npm --version
wails doctor

# Проверка портов
netstat -an | grep 34115
lsof -i :34115

# Проверка логов
tail -f ~/.shotgun/logs/app.log
tail -f ~/.shotgun/logs/error.log
```

## 🚨 Частые проблемы и решения

### 1. Проблемы с установкой

#### Ошибка: "wails: command not found"

**Симптомы:**
```bash
wails: command not found
```

**Причины:**
- Wails не установлен
- PATH не настроен правильно
- Неправильная версия Go

**Решения:**

1. **Переустановка Wails:**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

2. **Проверка PATH:**
```bash
echo $PATH
# Убедитесь, что $GOPATH/bin в PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

3. **Проверка версии Go:**
```bash
go version
# Должна быть 1.21+
```

#### Ошибка: "go: module not found"

**Симптомы:**
```bash
go: module not found
```

**Решения:**

1. **Обновление модулей:**
```bash
cd backend
go mod tidy
go mod download
```

2. **Очистка кэша:**
```bash
go clean -modcache
go mod download
```

### 2. Проблемы с запуском

#### Ошибка: "port already in use"

**Симптомы:**
```bash
Error: listen tcp :34115: bind: address already in use
```

**Решения:**

1. **Найти и остановить процесс:**
```bash
# Linux/macOS
lsof -ti:34115 | xargs kill -9

# Windows
netstat -ano | findstr :34115
taskkill /PID <PID> /F
```

2. **Изменить порт в конфигурации:**
```yaml
# ~/.shotgun/config.yaml
app:
  port: 34116  # Изменить на свободный порт
```

#### Ошибка: "permission denied"

**Симптомы:**
```bash
Error: permission denied
```

**Решения:**

1. **Проверка прав доступа:**
```bash
ls -la shotgun-code
chmod +x shotgun-code
```

2. **Запуск с правами администратора (если необходимо):**
```bash
sudo ./shotgun-code
```

### 3. Проблемы с API ключами

#### Ошибка: "API key not found"

**Симптомы:**
```bash
Error: OpenAI API key not found
```

**Решения:**

1. **Проверка переменных окружения:**
```bash
echo $OPENAI_API_KEY
echo $GEMINI_API_KEY
```

2. **Создание .env файла:**
```bash
# .env
OPENAI_API_KEY=your_openai_key_here
GEMINI_API_KEY=your_gemini_key_here
```

3. **Проверка конфигурации:**
```yaml
# ~/.shotgun/config.yaml
ai:
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
```

#### Ошибка: "Invalid API key"

**Симптомы:**
```bash
Error: Invalid API key
```

**Решения:**

1. **Проверка ключа:**
```bash
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

2. **Обновление ключа:**
- Получите новый ключ в [OpenAI Dashboard](https://platform.openai.com/api-keys)
- Обновите в .env файле

### 4. Проблемы с LSP серверами

#### Ошибка: "LSP server not found"

**Симптомы:**
```bash
Error: gopls not found
```

**Решения:**

1. **Установка LSP серверов:**
```bash
# Go
go install golang.org/x/tools/gopls@latest

# TypeScript
npm install -g typescript-language-server

# Python
pip install python-language-server[all]
```

2. **Проверка установки:**
```bash
which gopls
which typescript-language-server
which pylsp
```

#### Ошибка: "LSP server crashed"

**Симптомы:**
```bash
Error: LSP server crashed
```

**Решения:**

1. **Перезапуск LSP сервера:**
```bash
# Остановить все LSP процессы
pkill -f gopls
pkill -f typescript-language-server
pkill -f pylsp

# Перезапустить приложение
./shotgun-code
```

2. **Проверка конфигурации LSP:**
```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "typescript.preferences.includePackageJsonAutoImports": "on"
}
```

### 5. Проблемы с производительностью

#### Медленная работа приложения

**Симптомы:**
- Долгая загрузка файлов
- Медленная генерация контекста
- Высокое потребление памяти

**Решения:**

1. **Оптимизация настроек:**
```yaml
# ~/.shotgun/config.yaml
context:
  max_file_size: 524288  # Уменьшить до 512KB
  max_tokens: 2000       # Уменьшить лимит токенов
  compression_enabled: true
```

2. **Очистка кэша:**
```bash
rm -rf ~/.shotgun/cache/*
```

3. **Мониторинг ресурсов:**
```bash
# Мониторинг памяти
htop
# или
top

# Мониторинг диска
df -h
```

#### Ошибка: "out of memory"

**Симптомы:**
```bash
Error: out of memory
```

**Решения:**

1. **Увеличение лимитов памяти:**
```bash
# Linux
ulimit -v 2097152  # 2GB

# macOS
sudo launchctl limit maxfiles 65536 200000
```

2. **Оптимизация настроек:**
```yaml
# ~/.shotgun/config.yaml
app:
  max_memory: 1073741824  # 1GB
context:
  max_files: 100          # Ограничить количество файлов
```

### 6. Проблемы с сетью

#### Ошибка: "connection refused"

**Симптомы:**
```bash
Error: connection refused
```

**Решения:**

1. **Проверка файрвола:**
```bash
# Linux
sudo ufw status
sudo ufw allow 34115

# macOS
sudo pfctl -s rules
```

2. **Проверка сетевых настроек:**
```bash
# Проверка доступности порта
telnet localhost 34115
```

#### Ошибка: "timeout"

**Симптомы:**
```bash
Error: request timeout
```

**Решения:**

1. **Увеличение таймаутов:**
```yaml
# ~/.shotgun/config.yaml
app:
  timeout: 30000  # 30 секунд
```

2. **Проверка сети:**
```bash
ping api.openai.com
curl -I https://api.openai.com
```

### 7. Проблемы с файловой системой

#### Ошибка: "file not found"

**Симптомы:**
```bash
Error: file not found
```

**Решения:**

1. **Проверка путей:**
```bash
ls -la /path/to/file
pwd
```

2. **Проверка прав доступа:**
```bash
ls -la
chmod 644 filename
```

#### Ошибка: "permission denied"

**Симптомы:**
```bash
Error: permission denied
```

**Решения:**

1. **Изменение прав доступа:**
```bash
chmod 755 directory
chmod 644 file
```

2. **Проверка владельца:**
```bash
ls -la
sudo chown user:group file
```

## 🔧 Расширенная диагностика

### Включение отладки

```bash
# Через переменную окружения
export SHOTGUN_DEBUG=true
./shotgun-code

# Или в конфигурации
# ~/.shotgun/config.yaml
app:
  debug: true
```

### Сбор логов

```bash
# Создание директории для логов
mkdir -p ~/.shotgun/logs

# Запуск с логированием
./shotgun-code --log-level=debug --log-file=~/.shotgun/logs/app.log
```

### Профилирование

```bash
# Профилирование CPU
go tool pprof http://localhost:6060/debug/pprof/profile

# Профилирование памяти
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 📊 Мониторинг и метрики

### Проверка здоровья системы

```bash
# Проверка API
curl http://localhost:34115/api/v1/health

# Проверка метрик
curl http://localhost:34115/api/v1/metrics
```

### Мониторинг производительности

```bash
# Мониторинг в реальном времени
watch -n 1 'ps aux | grep shotgun-code'

# Мониторинг сети
iftop -i lo
```

## 🆘 Получение помощи

### Сбор информации для отчета об ошибке

```bash
# Системная информация
uname -a
go version
node --version
npm --version

# Конфигурация
cat ~/.shotgun/config.yaml

# Логи
tail -n 100 ~/.shotgun/logs/app.log
tail -n 100 ~/.shotgun/logs/error.log

# Состояние системы
df -h
free -h
```

### Создание отчета об ошибке

При создании Issue в GitHub включите:

1. **Описание проблемы**
2. **Шаги для воспроизведения**
3. **Ожидаемое поведение**
4. **Фактическое поведение**
5. **Системная информация**
6. **Логи ошибок**
7. **Конфигурация**

### Полезные ссылки

- [GitHub Issues](https://github.com/your-org/shotgun-code/issues)
- [Discord Community](https://discord.gg/shotgun-code)
- [Documentation](./README.md)
- [API Reference](./API_DOCUMENTATION.md)

## 🔄 Восстановление после сбоя

### Полная переустановка

```bash
# Остановка приложения
pkill -f shotgun-code

# Удаление конфигурации
rm -rf ~/.shotgun

# Переустановка зависимостей
cd backend && go mod download
cd ../frontend && npm install

# Пересборка
cd backend && wails build

# Запуск
./shotgun-code
```

### Восстановление из резервной копии

```bash
# Создание резервной копии
cp -r ~/.shotgun ~/.shotgun.backup

# Восстановление
cp -r ~/.shotgun.backup ~/.shotgun
```

## 📝 Заключение

Если проблема не решена с помощью этого руководства:

1. Проверьте [GitHub Issues](https://github.com/your-org/shotgun-code/issues)
2. Создайте новый Issue с подробным описанием
3. Обратитесь в Discord сообщество
4. Проверьте обновления проекта

Помните: всегда делайте резервные копии важных данных перед выполнением диагностических операций.
