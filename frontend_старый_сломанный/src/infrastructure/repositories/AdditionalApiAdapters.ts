import type { 
  ReportsRepository, 
  AutonomousRepository, 
  ExportRepository 
} from '@/domain/repositories/RepositoryInterfaces';
import type { 
  GenericReport, 
  AutonomousTaskRequest, 
  AutonomousTaskResponse, 
  AutonomousTaskStatus 
} from '@/types/dto';
import type { ExportSettings, ExportResult } from '@/types/api';
import { 
  GetReport,
  ListReports,
  GenerateReport,
  DeleteReport,
  ExportReport,
  StartAutonomousTask,
  CancelAutonomousTask,
  GetAutonomousTaskStatus,
  ListAutonomousTasks,
  GetTaskLogs,
  PauseTask,
  ResumeTask,
  ExportContext,
  ExportProject,
  GetExportHistory,
  CleanupTempFiles
} from '../../../wailsjs/go/main/App';
import { defaultWailsApiAdapter, type WailsApiAdapter } from '../api/WailsApiAdapter';
import { generateReportId } from '@/config/app-config';

/**
 * Reports API Adapter - Infrastructure implementation of ReportsRepository
 */
export class ReportsApiAdapter implements ReportsRepository {
  private readonly apiAdapter: WailsApiAdapter;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }

  async getReport(reportId: string): Promise<GenericReport> {
    return this.apiAdapter.callApi<string, GenericReport>(GetReport, reportId);
  }

  async listReports(reportType?: string): Promise<GenericReport[]> {
    return this.apiAdapter.callApi<string, GenericReport[]>(
      ListReports, 
      reportType || ''
    );
  }

  async generateReport(reportType: string, parameters?: Record<string, unknown>): Promise<GenericReport> {
    return this.apiAdapter.callApi<[string, Record<string, unknown>], GenericReport>(
      GenerateReport,
      [reportType, parameters || {}]
    );
  }

  async deleteReport(reportId: string): Promise<void> {
    return this.apiAdapter.callApi<string, void>(DeleteReport, reportId);
  }

  async exportReport(reportId: string, format: string): Promise<ExportResult> {
    return this.apiAdapter.callApi<[string, string], ExportResult>(
      ExportReport,
      [reportId, format]
    );
  }
}

/**
 * Autonomous API Adapter - Infrastructure implementation of AutonomousRepository
 */
export class AutonomousApiAdapter implements AutonomousRepository {
  private readonly apiAdapter: WailsApiAdapter;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }

  async startAutonomousTask(request: AutonomousTaskRequest): Promise<AutonomousTaskResponse> {
    return this.apiAdapter.callApi<AutonomousTaskRequest, AutonomousTaskResponse>(
      StartAutonomousTask,
      request
    );
  }

  async cancelAutonomousTask(taskId: string): Promise<void> {
    return this.apiAdapter.callApi<string, void>(CancelAutonomousTask, taskId);
  }

  async getAutonomousTaskStatus(taskId: string): Promise<AutonomousTaskStatus> {
    return this.apiAdapter.callApi<string, AutonomousTaskStatus>(
      GetAutonomousTaskStatus,
      taskId
    );
  }

  async listAutonomousTasks(projectPath?: string): Promise<AutonomousTaskStatus[]> {
    return this.apiAdapter.callApi<string, AutonomousTaskStatus[]>(
      ListAutonomousTasks,
      projectPath || ''
    );
  }

  async getTaskLogs(taskId: string): Promise<string[]> {
    return this.apiAdapter.callApi<string, string[]>(GetTaskLogs, taskId);
  }

  async pauseTask(taskId: string): Promise<void> {
    return this.apiAdapter.callApi<string, void>(PauseTask, taskId);
  }

  async resumeTask(taskId: string): Promise<void> {
    return this.apiAdapter.callApi<string, void>(ResumeTask, taskId);
  }
}

/**
 * Export API Adapter - Infrastructure implementation of ExportRepository
 */
export class ExportApiAdapter implements ExportRepository {
  private readonly apiAdapter: WailsApiAdapter;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }

  async exportContext(settings: ExportSettings): Promise<ExportResult> {
    const result = await this.apiAdapter.callApi<ExportSettings, ExportResult>(
      ExportContext,
      settings
    );

    // Handle large files with FilePath
    if (result.filePath && result.isLarge) {
      console.log('Large file exported to:', result.filePath);
    }

    return result;
  }

  async cleanupTempFiles(filePath: string): Promise<void> {
    return this.apiAdapter.callApi<string, void>(CleanupTempFiles, filePath);
  }

  async exportProject(
    projectPath: string, 
    format: string, 
    options?: Record<string, unknown>
  ): Promise<ExportResult> {
    return this.apiAdapter.callApi<[string, string, Record<string, unknown>], ExportResult>(
      ExportProject,
      [projectPath, format, options || {}]
    );
  }

  async getExportHistory(projectPath: string): Promise<ExportResult[]> {
    return this.apiAdapter.callApi<string, ExportResult[]>(
      GetExportHistory,
      projectPath
    );
  }
}
        format,
        includeMetadata: true,
        compressOutput: true,
        ...options
      };
      
      return await this.exportContext(exportSettings);
    } catch (error) {
      throw this.handleError(error, 'Failed to export project');
    }
  }

  async getExportHistory(projectPath?: string): Promise<ExportResult[]> {
    try {
      // TODO: Implement getExportHistory in backend
      return [];
    } catch (error) {
      throw this.handleError(error, 'Failed to get export history');
    }
  }

  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    return new Error(`${context}: ${message}`);
  }
}