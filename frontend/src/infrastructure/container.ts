/**
 * Dependency Injection Container for Clean Architecture
 * This file manages the instantiation and wiring of repositories, adapters, and use cases
 */

// Repository Interfaces
import type { ContextRepository } from '@/domain/repositories/ContextRepository';
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import type { GitRepository } from '@/domain/repositories/GitRepository';
import type { AIRepository } from '@/domain/repositories/AIRepository';
import type { SettingsRepository } from '@/domain/repositories/SettingsRepository';
import type { ReportsRepository } from '@/domain/repositories/ReportsRepository';

// Infrastructure Adapters
import { ContextApiAdapter } from '@/infrastructure/repositories/ContextApiAdapter';
import { ProjectApiAdapter } from '@/infrastructure/repositories/ProjectApiAdapter';
import { GitApiAdapter } from '@/infrastructure/repositories/GitApiAdapter';
import { AIApiAdapter } from '@/infrastructure/repositories/AIApiAdapter';
import { ReportsApiAdapter } from '@/infrastructure/repositories/ReportsApiAdapter';
import { SettingsApiAdapter } from '@/infrastructure/repositories/SettingsApiAdapter';

// Use Cases
import { BuildContextUseCase } from '@/application/use-cases/BuildContextUseCase';
import { GetContextContentUseCase } from '@/application/use-cases/GetContextContentUseCase';
import { CreateStreamingContextUseCase } from '@/application/use-cases/CreateStreamingContextUseCase';

import {
  LoadProjectFilesUseCase,
  SelectProjectDirectoryUseCase,
  ReadFileContentUseCase,
  StartFileWatchingUseCase,
  StopFileWatchingUseCase,
  GetFileStatsUseCase
} from '@/application/use-cases/ProjectUseCases';

// Add the new LoadProjectUseCase
import { LoadProjectUseCase } from '@/application/use-cases/LoadProjectUseCase';

import {
  GetGitBranchesUseCase,
  SwitchGitBranchUseCase,
  GetGitCommitsUseCase,
  GetGitStatusUseCase,
  ValidateGitRepositoryUseCase,
  GetGitRepositoryInfoUseCase
} from '@/application/use-cases/GitUseCases';

import {
  GenerateCodeUseCase,
  AnalyzeCodeUseCase,
  SuggestContextFilesUseCase,
  StartAutonomousTaskUseCase,
  MonitorAutonomousTaskUseCase,
  SmartCodeAssistantUseCase
} from '@/application/use-cases/AIUseCases';

import {
  GetSettingsUseCase,
  SaveSettingsUseCase,
  UpdateAISettingsUseCase,
  RefreshAIModelsUseCase,
  ValidateSettingsUseCase,
  ExportSettingsUseCase,
  ImportSettingsUseCase
} from '@/application/use-cases/SettingsUseCases';

import {
  GetReportUseCase,
  ListReportsUseCase,
  GenerateProjectAnalysisReportUseCase,
  GenerateCodeQualityReportUseCase,
  ExportReportUseCase
} from '@/application/use-cases/ReportsUseCases';

/**
 * Container class for managing dependencies
 */
class Container {
  // Repository instances (singletons)
  private _contextRepository?: ContextRepository;
  private _projectRepository?: ProjectRepository;
  private _gitRepository?: GitRepository;
  private _aiRepository?: AIRepository;
  private _settingsRepository?: SettingsRepository;
  private _reportsRepository?: ReportsRepository;

  // Use case instances (created per request for clean state)
  private useCaseCache = new Map<string, any>();

  /**
   * Get repository instances (singleton pattern)
   */
  get contextRepository(): ContextRepository {
    if (!this._contextRepository) {
      this._contextRepository = new ContextApiAdapter();
    }
    return this._contextRepository;
  }

  get projectRepository(): ProjectRepository {
    if (!this._projectRepository) {
      this._projectRepository = new ProjectApiAdapter();
    }
    return this._projectRepository;
  }

  get gitRepository(): GitRepository {
    if (!this._gitRepository) {
      this._gitRepository = new GitApiAdapter();
    }
    return this._gitRepository;
  }

  get aiRepository(): AIRepository {
    if (!this._aiRepository) {
      this._aiRepository = new AIApiAdapter();
    }
    return this._aiRepository;
  }

  get settingsRepository(): SettingsRepository {
    if (!this._settingsRepository) {
      this._settingsRepository = new SettingsApiAdapter();
    }
    return this._settingsRepository;
  }

  get reportsRepository(): ReportsRepository {
    if (!this._reportsRepository) {
      this._reportsRepository = new ReportsApiAdapter();
    }
    return this._reportsRepository;
  }

  /**
   * Context Use Cases
   */
  getBuildContextUseCase(): BuildContextUseCase {
    return this.getOrCreateUseCase('BuildContextUseCase', () => 
      new BuildContextUseCase(this.contextRepository, this.projectRepository)
    );
  }

  getGetContextContentUseCase(): GetContextContentUseCase {
    return this.getOrCreateUseCase('GetContextContentUseCase', () => 
      new GetContextContentUseCase(this.contextRepository)
    );
  }

  getCreateStreamingContextUseCase(): CreateStreamingContextUseCase {
    return this.getOrCreateUseCase('CreateStreamingContextUseCase', () => 
      new CreateStreamingContextUseCase(this.contextRepository, this.projectRepository)
    );
  }

  /**
   * Project Use Cases
   */
  getLoadProjectUseCase(): LoadProjectUseCase {
    return this.getOrCreateUseCase('LoadProjectUseCase', () => 
      new LoadProjectUseCase(this.projectRepository)
    );
  }

  getGetRecentProjectsUseCase(): SelectProjectDirectoryUseCase {
    return this.getOrCreateUseCase('GetRecentProjectsUseCase', () => 
      new SelectProjectDirectoryUseCase(this.projectRepository)
    );
  }

  getLoadFileTreeUseCase(): ReadFileContentUseCase {
    return this.getOrCreateUseCase('LoadFileTreeUseCase', () => 
      new ReadFileContentUseCase(this.projectRepository)
    );
  }

  getSearchFilesUseCase(): StartFileWatchingUseCase {
    return this.getOrCreateUseCase('SearchFilesUseCase', () => 
      new StartFileWatchingUseCase(this.projectRepository)
    );
  }

  getUpdateFileSelectionUseCase(): StopFileWatchingUseCase {
    return this.getOrCreateUseCase('UpdateFileSelectionUseCase', () => 
      new StopFileWatchingUseCase(this.projectRepository)
    );
  }

  /**
   * Git Use Cases
   */
  getGetGitBranchesUseCase(): GetGitBranchesUseCase {
    return this.getOrCreateUseCase('GetGitBranchesUseCase', () => 
      new GetGitBranchesUseCase(this.gitRepository)
    );
  }

  getSwitchGitBranchUseCase(): SwitchGitBranchUseCase {
    return this.getOrCreateUseCase('SwitchGitBranchUseCase', () => 
      new SwitchGitBranchUseCase(this.gitRepository)
    );
  }

  getGetGitCommitsUseCase(): GetGitCommitsUseCase {
    return this.getOrCreateUseCase('GetGitCommitsUseCase', () => 
      new GetGitCommitsUseCase(this.gitRepository)
    );
  }

  getGetGitStatusUseCase(): GetGitStatusUseCase {
    return this.getOrCreateUseCase('GetGitStatusUseCase', () => 
      new GetGitStatusUseCase(this.gitRepository)
    );
  }

  getValidateGitRepositoryUseCase(): ValidateGitRepositoryUseCase {
    return this.getOrCreateUseCase('ValidateGitRepositoryUseCase', () => 
      new ValidateGitRepositoryUseCase(this.gitRepository)
    );
  }

  getGetGitRepositoryInfoUseCase(): GetGitRepositoryInfoUseCase {
    return this.getOrCreateUseCase('GetGitRepositoryInfoUseCase', () => 
      new GetGitRepositoryInfoUseCase(this.gitRepository)
    );
  }

  /**
   * AI Use Cases
   */
  getGenerateCodeUseCase(): GenerateCodeUseCase {
    return this.getOrCreateUseCase('GenerateCodeUseCase', () => 
      new GenerateCodeUseCase(this.aiRepository)
    );
  }

  getAnalyzeCodeUseCase(): AnalyzeCodeUseCase {
    return this.getOrCreateUseCase('AnalyzeCodeUseCase', () => 
      new AnalyzeCodeUseCase(this.aiRepository)
    );
  }

  getSuggestContextFilesUseCase(): SuggestContextFilesUseCase {
    return this.getOrCreateUseCase('SuggestContextFilesUseCase', () => 
      new SuggestContextFilesUseCase(this.aiRepository)
    );
  }

  getStartAutonomousTaskUseCase(): StartAutonomousTaskUseCase {
    return this.getOrCreateUseCase('StartAutonomousTaskUseCase', () => 
      new StartAutonomousTaskUseCase(this.aiRepository)
    );
  }

  getMonitorAutonomousTaskUseCase(): MonitorAutonomousTaskUseCase {
    return this.getOrCreateUseCase('MonitorAutonomousTaskUseCase', () => 
      new MonitorAutonomousTaskUseCase(this.aiRepository)
    );
  }

  getSmartCodeAssistantUseCase(): SmartCodeAssistantUseCase {
    return this.getOrCreateUseCase('SmartCodeAssistantUseCase', () => 
      new SmartCodeAssistantUseCase(this.aiRepository)
    );
  }

  /**
   * Settings Use Cases
   */
  getGetSettingsUseCase(): GetSettingsUseCase {
    return this.getOrCreateUseCase('GetSettingsUseCase', () => 
      new GetSettingsUseCase(this.settingsRepository)
    );
  }

  getSaveSettingsUseCase(): SaveSettingsUseCase {
    return this.getOrCreateUseCase('SaveSettingsUseCase', () => 
      new SaveSettingsUseCase(this.settingsRepository)
    );
  }

  getUpdateAISettingsUseCase(): UpdateAISettingsUseCase {
    return this.getOrCreateUseCase('UpdateAISettingsUseCase', () => 
      new UpdateAISettingsUseCase(this.settingsRepository)
    );
  }

  getRefreshAIModelsUseCase(): RefreshAIModelsUseCase {
    return this.getOrCreateUseCase('RefreshAIModelsUseCase', () => 
      new RefreshAIModelsUseCase(this.settingsRepository)
    );
  }

  getValidateSettingsUseCase(): ValidateSettingsUseCase {
    return this.getOrCreateUseCase('ValidateSettingsUseCase', () => 
      new ValidateSettingsUseCase(this.settingsRepository)
    );
  }

  getExportSettingsUseCase(): ExportSettingsUseCase {
    return this.getOrCreateUseCase('ExportSettingsUseCase', () => 
      new ExportSettingsUseCase(this.settingsRepository)
    );
  }

  getImportSettingsUseCase(): ImportSettingsUseCase {
    return this.getOrCreateUseCase('ImportSettingsUseCase', () => 
      new ImportSettingsUseCase(this.settingsRepository)
    );
  }

  /**
   * Reports Use Cases
   */
  getGetReportUseCase(): GetReportUseCase {
    return this.getOrCreateUseCase('GetReportUseCase', () => 
      new GetReportUseCase(this.reportsRepository)
    );
  }

  getListReportsUseCase(): ListReportsUseCase {
    return this.getOrCreateUseCase('ListReportsUseCase', () => 
      new ListReportsUseCase(this.reportsRepository)
    );
  }

  getGenerateProjectAnalysisReportUseCase(): GenerateProjectAnalysisReportUseCase {
    return this.getOrCreateUseCase('GenerateProjectAnalysisReportUseCase', () => 
      new GenerateProjectAnalysisReportUseCase(this.reportsRepository)
    );
  }

  getGenerateCodeQualityReportUseCase(): GenerateCodeQualityReportUseCase {
    return this.getOrCreateUseCase('GenerateCodeQualityReportUseCase', () => 
      new GenerateCodeQualityReportUseCase(this.reportsRepository)
    );
  }

  getExportReportUseCase(): ExportReportUseCase {
    return this.getOrCreateUseCase('ExportReportUseCase', () => 
      new ExportReportUseCase(this.reportsRepository)
    );
  }

  /**
   * Utility methods
   */
  private getOrCreateUseCase<T>(key: string, factory: () => T): T {
    if (!this.useCaseCache.has(key)) {
      this.useCaseCache.set(key, factory());
    }
    return this.useCaseCache.get(key);
  }

  /**
   * Clear use case cache (useful for testing or memory cleanup)
   */
  clearUseCaseCache(): void {
    this.useCaseCache.clear();
  }

  /**
   * Reset all dependencies (useful for testing)
   */
  reset(): void {
    this._contextRepository = undefined;
    this._projectRepository = undefined;
    this._gitRepository = undefined;
    this._aiRepository = undefined;
    this._settingsRepository = undefined;
    this._reportsRepository = undefined;
    this.clearUseCaseCache();
  }
}

// Export singleton instance
export const container = new Container();

// Export type for dependency injection
export type AppContainer = Container;