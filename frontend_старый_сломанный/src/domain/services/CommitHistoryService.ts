import { ref, type Ref } from 'vue'
import type { Commit } from '@/types/dto'
import { APP_CONFIG } from '@/config/app-config'

export class CommitHistoryService {
  private selectedCommits: Ref<string[]> = ref([])

  constructor(
    private getCommits: () => Commit[],
    private getIsLoading: () => boolean,
    private getIsHistoryVisible: () => boolean,
    private getCurrentProjectPath: () => string | null,
    private hideHistory: () => void,
    private applyCommitSelection: (selectedCommits: string[]) => void,
    private showQuickLook: (params: any) => void,
    private hideQuickLook: () => void
  ) {}

  // Selected commits management
  getSelectedCommits() {
    return this.selectedCommits.value
  }

  setSelectedCommits(commits: string[]) {
    this.selectedCommits.value = commits
  }

  toggleCommitSelection(commitHash: string) {
    const index = this.selectedCommits.value.indexOf(commitHash)
    if (index > -1) {
      this.selectedCommits.value.splice(index, 1)
    } else {
      this.selectedCommits.value.push(commitHash)
    }
  }

  clearSelectedCommits() {
    this.selectedCommits.value = []
  }

  // Commit data accessors
  get commits() {
    return this.getCommits()
  }

  get isLoading() {
    return this.getIsLoading()
  }

  get isHistoryVisible() {
    return this.getIsHistoryVisible()
  }

  // File preview functionality
  showFilePreview(file: string, commitHash: string, event: MouseEvent) {
    const rootDir = this.getCurrentProjectPath()
    if (!rootDir) return
    
    if (event.ctrlKey || event.metaKey) {
      this.showQuickLook({
        rootDir,
        path: file,
        type: "git",
        commitHash,
        position: { x: event.clientX + 10, y: event.clientY + 10 },
        isPinned: false,
      })
    }
  }

  hideFilePreview() {
    this.hideQuickLook()
  }

  openPinnedFile(file: string, commitHash: string, event: MouseEvent) {
    const rootDir = this.getCurrentProjectPath()
    if (!rootDir) return
    
    this.showQuickLook({
      rootDir,
      path: file,
      type: "git",
      commitHash,
      // Центрируем закреплённый предпросмотр для удобства чтения
      position: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
      isPinned: true,
    })
  }

  // Action handlers
  applySelection() {
    this.applyCommitSelection(this.selectedCommits.value)
    this.hideHistory()
    this.clearSelectedCommits()
  }

  cancelSelection() {
    this.hideHistory()
    this.clearSelectedCommits()
  }

  // Computed properties
  get selectedCommitsCount() {
    return this.selectedCommits.value.length
  }

  get canApplySelection() {
    return this.selectedCommits.value.length > 0
  }
}