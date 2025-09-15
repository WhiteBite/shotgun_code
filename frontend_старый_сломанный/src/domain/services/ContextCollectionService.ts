import { ref, type Ref } from 'vue'
import type { ContextSummary, ContextSuggestion } from '@/domain/entities/ContextSummary'
import { APP_CONFIG } from '@/config/app-config'

export class ContextCollectionService {
  // State management
  private _activeView: Ref<'explorer' | 'builder'> = ref<'explorer' | 'builder'>('explorer')
  
  // Getters
  get activeView() {
    return this._activeView.value
  }
  
  set activeView(view: 'explorer' | 'builder') {
    this._activeView.value = view
  }
  
  // Methods for view management
  switchToExplorer() {
    this._activeView.value = 'explorer'
  }
  
  switchToBuilder() {
    this._activeView.value = 'builder'
  }
  
  toggleView() {
    this._activeView.value = this._activeView.value === 'explorer' ? 'builder' : 'explorer'
  }
}