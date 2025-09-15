import { ref, type Ref } from 'vue'
import type { ContextSummary } from '@/domain/entities/ContextSummary'
import type { ContextChunk } from '@/infrastructure/context/context-chunking.service'
import { APP_CONFIG } from '@/config/app-config'

export class ContextPanelService {
  // State management
  private _enableSplit: Ref<boolean> = ref(false)
  private _chunkedContext: Ref<ContextChunk[]> = ref([])
  private _activeChunkIndex: Ref<number> = ref(0)
  private _paginatedContextContent: Ref<string> = ref('')
  
  // Getters
  get enableSplit() {
    return this._enableSplit.value
  }
  
  set enableSplit(value: boolean) {
    this._enableSplit.value = value
  }
  
  get chunkedContext() {
    return this._chunkedContext.value
  }
  
  set chunkedContext(value: ContextChunk[]) {
    this._chunkedContext.value = value
  }
  
  get activeChunkIndex() {
    return this._activeChunkIndex.value
  }
  
  set activeChunkIndex(value: number) {
    this._activeChunkIndex.value = value
  }
  
  get paginatedContextContent() {
    return this._paginatedContextContent.value
  }
  
  set paginatedContextContent(value: string) {
    this._paginatedContextContent.value = value
  }
  
  // Methods for panel management
  toggleSplit() {
    this._enableSplit.value = !this._enableSplit.value
  }
  
  setSplit(enabled: boolean) {
    this._enableSplit.value = enabled
  }
  
  setActiveChunkIndex(index: number) {
    this._activeChunkIndex.value = index
  }
  
  setChunkedContext(chunks: ContextChunk[]) {
    this._chunkedContext.value = chunks
  }
  
  setPaginatedContextContent(content: string) {
    this._paginatedContextContent.value = content
  }
  
  clearContent() {
    this._paginatedContextContent.value = ''
    this._chunkedContext.value = []
    this._activeChunkIndex.value = 0
  }
  
  // Methods for context operations
  handleChunkSelect(index: number) {
    this._activeChunkIndex.value = index
  }
  
  handleChunkHover(index: number | null) {
    // Handle chunk hover if needed
    console.log('Chunk hover:', index)
  }
  
  handleCopyChunk(chunk: ContextChunk) {
    // Copy chunk functionality
    console.log('Copy chunk:', chunk)
  }
  
  handleCopyAll(content: string) {
    // Copy all content functionality
    console.log('Copy all content:', content)
  }
}