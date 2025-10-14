/**
 * Централизованный API сервис для взаимодействия с Go-бэкендом
 * Все вызовы к Wails API должны проходить через этот сервис
 * Это обеспечивает единую точку для обработки ошибок, логирования и тестирования
 */

import * as wails from '@/wailsjs/go/main/App'
import type { domain } from '@/wailsjs/go/models'

/**
 * Класс API сервиса с обертками для всех функций Wails
 */
class ApiService {
  // ============================================
  // Управление проектами
  // ============================================

  async getRecentProjects(): Promise<string> {
    try {
      return await wails.GetRecentProjects()
    } catch (error) {
      console.error('[ApiService] Error fetching recent projects:', error)
      throw new Error('Failed to load recent projects.')
    }
  }

  async addRecentProject(path: string, name: string): Promise<void> {
    try {
      return await wails.AddRecentProject(path, name)
    } catch (error) {
      console.error(`[ApiService] Error adding recent project "${name}":`, error)
      throw new Error('Failed to add recent project.')
    }
  }

  async removeRecentProject(path: string): Promise<void> {
    try {
      return await wails.RemoveRecentProject(path)
    } catch (error) {
      console.error(`[ApiService] Error removing recent project "${path}":`, error)
      throw new Error('Failed to remove recent project.')
    }
  }

  async selectDirectory(): Promise<string> {
    try {
      return await wails.SelectDirectory()
    } catch (error) {
      console.error('[ApiService] Error selecting directory:', error)
      throw new Error('Failed to select directory.')
    }
  }

  async getCurrentDirectory(): Promise<string> {
    try {
      return await wails.GetCurrentDirectory()
    } catch (error) {
      console.error('[ApiService] Error getting current directory:', error)
      throw new Error('Failed to get current directory.')
    }
  }

  // ============================================
  // Работа с файлами
  // ============================================

  async listFiles(path: string, includeHidden: boolean = false, sortByName: boolean = true): Promise<domain.FileNode[]> {
    try {
      return await wails.ListFiles(path, includeHidden, sortByName)
    } catch (error) {
      console.error(`[ApiService] Error listing files for path "${path}":`, error)
      throw new Error('Failed to load file tree.')
    }
  }

  async readFileContent(projectPath: string, filePath: string): Promise<string> {
    try {
      return await wails.ReadFileContent(projectPath, filePath)
    } catch (error) {
      console.error(`[ApiService] Error reading file "${filePath}":`, error)
      throw new Error('Failed to read file content.')
    }
  }

  async getFileStats(path: string): Promise<string> {
    try {
      return await wails.GetFileStats(path)
    } catch (error) {
      console.error(`[ApiService] Error getting file stats for "${path}":`, error)
      throw new Error('Failed to get file statistics.')
    }
  }

  // ============================================
  // Контекст
  // ============================================

  async buildContext(projectPath: string, files: string[], task: string): Promise<string> {
    try {
      return await wails.BuildContext(projectPath, files, task)
    } catch (error) {
      console.error('[ApiService] Error building context:', error)
      throw new Error('Failed to build context.')
    }
  }

  async buildContextFromRequest(
    projectPath: string,
    files: string[],
    options: domain.ContextBuildOptions
  ): Promise<domain.ContextSummaryInfo> {
    try {
      return await wails.BuildContextFromRequest(projectPath, files, options)
    } catch (error) {
      console.error('[ApiService] Error building context from request:', error)
      throw new Error('Failed to build context.')
    }
  }

  async getContext(contextId: string): Promise<string> {
    try {
      return await wails.GetContext(contextId)
    } catch (error) {
      console.error(`[ApiService] Error getting context "${contextId}":`, error)
      throw new Error('Failed to get context.')
    }
  }

  async deleteContext(contextId: string): Promise<void> {
    try {
      return await wails.DeleteContext(contextId)
    } catch (error) {
      console.error(`[ApiService] Error deleting context "${contextId}":`, error)
      throw new Error('Failed to delete context.')
    }
  }

  async getProjectContexts(projectPath: string): Promise<string> {
    try {
      return await wails.GetProjectContexts(projectPath)
    } catch (error) {
      console.error('[ApiService] Error getting project contexts:', error)
      throw new Error('Failed to get project contexts.')
    }
  }

  async exportContext(contextId: string): Promise<domain.ExportResult> {
    try {
      return await wails.ExportContext(contextId)
    } catch (error) {
      console.error(`[ApiService] Error exporting context "${contextId}":`, error)
      throw new Error('Failed to export context.')
    }
  }

  async suggestContextFiles(taskDescription: string, files: domain.FileNode[]): Promise<string[]> {
    try {
      return await wails.SuggestContextFiles(taskDescription, files)
    } catch (error) {
      console.error('[ApiService] Error suggesting context files:', error)
      throw new Error('Failed to suggest context files.')
    }
  }

  // ============================================
  // AI и генерация кода
  // ============================================

  async generateCode(context: string, task: string): Promise<string> {
    try {
      return await wails.GenerateCode(context, task)
    } catch (error) {
      console.error('[ApiService] Error generating code:', error)
      throw new Error('Failed to generate code.')
    }
  }

  async generateIntelligentCode(context: string, task: string, options: string): Promise<string> {
    try {
      return await wails.GenerateIntelligentCode(context, task, options)
    } catch (error) {
      console.error('[ApiService] Error generating intelligent code:', error)
      throw new Error('Failed to generate code.')
    }
  }

  async listAvailableModels(): Promise<string[]> {
    try {
      return await wails.ListAvailableModels()
    } catch (error) {
      console.error('[ApiService] Error listing available models:', error)
      throw new Error('Failed to list AI models.')
    }
  }

  async getProviderInfo(): Promise<string> {
    try {
      return await wails.GetProviderInfo()
    } catch (error) {
      console.error('[ApiService] Error getting provider info:', error)
      throw new Error('Failed to get provider information.')
    }
  }

  // ============================================
  // Анализ
  // ============================================

  async analyzeProject(path: string, analyzers: string[]): Promise<domain.StaticAnalysisReport> {
    try {
      return await wails.AnalyzeProject(path, analyzers)
    } catch (error) {
      console.error('[ApiService] Error analyzing project:', error)
      throw new Error('Failed to analyze project.')
    }
  }

  async analyzeFile(projectPath: string, filePath: string): Promise<domain.StaticAnalysisResult> {
    try {
      return await wails.AnalyzeFile(projectPath, filePath)
    } catch (error) {
      console.error(`[ApiService] Error analyzing file "${filePath}":`, error)
      throw new Error('Failed to analyze file.')
    }
  }

  async detectLanguages(projectPath: string): Promise<string[]> {
    try {
      return await wails.DetectLanguages(projectPath)
    } catch (error) {
      console.error('[ApiService] Error detecting languages:', error)
      throw new Error('Failed to detect project languages.')
    }
  }

  async getSupportedAnalyzers(): Promise<domain.StaticAnalyzerType[]> {
    try {
      return await wails.GetSupportedAnalyzers()
    } catch (error) {
      console.error('[ApiService] Error getting supported analyzers:', error)
      throw new Error('Failed to get supported analyzers.')
    }
  }

  // ============================================
  // Git операции
  // ============================================

  async getUncommittedFiles(repoPath: string): Promise<domain.FileStatus[]> {
    try {
      return await wails.GetUncommittedFiles(repoPath)
    } catch (error) {
      console.error('[ApiService] Error getting uncommitted files:', error)
      throw new Error('Failed to get git status.')
    }
  }

  async getBranches(repoPath: string): Promise<string> {
    try {
      return await wails.GetBranches(repoPath)
    } catch (error) {
      console.error('[ApiService] Error getting branches:', error)
      throw new Error('Failed to get git branches.')
    }
  }

  async getCurrentBranch(repoPath: string): Promise<string> {
    try {
      return await wails.GetCurrentBranch(repoPath)
    } catch (error) {
      console.error('[ApiService] Error getting current branch:', error)
      throw new Error('Failed to get current branch.')
    }
  }

  async getRichCommitHistory(repoPath: string, filePath: string, limit: number): Promise<domain.CommitWithFiles[]> {
    try {
      return await wails.GetRichCommitHistory(repoPath, filePath, limit)
    } catch (error) {
      console.error('[ApiService] Error getting commit history:', error)
      throw new Error('Failed to get commit history.')
    }
  }

  async isGitAvailable(): Promise<boolean> {
    try {
      return await wails.IsGitAvailable()
    } catch (error) {
      console.error('[ApiService] Error checking git availability:', error)
      return false
    }
  }

  // ============================================
  // Настройки
  // ============================================

  async getSettings(): Promise<domain.SettingsDTO> {
    try {
      return await wails.GetSettings()
    } catch (error) {
      console.error('[ApiService] Error fetching settings:', error)
      throw new Error('Failed to load settings.')
    }
  }

  async saveSettings(settings: string): Promise<void> {
    try {
      return await wails.SaveSettings(settings)
    } catch (error) {
      console.error('[ApiService] Error saving settings:', error)
      throw new Error('Failed to save settings.')
    }
  }

  // ============================================
  // Тестирование
  // ============================================

  async runTests(config: domain.TestConfig): Promise<domain.TestResult[]> {
    try {
      return await wails.RunTests(config)
    } catch (error) {
      console.error('[ApiService] Error running tests:', error)
      throw new Error('Failed to run tests.')
    }
  }

  async discoverTests(projectPath: string, language: string): Promise<domain.TestSuite> {
    try {
      return await wails.DiscoverTests(projectPath, language)
    } catch (error) {
      console.error('[ApiService] Error discovering tests:', error)
      throw new Error('Failed to discover tests.')
    }
  }

  // ============================================
  // Сборка
  // ============================================

  async build(projectPath: string, language: string): Promise<domain.BuildResult> {
    try {
      return await wails.Build(projectPath, language)
    } catch (error) {
      console.error('[ApiService] Error building project:', error)
      throw new Error('Failed to build project.')
    }
  }

  async typeCheck(projectPath: string, language: string): Promise<domain.TypeCheckResult> {
    try {
      return await wails.TypeCheck(projectPath, language)
    } catch (error) {
      console.error('[ApiService] Error type checking:', error)
      throw new Error('Failed to type check.')
    }
  }

  // ============================================
  // Diff и Apply
  // ============================================

  async generateDiff(original: string, modified: string, format: domain.DiffFormat): Promise<domain.DiffResult> {
    try {
      return await wails.GenerateDiff(original, modified, format)
    } catch (error) {
      console.error('[ApiService] Error generating diff:', error)
      throw new Error('Failed to generate diff.')
    }
  }

  async applyEdits(edits: domain.EditsJSON): Promise<domain.ApplyResult[]> {
    try {
      return await wails.ApplyEdits(edits)
    } catch (error) {
      console.error('[ApiService] Error applying edits:', error)
      throw new Error('Failed to apply edits.')
    }
  }

  async applySingleEdit(edit: domain.Edit): Promise<domain.ApplyResult> {
    try {
      return await wails.ApplySingleEdit(edit)
    } catch (error) {
      console.error('[ApiService] Error applying edit:', error)
      throw new Error('Failed to apply edit.')
    }
  }

  // ============================================
  // Отчеты и экспорт
  // ============================================

  async generateReport(contextId: string, format: string): Promise<string> {
    try {
      return await wails.GenerateReport(contextId, format)
    } catch (error) {
      console.error('[ApiService] Error generating report:', error)
      throw new Error('Failed to generate report.')
    }
  }

  async listReports(projectPath: string): Promise<string> {
    try {
      return await wails.ListReports(projectPath)
    } catch (error) {
      console.error('[ApiService] Error listing reports:', error)
      throw new Error('Failed to list reports.')
    }
  }

  async getReport(reportId: string): Promise<string> {
    try {
      return await wails.GetReport(reportId)
    } catch (error) {
      console.error('[ApiService] Error getting report:', error)
      throw new Error('Failed to get report.')
    }
  }

  async exportProject(projectPath: string, format: string, outputPath: string): Promise<string> {
    try {
      return await wails.ExportProject(projectPath, format, outputPath)
    } catch (error) {
      console.error('[ApiService] Error exporting project:', error)
      throw new Error('Failed to export project.')
    }
  }

  // ============================================
  // Task Protocol
  // ============================================

  async executeTaskProtocol(configPath: string): Promise<string> {
    try {
      return await wails.ExecuteTaskProtocol(configPath)
    } catch (error) {
      console.error('[ApiService] Error executing task protocol:', error)
      throw new Error('Failed to execute task protocol.')
    }
  }

  async getTaskProtocolConfiguration(projectPath: string, languages: string[]): Promise<string> {
    try {
      return await wails.GetTaskProtocolConfiguration(projectPath, languages)
    } catch (error) {
      console.error('[ApiService] Error getting task protocol config:', error)
      throw new Error('Failed to get task protocol configuration.')
    }
  }

  // ============================================
  // Guardrails
  // ============================================

  async validatePath(path: string): Promise<domain.GuardrailViolation[]> {
    try {
      return await wails.ValidatePath(path)
    } catch (error) {
      console.error('[ApiService] Error validating path:', error)
      throw new Error('Failed to validate path.')
    }
  }

  async getGuardrailPolicies(): Promise<domain.GuardrailPolicy[]> {
    try {
      return await wails.GetGuardrailPolicies()
    } catch (error) {
      console.error('[ApiService] Error getting guardrail policies:', error)
      throw new Error('Failed to get guardrail policies.')
    }
  }

  async getBudgetPolicies(): Promise<domain.BudgetPolicy[]> {
    try {
      // @ts-ignore
      return await wails.GetBudgetPolicies()
    } catch (error) {
      console.error('[ApiService] Error getting budget policies:', error)
      throw new Error('Failed to get budget policies.')
    }
  }
}

// Экспортируем синглтон экземпляр сервиса
export const apiService = new ApiService()

// Экспортируем класс для возможности моков в тестах
export { ApiService }
