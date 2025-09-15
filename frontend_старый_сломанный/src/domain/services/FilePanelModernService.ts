import { ref, type Ref } from 'vue'
import { FileTreeFilteringService } from '@/domain/services/FileTreeFilteringService'
import { UIFormattingService } from '@/domain/services/UIFormattingService'
import { APP_CONFIG } from '@/config/app-config'

export class FilePanelModernService {
  // Services
  private _filteringService: FileTreeFilteringService
  private _formattingService: UIFormattingService
  
  // State management
  private _searchQuery: Ref<string> = ref('')
  private _showHidden: Ref<boolean> = ref(false)
  private _expandedFolders: Ref<Set<string>> = ref(new Set())
  
  constructor() {
    this._filteringService = new FileTreeFilteringService()
    this._formattingService = new UIFormattingService()
  }
  
  // Getters
  get filteringService() {
    return this._filteringService
  }
  
  get formattingService() {
    return this._formattingService
  }
  
  get searchQuery() {
    return this._searchQuery.value
  }
  
  set searchQuery(value: string) {
    this._searchQuery.value = value
  }
  
  get showHidden() {
    return this._showHidden.value
  }
  
  set showHidden(value: boolean) {
    this._showHidden.value = value
  }
  
  get expandedFolders() {
    return this._expandedFolders.value
  }
  
  set expandedFolders(value: Set<string>) {
    this._expandedFolders.value = value
  }
  
  // Methods for panel management
  toggleShowHidden() {
    this._showHidden.value = !this._showHidden.value
  }
  
  handleSearch() {
    // Search is reactive through computed property
  }
  
  clearSearch() {
    this._searchQuery.value = ''
  }
  
  handleFileSelect(
    filePath: string, 
    isSelected: boolean, 
    toggleSelection: (filePath: string) => void,
    _unusedRemove?: (filePath: string) => void
  ) {
    // Since we're using toggle, we just call it once
    toggleSelection(filePath)
  }
  
  handleFolderToggle(folderPath: string, isExpanded: boolean) {
    if (isExpanded) {
      this._expandedFolders.value.add(folderPath)
    } else {
      this._expandedFolders.value.delete(folderPath)
    }
  }
  
  handleFileContextMenu(filePath: string, event: MouseEvent) {
    // Handle context menu
    console.log('Context menu for:', filePath, event)
  }
  
  clearSelection(clearSelectionFn: () => void) {
    clearSelectionFn()
  }
  
  buildContext(
    projectPath: string | undefined, 
    buildContextFromSelection: (projectPath: string) => void
  ) {
    if (projectPath) {
      buildContextFromSelection(projectPath)
    }
  }
  
  formatFileSize(bytes: number): string {
    return this._formattingService.formatFileSize(bytes)
  }
  
  // Use the filtering service instead of local functions
  getFilteredFileTree(fileTree: any[], searchQuery: string) {
    if (!searchQuery || !fileTree) return fileTree || []
    return this._filteringService.filterNodes(fileTree, searchQuery)
  }
  
  getFilteredFileCount(searchQuery: string, totalFileCount: number, filteredFileTree: any[], fileTree: any[]) {
    if (!searchQuery) return totalFileCount
    if (!filteredFileTree) return 0
    return this._filteringService.countFilteredFiles(filteredFileTree)
  }
}