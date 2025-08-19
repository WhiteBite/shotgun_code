#!/usr/bin/env python3
"""
Скрипт для тестирования и отладки проблем с интеллектуальным анализом контекста
"""

import json
import sys
import os

def main():
    print("=== ТЕСТИРОВАНИЕ ИНТЕЛЛЕКТУАЛЬНОГО АНАЛИЗА КОНТЕКСТА ===")
    print()
    
    print("1. Проверка структуры проекта:")
    print(f"   - Текущая директория: {os.getcwd()}")
    print(f"   - Backend существует: {os.path.exists('backend')}")
    print(f"   - Frontend существует: {os.path.exists('frontend')}")
    print(f"   - wails.json существует: {os.path.exists('backend/wails.json')}")
    print()
    
    print("2. Проверка ключевых файлов:")
    key_files = [
        'backend/app.go',
        'backend/application/context_analysis_service.go',
        'frontend/src/components/workspace/IntelligentContextPanel.vue',
        'frontend/src/services/api.service.ts',
        'frontend/wailsjs/go/main/App.js'
    ]
    
    for file_path in key_files:
        exists = os.path.exists(file_path)
        print(f"   - {file_path}: {'✅' if exists else '❌'}")
    
    print()
    print("3. Возможные проблемы и решения:")
    print("   a) AI провайдер не настроен:")
    print("      - Перейдите в настройки приложения")
    print("      - Выберите AI провайдера (OpenAI, Gemini, etc.)")
    print("      - Введите API ключ")
    print()
    print("   b) Нет файлов для анализа:")
    print("      - Откройте проект в приложении")
    print("      - Убедитесь, что файлы загружены в дерево файлов")
    print()
    print("   c) Backend не запущен:")
    print("      - Запустите 'wails dev' из папки backend")
    print("      - Проверьте консоль на ошибки")
    print()
    print("   d) Проблемы с Wails bindings:")
    print("      - Выполните 'wails generate module' в папке backend")
    print("      - Перезапустите приложение")
    print()
    
    print("4. Отладочные шаги:")
    print("   - Откройте DevTools в приложении (F12)")
    print("   - Перейдите на вкладку Console")
    print("   - Попробуйте выполнить анализ задачи")
    print("   - Проверьте логи на ошибки")
    print()
    
    print("5. Тестирование:")
    print("   - Нажмите кнопку 'Тест' в панели интеллектуального анализа")
    print("   - Проверьте, что backend отвечает")
    print("   - Если тест проходит, проблема в AI настройках")
    print("   - Если тест не проходит, проблема в backend")
    print()
    
    print("=== КОНЕЦ ТЕСТИРОВАНИЯ ===")
    
    # Спрашиваем пользователя о следующем шаге
    print("\nЧто вы хотите сделать дальше?")
    print("1. Запустить тест backend")
    print("2. Проверить настройки AI")
    print("3. Открыть DevTools и проанализировать ошибки")
    print("4. Перезапустить приложение")
    print("5. Остановить")
    
    choice = input("\nВведите номер (1-5): ").strip()
    
    if choice == "1":
        print("Запускаю тест backend...")
        # Здесь можно добавить автоматический тест
    elif choice == "2":
        print("Проверяю настройки AI...")
        # Здесь можно добавить проверку настроек
    elif choice == "3":
        print("Откройте DevTools (F12) и проверьте консоль на ошибки")
    elif choice == "4":
        print("Перезапустите приложение командой 'wails dev' в папке backend")
    elif choice == "5":
        print("Останавливаю...")
    else:
        print("Неверный выбор")

if __name__ == "__main__":
    main()
