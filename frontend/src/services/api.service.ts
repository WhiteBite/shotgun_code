/**
 * Централизованный API сервис для взаимодействия с Go-бэкендом
 * Все вызовы к Wails API должны проходить через этот сервис
 * Это обеспечивает единую точку для обработки ошибок, логирования и тестирования
 */

import * as wails from '#wailsjs/go/main/App'
import type { domain } from '#wailsjs/go/models'

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

  async pathExists(path: string): Promise<boolean> {
    try {
      return await wails.PathExists(path)
    } catch (error) {
      console.error(`[ApiService] Error checking path "${path}":`, error)
      return false
    }
  }

  // ============================================
  // Работа с файлами
  // ============================================

  async listFiles(path: string, useGitignore: boolean = true, useCustomIgnore: boolean = true): Promise<domain.FileNode[]> {
    try {
      return await wails.ListFiles(path, useGitignore, useCustomIgnore)
    } catch (error) {
      console.error(`[ApiService] Error listing files for path "${path}":`, error)
      throw new Error('Failed to load file tree.')
    }
  }

  async clearFileTreeCache(): Promise<void> {
    try {
      await wails.ClearFileTreeCache()
    } catch (error) {
      console.error('[ApiService] Error clearing file tree cache:', error)
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
  ): Promise<domain.ContextSummary> {
    try {
      return await wails.BuildContextFromRequest(projectPath, files, options)
    } catch (error) {
      console.error('[ApiService] Error building context from request:', error)

      // Parse error - handle both string and object formats
      let errorMsg = ''

      if (error && typeof error === 'object') {
        // Handle domain_error format: {"cause":"...", "message":"..."}
        const err = error as any
        errorMsg = err.cause || err.message || String(error)
      } else if (error instanceof Error) {
        errorMsg = error.message
      } else {
        errorMsg = String(error)
      }

      // Extract token limit information
      if (errorMsg.includes('token limit')) {
        // Match pattern: "number > number" or "number \u003e number"
        // Note: \u003e in JSON is already decoded to > in JavaScript string
        const match = errorMsg.match(/(\d+)\s*[>\u003e]\s*(\d+)/)
        if (match) {
          const actual = Number(match[1])
          const limit = Number(match[2])
          // Create structured error with token info for UI to parse
          const error = new Error('TOKEN_LIMIT_EXCEEDED')
            ; (error as any).tokenInfo = { actual, limit }
          throw error
        }
        // If no match but contains "token limit", throw generic message
        throw new Error('TOKEN_LIMIT_EXCEEDED')
      }

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

  async exportContext(exportSettings: unknown): Promise<domain.ExportResult> {
    try {
      return await wails.ExportContext(JSON.stringify(exportSettings))
    } catch (error) {
      console.error('[ApiService] Error exporting context:', error)
      throw new Error('Failed to export context.')
    }
  }

  async getFullContextContent(contextId: string): Promise<string> {
    try {
      return await wails.GetFullContextContent(contextId)
    } catch (error) {
      console.error(`[ApiService] Error getting full context content "${contextId}":`, error)
      throw new Error('Failed to get full context content.')
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

  /**
   * Analyze task and collect relevant context using AI
   * Returns suggested files with relevance scores
   */
  async analyzeTaskAndCollectContext(task: string, allFilesJson: string, rootDir: string): Promise<string> {
    try {
      return await wails.AnalyzeTaskAndCollectContext(task, allFilesJson, rootDir)
    } catch (error) {
      console.error('[ApiService] Error analyzing task:', error)
      throw new Error('Failed to analyze task and collect context.')
    }
  }

  /**
   * Agentic chat with tool use - AI can explore codebase autonomously
   */
  async agenticChat(task: string, projectRoot: string): Promise<AgenticChatResponse> {
    try {
      const request = { task, projectRoot }
      const result = await wails.AgenticChat(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error in agentic chat:', error)
      throw new Error('Failed to execute agentic chat.')
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

  /**
   * Start streaming code generation - results come via Wails events
   */
  generateCodeStream(context: string, task: string): void {
    try {
      wails.GenerateCodeStream(context, task)
    } catch (error) {
      console.error('[ApiService] Error starting code stream:', error)
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

  async getSupportedAnalyzers(): Promise<string[]> {
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

  async isGitRepository(projectPath: string): Promise<boolean> {
    try {
      return await wails.IsGitRepository(projectPath)
    } catch (error) {
      console.error('[ApiService] Error checking git repository:', error)
      return false
    }
  }

  async cloneRepository(url: string): Promise<string> {
    try {
      return await wails.CloneRepository(url)
    } catch (error) {
      console.error('[ApiService] Error cloning repository:', error)
      throw new Error('Failed to clone repository.')
    }
  }

  async checkoutBranch(projectPath: string, branch: string): Promise<void> {
    try {
      await wails.CheckoutBranch(projectPath, branch)
    } catch (error) {
      console.error('[ApiService] Error checking out branch:', error)
      throw new Error('Failed to checkout branch.')
    }
  }

  async checkoutCommit(projectPath: string, commitHash: string): Promise<void> {
    try {
      await wails.CheckoutCommit(projectPath, commitHash)
    } catch (error) {
      console.error('[ApiService] Error checking out commit:', error)
      throw new Error('Failed to checkout commit.')
    }
  }

  async getCommitHistory(projectPath: string, limit: number = 50): Promise<CommitInfo[]> {
    try {
      const result = await wails.GetCommitHistory(projectPath, limit)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting commit history:', error)
      throw new Error('Failed to get commit history.')
    }
  }

  async getRemoteBranches(projectPath: string): Promise<string[]> {
    try {
      const result = await wails.GetRemoteBranches(projectPath)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting remote branches:', error)
      throw new Error('Failed to get remote branches.')
    }
  }

  async cleanupTempRepository(path: string): Promise<void> {
    try {
      await wails.CleanupTempRepository(path)
    } catch (error) {
      console.error('[ApiService] Error cleaning up temp repository:', error)
      // Don't throw - cleanup failure is not critical
    }
  }

  async listFilesAtRef(projectPath: string, ref: string): Promise<string[]> {
    try {
      const result = await wails.ListFilesAtRef(projectPath, ref)
      const parsed = JSON.parse(result)
      return Array.isArray(parsed) ? parsed : []
    } catch (error) {
      console.error('[ApiService] Error listing files at ref:', error)
      throw new Error('Failed to list files at ref.')
    }
  }

  async getFileAtRef(projectPath: string, filePath: string, ref: string): Promise<string> {
    try {
      return await wails.GetFileAtRef(projectPath, filePath, ref)
    } catch (error) {
      console.error('[ApiService] Error getting file at ref:', error)
      throw new Error('Failed to get file at ref.')
    }
  }

  async buildContextAtRef(projectPath: string, files: string[], ref: string): Promise<string> {
    try {
      return await wails.BuildContextAtRef(projectPath, files, ref, '{}')
    } catch (error) {
      console.error('[ApiService] Error building context at ref:', error)
      throw new Error('Failed to build context at ref.')
    }
  }

  // ============================================
  // GitHub API (no clone required)
  // ============================================

  async isGitHubURL(url: string): Promise<boolean> {
    try {
      return await wails.IsGitHubURL(url)
    } catch (error) {
      return false
    }
  }

  async gitHubGetDefaultBranch(repoURL: string): Promise<string> {
    try {
      return await wails.GitHubGetDefaultBranch(repoURL)
    } catch (error) {
      console.error('[ApiService] Error getting GitHub default branch:', error)
      throw new Error('Failed to get default branch.')
    }
  }

  async gitHubGetBranches(repoURL: string): Promise<GitHubBranch[]> {
    try {
      const result = await wails.GitHubGetBranches(repoURL)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting GitHub branches:', error)
      throw new Error('Failed to get GitHub branches.')
    }
  }

  async gitHubGetCommits(repoURL: string, branch: string, limit: number = 50): Promise<GitHubCommit[]> {
    try {
      const result = await wails.GitHubGetCommits(repoURL, branch, limit)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting GitHub commits:', error)
      throw new Error('Failed to get GitHub commits.')
    }
  }

  async gitHubListFiles(repoURL: string, ref: string): Promise<string[]> {
    try {
      const result = await wails.GitHubListFiles(repoURL, ref)
      const parsed = JSON.parse(result)
      return Array.isArray(parsed) ? parsed : []
    } catch (error) {
      console.error('[ApiService] Error listing GitHub files:', error)
      throw new Error('Failed to list GitHub files.')
    }
  }

  async gitHubGetFileContent(repoURL: string, filePath: string, ref: string): Promise<string> {
    try {
      return await wails.GitHubGetFileContent(repoURL, filePath, ref)
    } catch (error) {
      console.error('[ApiService] Error getting GitHub file content:', error)
      throw new Error('Failed to get GitHub file content.')
    }
  }

  async gitHubBuildContext(repoURL: string, files: string[], ref: string): Promise<string> {
    try {
      return await wails.GitHubBuildContext(repoURL, files, ref)
    } catch (error) {
      console.error('[ApiService] Error building GitHub context:', error)
      throw new Error('Failed to build GitHub context.')
    }
  }

  // ============================================
  // GitLab API (no clone required)
  // ============================================

  async isGitLabURL(url: string): Promise<boolean> {
    try {
      return await wails.IsGitLabURL(url)
    } catch (error) {
      return false
    }
  }

  async gitLabGetDefaultBranch(repoURL: string): Promise<string> {
    try {
      return await wails.GitLabGetDefaultBranch(repoURL)
    } catch (error) {
      console.error('[ApiService] Error getting GitLab default branch:', error)
      throw new Error('Failed to get default branch.')
    }
  }

  async gitLabGetBranches(repoURL: string): Promise<GitLabBranch[]> {
    try {
      const result = await wails.GitLabGetBranches(repoURL)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting GitLab branches:', error)
      throw new Error('Failed to get GitLab branches.')
    }
  }

  async gitLabGetCommits(repoURL: string, branch: string, limit: number = 50): Promise<GitLabCommit[]> {
    try {
      const result = await wails.GitLabGetCommits(repoURL, branch, limit)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting GitLab commits:', error)
      throw new Error('Failed to get GitLab commits.')
    }
  }

  async gitLabListFiles(repoURL: string, ref: string): Promise<string[]> {
    try {
      const result = await wails.GitLabListFiles(repoURL, ref)
      const parsed = JSON.parse(result)
      return Array.isArray(parsed) ? parsed : []
    } catch (error) {
      console.error('[ApiService] Error listing GitLab files:', error)
      throw new Error('Failed to list GitLab files.')
    }
  }

  async gitLabGetFileContent(repoURL: string, filePath: string, ref: string): Promise<string> {
    try {
      return await wails.GitLabGetFileContent(repoURL, filePath, ref)
    } catch (error) {
      console.error('[ApiService] Error getting GitLab file content:', error)
      throw new Error('Failed to get GitLab file content.')
    }
  }

  async gitLabBuildContext(repoURL: string, files: string[], ref: string): Promise<string> {
    try {
      return await wails.GitLabBuildContext(repoURL, files, ref)
    } catch (error) {
      console.error('[ApiService] Error building GitLab context:', error)
      throw new Error('Failed to build GitLab context.')
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
  // Ignore Rules
  // ============================================

  async getGitignoreContent(projectPath: string): Promise<string> {
    try {
      return await wails.GetGitignoreContentForProject(projectPath)
    } catch (error) {
      console.error('[ApiService] Error getting .gitignore content:', error)
      throw new Error('Failed to load .gitignore content.')
    }
  }

  async getCustomIgnoreRules(): Promise<string> {
    try {
      return await wails.GetCustomIgnoreRules()
    } catch (error) {
      console.error('[ApiService] Error getting custom ignore rules:', error)
      throw new Error('Failed to load custom ignore rules.')
    }
  }

  async updateCustomIgnoreRules(rules: string): Promise<void> {
    try {
      return await wails.UpdateCustomIgnoreRules(rules)
    } catch (error) {
      console.error('[ApiService] Error updating custom ignore rules:', error)
      throw new Error('Failed to update custom ignore rules.')
    }
  }

  async testIgnoreRules(projectPath: string, rules: string): Promise<string[]> {
    try {
      return await wails.TestIgnoreRules(projectPath, rules)
    } catch (error) {
      console.error('[ApiService] Error testing ignore rules:', error)
      throw new Error('Failed to test ignore rules.')
    }
  }

  async addToGitignore(projectPath: string, pattern: string): Promise<void> {
    try {
      return await wails.AddToGitignore(projectPath, pattern)
    } catch (error) {
      console.error('[ApiService] Error adding to .gitignore:', error)
      throw new Error('Failed to add to .gitignore.')
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

  async generateDiff(original: string, modified: string, format: string): Promise<domain.DiffResult> {
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

  // ============================================
  // Qwen Task Execution
  // ============================================

  /**
   * Execute a task using Qwen with smart context collection
   */
  async qwenExecuteTask(request: QwenTaskRequest): Promise<QwenTaskResponse> {
    try {
      const result = await wails.QwenExecuteTask(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error executing Qwen task:', error)
      throw new Error('Failed to execute task with Qwen.')
    }
  }

  /**
   * Preview the context that would be collected for a task
   */
  async qwenPreviewContext(request: QwenTaskRequest): Promise<QwenContextPreview> {
    try {
      const result = await wails.QwenPreviewContext(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error previewing Qwen context:', error)
      throw new Error('Failed to preview context.')
    }
  }

  /**
   * Get available Qwen models
   */
  async qwenGetAvailableModels(): Promise<QwenModelInfo[]> {
    try {
      const result = await wails.QwenGetAvailableModels()
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting Qwen models:', error)
      throw new Error('Failed to get Qwen models.')
    }
  }

  // ============================================
  // ============================================
  // Semantic Search
  // ============================================

  /**
   * Check if semantic search is available
   */
  async isSemanticSearchAvailable(): Promise<boolean> {
    try {
      // @ts-ignore - method may not exist in wails bindings yet
      return await wails.IsSemanticSearchAvailable()
    } catch (error) {
      console.warn('[ApiService] Semantic search not available:', error)
      return false
    }
  }

  /**
   * Perform semantic search on the project
   */
  async semanticSearch(request: SemanticSearchRequest): Promise<SemanticSearchResponse> {
    try {
      // @ts-ignore
      const result = await wails.SemanticSearch(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error in semantic search:', error)
      throw new Error('Failed to perform semantic search.')
    }
  }

  /**
   * Find similar code
   */
  async semanticFindSimilar(request: FindSimilarRequest): Promise<SemanticSearchResponse> {
    try {
      // @ts-ignore
      const result = await wails.SemanticFindSimilar(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error finding similar code:', error)
      throw new Error('Failed to find similar code.')
    }
  }

  /**
   * Index a project for semantic search
   */
  async semanticIndexProject(projectRoot: string): Promise<void> {
    try {
      // @ts-ignore
      await wails.SemanticIndexProject(projectRoot)
    } catch (error) {
      console.error('[ApiService] Error indexing project:', error)
      throw new Error('Failed to index project.')
    }
  }

  /**
   * Index a single file
   */
  async semanticIndexFile(projectRoot: string, filePath: string): Promise<void> {
    try {
      // @ts-ignore
      await wails.SemanticIndexFile(projectRoot, filePath)
    } catch (error) {
      console.error('[ApiService] Error indexing file:', error)
      throw new Error('Failed to index file.')
    }
  }

  /**
   * Get semantic search index statistics
   */
  async semanticGetStats(projectRoot: string): Promise<SemanticIndexStats> {
    try {
      // @ts-ignore
      const result = await wails.SemanticGetStats(projectRoot)
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error getting semantic stats:', error)
      throw new Error('Failed to get semantic search statistics.')
    }
  }

  /**
   * Check if a project is indexed
   */
  async semanticIsIndexed(projectRoot: string): Promise<boolean> {
    try {
      // @ts-ignore
      return await wails.SemanticIsIndexed(projectRoot)
    } catch (error) {
      console.warn('[ApiService] Error checking index status:', error)
      return false
    }
  }

  /**
   * Retrieve relevant context using RAG
   */
  async semanticRetrieveContext(request: RetrieveContextRequest): Promise<CodeChunk[]> {
    try {
      // @ts-ignore
      const result = await wails.SemanticRetrieveContext(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error retrieving context:', error)
      throw new Error('Failed to retrieve context.')
    }
  }

  /**
   * Perform hybrid keyword + semantic search
   */
  async semanticHybridSearch(request: SemanticSearchRequest): Promise<SemanticSearchResponse> {
    try {
      // @ts-ignore
      const result = await wails.SemanticHybridSearch(JSON.stringify(request))
      return JSON.parse(result)
    } catch (error) {
      console.error('[ApiService] Error in hybrid search:', error)
      throw new Error('Failed to perform hybrid search.')
    }
  }
}

// Agentic Chat types
export interface AgenticChatRequest {
  task: string
  projectRoot: string
  maxTokens?: number
}

// ============================================
// Semantic Search Types
// ============================================

export interface SemanticSearchRequest {
  query: string
  projectRoot: string
  topK?: number
  minScore?: number
  searchType?: 'semantic' | 'keyword' | 'hybrid'
  languages?: string[]
  chunkTypes?: string[]
}

export interface SemanticSearchResponse {
  results: SemanticSearchResult[]
  totalResults: number
  queryTime: number
  searchType: string
}

export interface SemanticSearchResult {
  chunk: CodeChunk
  score: number
  highlights?: string[]
  reason?: string
}

export interface CodeChunk {
  id: string
  filePath: string
  content: string
  startLine: number
  endLine: number
  chunkType: 'file' | 'function' | 'class' | 'method' | 'block'
  symbolName?: string
  symbolKind?: string
  language: string
  tokenCount: number
  hash: string
}

export interface FindSimilarRequest {
  filePath: string
  startLine: number
  endLine: number
  topK?: number
  minScore?: number
  excludeSelf?: boolean
}

export interface RetrieveContextRequest {
  query: string
  projectRoot: string
  maxTokens?: number
}

export interface SemanticIndexStats {
  totalChunks: number
  totalFiles: number
  totalTokens: number
  lastUpdated: string
  indexSize: number
  dimensions: number
}

// ============================================
// Qwen Types
// ============================================

export interface QwenTaskRequest {
  task: string
  projectRoot: string
  selectedFiles?: string[]
  selectedCode?: string
  sourceFile?: string
  model?: string
  maxTokens?: number
  temperature?: number
}

export interface QwenTaskResponse {
  content: string
  model: string
  tokensUsed: number
  processingTime: string
  contextSummary: QwenContextSummary
  success: boolean
  error?: string
}

export interface QwenContextSummary {
  totalFiles: number
  totalTokens: number
  includedFiles: string[]
  truncatedFiles: string[]
  excludedFiles: string[]
}

export interface QwenContextPreview {
  totalFiles: number
  totalTokens: number
  files: QwenFilePreview[]
  truncatedFiles: string[]
  excludedFiles: string[]
  callStackInfo?: QwenCallStackInfo
  relevanceScores: Record<string, number>
}

export interface QwenFilePreview {
  path: string
  tokens: number
  relevance: number
  reason: string
}

export interface QwenCallStackInfo {
  rootSymbol: string
  callers: string[]
  callees: string[]
  dependencies: string[]
}

export interface QwenModelInfo {
  id: string
  name: string
  description: string
  maxContext: number
  recommended: boolean
}

// Git types
export interface CommitInfo {
  hash: string
  subject: string
  author: string
  date: string
}

// GitHub API types
export interface GitHubBranch {
  name: string
  commit: {
    sha: string
  }
}

export interface GitHubCommit {
  sha: string
  commit: {
    message: string
    author: {
      name: string
      date: string
    }
  }
}

// GitLab API types
export interface GitLabBranch {
  name: string
  commit: {
    id: string
  }
  default: boolean
}

export interface GitLabCommit {
  id: string
  short_id: string
  title: string
  message: string
  author_name: string
  committed_date: string
}

// Agentic Chat types
export interface AgenticChatResponse {
  response: string
  toolCalls: ToolCallLog[]
  iterations: number
  context: string[]
}

export interface ToolCallLog {
  tool: string
  arguments: string
  result: string
}

// Экспортируем синглтон экземпляр сервиса
export const apiService = new ApiService()

// Экспортируем класс для возможности моков в тестах
export { ApiService }

