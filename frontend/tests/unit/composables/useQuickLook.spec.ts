import { useQuickLook } from '@/features/files/composables/useQuickLook'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

// Mock stores
const mockToggleSelect = vi.fn()
const mockToggleExpand = vi.fn()
const mockAddToast = vi.fn()

vi.mock('@/features/files/model/file.store', () => ({
    useFileStore: () => ({
        toggleSelect: mockToggleSelect,
        toggleExpand: mockToggleExpand,
    }),
}))

vi.mock('@/stores/ui.store', () => ({
    useUIStore: () => ({
        addToast: mockAddToast,
    }),
}))

vi.mock('@/composables/useI18n', () => ({
    useI18n: () => ({
        t: (key: string) => key,
    }),
}))

describe('useQuickLook', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('should initialize with closed state', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        expect(quickLook.isVisible.value).toBe(false)
        expect(quickLook.currentPath.value).toBe('')
    })

    it('should open with specified path', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.open('/path/to/file.ts')

        expect(quickLook.isVisible.value).toBe(true)
        expect(quickLook.currentPath.value).toBe('/path/to/file.ts')
    })

    it('should close when same path is opened again', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.open('/path/to/file.ts')
        expect(quickLook.isVisible.value).toBe(true)

        quickLook.open('/path/to/file.ts')
        expect(quickLook.isVisible.value).toBe(false)
    })

    it('should switch to different path', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.open('/path/to/file1.ts')
        quickLook.open('/path/to/file2.ts')

        expect(quickLook.isVisible.value).toBe(true)
        expect(quickLook.currentPath.value).toBe('/path/to/file2.ts')
    })

    it('should close explicitly', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.open('/path/to/file.ts')
        quickLook.close()

        expect(quickLook.isVisible.value).toBe(false)
    })

    it('should add file to context', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.addToContext('/path/to/file.ts')

        expect(mockToggleSelect).toHaveBeenCalledWith('/path/to/file.ts')
        expect(mockAddToast).toHaveBeenCalledWith('files.addToContext', 'success')
    })

    it('should handle spacebar for hovered file', () => {
        const hoveredFile = {
            path: ref<string | null>('/path/to/file.ts'),
            isDir: ref<boolean | null>(false),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        const handled = quickLook.handleSpacebarPreview()

        expect(handled).toBe(true)
        expect(quickLook.isVisible.value).toBe(true)
    })

    it('should expand directory on spacebar', () => {
        const hoveredFile = {
            path: ref<string | null>('/path/to/folder'),
            isDir: ref<boolean | null>(true),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        const handled = quickLook.handleSpacebarPreview()

        expect(handled).toBe(true)
        expect(mockToggleExpand).toHaveBeenCalledWith('/path/to/folder')
    })

    it('should close quicklook on spacebar when nothing hovered', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        quickLook.open('/path/to/file.ts')
        const handled = quickLook.handleSpacebarPreview()

        expect(handled).toBe(true)
        expect(quickLook.isVisible.value).toBe(false)
    })

    it('should return false when nothing to do', () => {
        const hoveredFile = {
            path: ref<string | null>(null),
            isDir: ref<boolean | null>(null),
            setHovered: vi.fn(),
            clearHovered: vi.fn(),
        }
        const quickLook = useQuickLook({ hoveredFile })

        const handled = quickLook.handleSpacebarPreview()

        expect(handled).toBe(false)
    })
})
