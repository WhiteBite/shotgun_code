import { ref, type Ref } from 'vue'
import { createTokenEstimator } from '@/infrastructure/context/token-estimator.service'
import { APP_CONFIG } from '@/config/app-config'

export class FilePanelService {
  // State management
  private _estimator = createTokenEstimator()
  
  // Getters
  get estimator() {
    return this._estimator
  }
  
  // Methods for panel management
  calculateSelectionPercent(selectedFilesCount: number, totalFiles: number): number {
    const total = totalFiles || 1
    return (selectedFilesCount / total) * 100
  }
  
  estimateTokens(selectedFiles: string[]): number {
    // Грубая оценка: суммируем имена путей как прокси пока контент не загружен
    const concat = selectedFiles.join("\n")
    return this._estimator.estimate(concat)
  }
  
  calculateSelectedChars(selectedFiles: string[]): number {
    const concat = selectedFiles.join("\n")
    return concat.length
  }
  
  async updateIgnoreRules(refreshFiles: () => Promise<void>) {
    await refreshFiles()
  }
  
  openIgnoreDrawer(openDrawer: (drawer: string) => void) {
    openDrawer("ignore")
  }
}