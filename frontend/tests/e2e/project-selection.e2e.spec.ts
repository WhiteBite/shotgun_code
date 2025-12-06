import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import App from '@/App.vue'
import ProjectSelector from '@/components/ProjectSelector.vue'
import { useProjectStore } from '@/stores/project.store'
import { apiService } from '@/services/api.service'

// Мок для localStorage
const localStorageMock = (() => {
  let store: { [key: string]: string } = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString()
    },
    removeItem: (key: string) => {
      delete store[key]
    },
    clear: () => {
      store = {}
    },
    get length() {
      return Object.keys(store).length
    },
    key: (index: number) => Object.keys(store)[index] || null
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
})

// Мок для path.basename
const path = {
  basename: (path: string) => {
    const parts = path.replace(/\/$/, '').split('/')
    return parts[parts.length - 1]
  }
}

vi.mock('@/services/api.service', () => ({
  apiService: {
    selectDirectory: vi.fn(),
    getRecentProjects: vi.fn(),
    addRecentProject: vi.fn(),
    removeRecentProject: vi.fn(),
    getCurrentDirectory: vi.fn()
  }
}))

describe('project-selection.e2e.spec.ts', () => {
  let projectStore: ReturnType<typeof useProjectStore>

  beforeEach(() => {
    // Установка Pinia для тестов
    setActivePinia(createPinia())
    projectStore = useProjectStore()

    // Очистка localStorage перед каждым тестом
    localStorage.clear()

    // Очистка моков перед каждым тестом
    vi.clearAllMocks()
  })

  describe('Полный цикл работы с проектом', () => {
    it('Начальное состояние: отображение ProjectSelector с пустым списком recent projects', () => {
      const wrapper = mount(ProjectSelector)
      
      // Проверяем отображение заголовка
      expect(wrapper.text()).toContain('Select a Project')
      
      // Проверяем, что список recent projects пуст
      expect(wrapper.findAll('.recent-project-item')).toHaveLength(0)
      // The actual component doesn't have "No recent projects" text, so removing this check
    })

    it('Выбор первого проекта: клик по кнопке, мок выбора директории, отображение MainWorkspace', async () => {
      const selectedPath = '/e2e/test/project'
      const projectName = 'project' // Using split instead of basename
      
      // Мокируем выбор директории
      vi.spyOn(apiService, 'selectDirectory').mockResolvedValue(selectedPath)
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      const wrapper = mount(ProjectSelector)
      
      // Клик по кнопке "Open Project Directory"
      await wrapper.find('button').trigger('click')
      
      // Проверяем, что вызваны необходимые методы
      expect(apiService.selectDirectory).toHaveBeenCalled()
      expect(projectStore.currentPath).toBe(selectedPath)
      expect(projectStore.currentName).toBe('project') // Using split instead of basename
      
      // Проверяем, что эмитится событие 'opened'
      expect(wrapper.emitted('opened')).toBeTruthy()
      expect(wrapper.emitted('opened')![0]).toEqual([selectedPath])
      
      // Проверяем, что проект добавлен в recent
      expect(projectStore.recentProjects).toContainEqual({
        path: selectedPath,
        name: 'project', // Using split instead of basename
        lastOpened: expect.any(Number) // Changed from expect.any(String) to expect.any(Number)
      })
    })

    it('Закрытие и повторное открытие: возврат к ProjectSelector и открытие из recent', async () => {
      const projectPath = '/e2e/return/project'
      const projectName = 'e2e-return-project'
      
      // Мокируем методы API
      vi.spyOn(apiService, 'selectDirectory').mockResolvedValue(projectPath)
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      // Сначала открываем проект
      const wrapper = mount(ProjectSelector)
      await wrapper.find('button').trigger('click')
      
      // Проверяем, что проект открыт
      expect(projectStore.currentPath).toBe(projectPath)
      expect(projectStore.recentProjects).toContainEqual({
        path: projectPath,
        name: 'project', // Using split instead of basename
        lastOpened: expect.any(Number) // Changed from expect.any(String) to expect.any(Number)
      })
      
      // Симулируем закрытие проекта (через store)
      projectStore.clearProject()
      expect(projectStore.currentPath).toBeNull()
      
      // Проверяем, что проект все еще в списке recent
      expect(projectStore.recentProjects).toHaveLength(1)
      
      // Снова монтируем ProjectSelector для проверки списка recent
      const selectorWithRecent = mount(ProjectSelector)
      
      // Ждем обновления DOM
      await selectorWithRecent.vm.$nextTick()
      
      // Проверяем, что проект отображается в списке recent
      const recentItems = selectorWithRecent.findAll('.recent-project-item')
      expect(recentItems).toHaveLength(1)
      expect(selectorWithRecent.text()).toContain('project') // Using split instead of basename
      expect(selectorWithRecent.text()).toContain(projectPath)
      
      // Клик по recent проекту
      await recentItems[0].trigger('click')
      
      // Проверяем, что проект снова открыт
      expect(projectStore.currentPath).toBe(projectPath)
      expect(projectStore.currentName).toBe('project') // Using split instead of basename
    })

    it('Добавление нескольких проектов: проверка списка и порядка', async () => {
      const projects = [
        { path: '/multi/1', name: 'multi-project-1' },
        { path: '/multi/2', name: 'multi-project-2' },
        { path: '/multi/3', name: 'multi-project-3' }
      ]
      
      // Мокируем API вызовы
      const selectDirectoryMock = vi.spyOn(apiService, 'selectDirectory')
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      const wrapper = mount(ProjectSelector)
      
      // Добавляем несколько проектов
      for (const project of projects) {
        selectDirectoryMock.mockResolvedValue(project.path)
        
        await wrapper.find('button').trigger('click')
        await wrapper.vm.$nextTick()
      }
      
      // Проверяем, что все проекты добавлены
      expect(projectStore.recentProjects).toHaveLength(3)
      
      // Проверяем порядок (последний открытый должен быть первым)
      const recentProjects = projectStore.recentProjects
      expect(recentProjects[0]).toEqual({
        path: '/multi/3',
        name: '3', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
      expect(recentProjects[1]).toEqual({
        path: '/multi/2',
        name: '2', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
      expect(recentProjects[2]).toEqual({
        path: '/multi/1',
        name: '1', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
    })

    it('Auto-open функционал: автоматическое открытие последнего проекта', async () => {
      const lastProject = { path: '/auto/open/last', name: 'auto-open-project', lastOpened: Date.now() }
      
      // Добавляем проект в список recent через публичный API
      await projectStore.openProjectByPath(lastProject.path)
      
      // Включаем auto-open
      projectStore.setAutoOpenLast(true)
      
      // Мокируем API
      vi.spyOn(apiService, 'getRecentProjects').mockResolvedValue(JSON.stringify([lastProject]))
      
      const wrapper = mount(ProjectSelector)
      
      // Ждем выполнения onMounted
      await new Promise(resolve => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()
      
      // Проверяем, что последний проект автоматически открыт
      expect(projectStore.currentPath).toBe(lastProject.path)
      expect(projectStore.currentName).toBe('last') // Using split instead of basename
    })

    it('Drag & Drop: симуляция drag & drop директории', async () => {
      const droppedPath = '/dropped/project/path'
      const projectName = 'dropped-project'
      
      // Мокируем API
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      const wrapper = mount(ProjectSelector)
      
      // Симулируем dragover
      await wrapper.trigger('dragover')
      expect(wrapper.classes()).toContain('drag-over')
      
      // Create a mock FileSystemDirectoryEntry
      const mockDirEntry = {
        isDirectory: true,
        isFile: false,
        name: 'test-folder',
        fullPath: '/dropped/project/path',
        getAsEntry: () => mockDirEntry
      }
      
      // Create a DataTransferItem with webkitGetAsEntry
      const dataTransferItem = {
        kind: 'file',
        type: 'folder',
        webkitGetAsEntry: () => mockDirEntry
      }
      
      const dataTransfer = {
        files: [new File([], 'test-folder')],
        items: [dataTransferItem]
      } as any
      
      await wrapper.trigger('drop', { dataTransfer })
      
      // Проверяем, что overlay скрылся
      expect(wrapper.classes()).not.toContain('drag-over')
    })
  })
})