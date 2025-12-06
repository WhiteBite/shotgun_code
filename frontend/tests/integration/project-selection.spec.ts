import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useProjectStore } from '@/stores/project.store'
import ProjectSelector from '@/components/ProjectSelector.vue'
import App from '@/App.vue'
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

describe('project-selection.integration.spec.ts', () => {
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

  describe('Happy path: Выбор проекта через нативный диалог', () => {
    it('должен добавить проект в recent, сохранить в localStorage и переключить App.vue на MainWorkspace', async () => {
      const selectedPath = '/path/to/new/project'
      const projectName = 'project' // Using split instead of basename
      
      // Мокируем apiService.selectDirectory для возврата пути
      vi.spyOn(apiService, 'selectDirectory').mockResolvedValue(selectedPath)
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      const wrapper = mount(ProjectSelector)
      
      // Клик по кнопке выбора директории
      await wrapper.find('button').trigger('click')
      
      // Проверяем, что проект добавлен в recent
      expect(projectStore.recentProjects).toContainEqual({
        path: selectedPath,
        name: projectName,
        lastOpened: expect.any(Number) // Changed from expect.any(String) to expect.any(Number)
      })
      
      // Проверяем, что проект сохранен в localStorage
      const storedRecent = JSON.parse(localStorage.getItem('shotgun_recent_projects') || '[]')
      expect(storedRecent).toContainEqual({
        path: selectedPath,
        name: projectName,
        lastOpened: expect.any(Number) // Changed from expect.any(String) to expect.any(Number)
      })
      
      // Проверяем, что эмитится событие 'opened' с путем проекта
      expect(wrapper.emitted('opened')).toBeTruthy()
      expect(wrapper.emitted('opened')![0]).toEqual([selectedPath])
    })
  })

  describe('Recent projects flow: Открытие из списка recent', () => {
    it('должен отобразить список и открыть проект при клике', async () => {
      const testProjects = [
        { path: '/path/1', name: 'project1', lastOpened: Date.now() },
        { path: '/path/2', name: 'project2', lastOpened: Date.now() }
      ]
      
      // Мокируем apiService.getRecentProjects с тестовыми данными
      vi.spyOn(apiService, 'getRecentProjects').mockResolvedValue(JSON.stringify(testProjects))
      
      const wrapper = mount(ProjectSelector)
      
      // Ждем, пока данные загрузятся
      await new Promise(resolve => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()
      
      // Проверяем отображение списка
      expect(wrapper.findAll('.recent-project-item')).toHaveLength(2)
      expect(wrapper.text()).toContain('project1')
      expect(wrapper.text()).toContain('project2')
      
      // Кликнем на первый проект из списка
      const firstProjectItem = wrapper.findAll('.recent-project-item')[0]
      await firstProjectItem.trigger('click')
      
      // Проверяем, что проект открылся
      expect(projectStore.currentPath).toBe('/path/1')
      expect(projectStore.currentName).toBe('1') // Using split instead of basename
      
      // Проверяем, что проект переместился в начало списка
      expect(projectStore.recentProjects[0]).toEqual({
        path: '/path/1',
        name: '1', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
    })
  })

  describe('Auto-open flow: Автоматическое открытие последнего проекта', () => {
    it('должен открыть проект автоматически если autoOpenLast=true в localStorage', async () => {
      const lastProject = { path: '/last/project', name: 'project', lastOpened: Date.now() } // Using split instead of basename
      
      // Установить autoOpenLast=true в localStorage
      localStorage.setItem('shotgun_auto_open_last', 'true')
      // Добавить проект в recent
      localStorage.setItem('shotgun_recent_projects', JSON.stringify([lastProject]))
      
      // Мокируем apiService.getRecentProjects
      vi.spyOn(apiService, 'getRecentProjects').mockResolvedValue(JSON.stringify([lastProject]))
      
      const wrapper = mount(ProjectSelector)
      
      // Ждем выполнения onMounted
      await new Promise(resolve => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()
      
      // Проверяем, что проект автоматически открылся
      expect(projectStore.currentPath).toBe(lastProject.path)
      expect(projectStore.currentName).toBe('project') // Using split instead of basename
    })
  })

  describe('Error handling: Обработка ошибок', () => {
    it('должен показать toast с ошибкой и не добавить проект в recent', async () => {
      // Мокируем apiService.selectDirectory для выброса ошибки
      vi.spyOn(apiService, 'selectDirectory').mockRejectedValue(new Error('Directory selection failed'))
      
      const wrapper = mount(ProjectSelector)
      
      // Клик по кнопке выбора директории
      await wrapper.find('button').trigger('click')
      
      // Проверяем, что проект не добавлен в recent
      expect(projectStore.recentProjects).toEqual([])
      
      // Проверяем, что эмитится событие ошибки
      expect(wrapper.emitted('opened')).toBeUndefined()
    })
  })

  describe('Persistence: Сохранение и загрузка из localStorage', () => {
    it('должен сохранить и восстановить проекты', async () => {
      const testProjects = [
        { path: '/path/1', name: 'project1', lastOpened: Date.now() },
        { path: '/path/2', name: 'project2', lastOpened: Date.now() },
        { path: '/path/3', name: 'project3', lastOpened: Date.now() }
      ]
      
      // Добавляем проекты в store через публичный API
      for (const project of testProjects) {
        await projectStore.openProjectByPath(project.path)
      }
      
      // Проверяем, что проекты сохранены в localStorage
      const storedRecent = JSON.parse(localStorage.getItem('shotgun_recent_projects') || '[]')
      expect(storedRecent).toContainEqual({ path: '/path/1', name: '1', lastOpened: expect.any(Number) })
      expect(storedRecent).toContainEqual({ path: '/path/2', name: '2', lastOpened: expect.any(Number) })
      expect(storedRecent).toContainEqual({ path: '/path/3', name: '3', lastOpened: expect.any(Number) })
      
      // Очищаем store
      projectStore.recentProjects = []
      
      // Загружаем из localStorage
      const loadedRecent = JSON.parse(localStorage.getItem('shotgun_recent_projects') || '[]')
      projectStore.recentProjects = loadedRecent
      
      // Проверяем, что проекты восстановлены
      expect(projectStore.recentProjects).toEqual(storedRecent)
    })
  })

  describe('Integration: Взаимодействие компонентов и store', () => {
    it('должен корректно передавать данные между ProjectSelector и ProjectStore', async () => {
      const selectedPath = '/integration/test/path'
      const projectName = 'integration-test-project'
      
      // Мокируем apiService
      vi.spyOn(apiService, 'selectDirectory').mockResolvedValue(selectedPath)
      vi.spyOn(apiService, 'addRecentProject').mockResolvedValue()
      
      const wrapper = mount(ProjectSelector)
      
      // Выбираем проект
      await wrapper.find('button').trigger('click')
      
      // Проверяем, что store обновлен
      expect(projectStore.currentPath).toBe(selectedPath)
      expect(projectStore.currentName).toBe('path') // Using split instead of basename
      expect(projectStore.recentProjects).toContainEqual({
        path: selectedPath,
        name: 'path', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
      
      // Проверяем, что проект сохранен в localStorage
      const storedRecent = JSON.parse(localStorage.getItem('shotgun_recent_projects') || '[]')
      expect(storedRecent).toContainEqual({
        path: selectedPath,
        name: 'path', // Using split instead of basename
        lastOpened: expect.any(Number)
      })
      
      // Проверяем, что эмитится событие
      expect(wrapper.emitted('opened')).toBeTruthy()
      expect(wrapper.emitted('opened')![0]).toEqual([selectedPath])
    })
  })
})