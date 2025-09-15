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
  StartAutonomousTask,
  CancelAutonomousTask,
  GetAutonomousTaskStatus,
  ExportContext
} from '../../../wailsjs/go/main/App';

/**
 * Reports API Adapter - Infrastructure implementation of ReportsRepository
 */
export class ReportsApiAdapter implements ReportsRepository {
  async getReport(reportId: string): Promise<GenericReport> {
    try {
      const reportJson = await GetReport(reportId);
      return JSON.parse(reportJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to get report');
    }
  }

  async listReports(reportType?: string): Promise<GenericReport[]> {
    try {
      const reportsJson = await ListReports(reportType || '');
      return JSON.parse(reportsJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to list reports');
    }
  }

  async generateReport(reportType: string, parameters?: Record<string, any>): Promise<GenericReport> {
    try {
      // TODO: Implement generateReport in backend
      // For now, return a placeholder
      return {
        id: `${reportType}-${Date.now()}`,
        type: reportType,
        title: `Generated ${reportType} Report`,
        createdAt: new Date().toISOString(),
        data: parameters || {},
        status: 'completed'
      };
    } catch (error) {
      throw this.handleError(error, 'Failed to generate report');
    }
  }

  async deleteReport(reportId: string): Promise<void> {
    try {
      // TODO: Implement deleteReport in backend
      console.log('Delete report:', reportId);
    } catch (error) {
      throw this.handleError(error, 'Failed to delete report');
    }
  }

  async exportReport(reportId: string, format: string): Promise<ExportResult> {
    try {
      // TODO: Implement exportReport in backend
      return {
        success: true,
        filePath: `/tmp/report-${reportId}.${format}`,
        isLarge: false,
        message: `Report exported as ${format}`
      };
    } catch (error) {
      throw this.handleError(error, 'Failed to export report');
    }
  }

  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    return new Error(`${context}: ${message}`);
  }
}

/**
 * Autonomous API Adapter - Infrastructure implementation of AutonomousRepository
 */
export class AutonomousApiAdapter implements AutonomousRepository {
  async startAutonomousTask(request: AutonomousTaskRequest): Promise<AutonomousTaskResponse> {
    try {
      const requestJson = JSON.stringify(request);
      const responseJson = await StartAutonomousTask(requestJson);
      return JSON.parse(responseJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to start autonomous task');
    }
  }

  async cancelAutonomousTask(taskId: string): Promise<void> {
    try {
      await CancelAutonomousTask(taskId);
    } catch (error) {
      throw this.handleError(error, 'Failed to cancel autonomous task');
    }
  }

  async getAutonomousTaskStatus(taskId: string): Promise<AutonomousTaskStatus> {
    try {
      const statusJson = await GetAutonomousTaskStatus(taskId);
      return JSON.parse(statusJson);
    } catch (error) {
      throw this.handleError(error, 'Failed to get autonomous task status');
    }
  }

  async listAutonomousTasks(projectPath?: string): Promise<AutonomousTaskStatus[]> {
    try {
      // TODO: Implement listAutonomousTasks in backend
      // For now, return empty array
      return [];
    } catch (error) {
      throw this.handleError(error, 'Failed to list autonomous tasks');
    }
  }

  async getTaskLogs(taskId: string): Promise<string[]> {
    try {
      // TODO: Implement getTaskLogs in backend
      return [`Log entry for task ${taskId}`];
    } catch (error) {
      throw this.handleError(error, 'Failed to get task logs');
    }
  }

  async pauseTask(taskId: string): Promise<void> {
    try {
      // TODO: Implement pauseTask in backend
      console.log('Pause task:', taskId);
    } catch (error) {
      throw this.handleError(error, 'Failed to pause task');
    }
  }

  async resumeTask(taskId: string): Promise<void> {
    try {
      // TODO: Implement resumeTask in backend
      console.log('Resume task:', taskId);
    } catch (error) {
      throw this.handleError(error, 'Failed to resume task');
    }
  }

  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    return new Error(`${context}: ${message}`);
  }
}

/**
 * Export API Adapter - Infrastructure implementation of ExportRepository
 */
export class ExportApiAdapter implements ExportRepository {
  async exportContext(settings: ExportSettings): Promise<ExportResult> {
    try {
      const result = await ExportContext(JSON.stringify(settings));

      // Handle large files with FilePath
      if (result.filePath && result.isLarge) {
        console.log('Large file exported to:', result.filePath);
      }

      return result;
    } catch (error) {
      throw this.handleError(error, 'Failed to export context');
    }
  }

  async cleanupTempFiles(filePath: string): Promise<void> {
    try {
      // TODO: Implement cleanupTempFiles in backend
      console.log('Cleanup temp file:', filePath);
    } catch (error) {
      throw this.handleError(error, 'Failed to cleanup temp files');
    }
  }

  async exportProject(
    projectPath: string, 
    format: string, 
    options?: Record<string, any>
  ): Promise<ExportResult> {
    try {
      // Use context export as base for project export
      const exportSettings: ExportSettings = {
        projectPath,
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