import type { ExportRepository } from '@/domain/repositories/RepositoryInterfaces';
import type { ExportSettings, ExportResult } from '@/types/api';
import { 
  ExportContext,
  ExportProject,
  GetExportHistory,
  CleanupTempFiles
} from '../../../wailsjs/go/main/App';
import { defaultWailsApiAdapter, type WailsApiAdapter } from '../api/WailsApiAdapter';

/**
 * Export API Adapter - Infrastructure implementation of ExportRepository
 * This handles export operations while conforming to Clean Architecture
 */
export class ExportApiAdapter implements ExportRepository {
  private readonly apiAdapter: WailsApiAdapter;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }

  async exportContext(settings: ExportSettings): Promise<ExportResult> {
    try {
      const result = await this.apiAdapter.callApi<ExportSettings, ExportResult>(
        ExportContext,
        settings
      );

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
      await this.apiAdapter.callApi<string, void>(CleanupTempFiles, filePath);
    } catch (error) {
      throw this.handleError(error, 'Failed to cleanup temp files');
    }
  }

  async exportProject(
    projectPath: string, 
    format: string, 
    options?: Record<string, unknown>
  ): Promise<ExportResult> {
    try {
      return await this.apiAdapter.callApi<[string, string, Record<string, unknown>], ExportResult>(
        ExportProject,
        [projectPath, format, options || {}]
      );
    } catch (error) {
      throw this.handleError(error, 'Failed to export project');
    }
  }

  async getExportHistory(projectPath?: string): Promise<ExportResult[]> {
    try {
      return await this.apiAdapter.callApi<string, ExportResult[]>(
        GetExportHistory,
        projectPath || ''
      );
    } catch (error) {
      throw this.handleError(error, 'Failed to get export history');
    }
  }

  // Private helper methods
  private handleError(error: unknown, context: string): Error {
    const message = error instanceof Error ? error.message : String(error);
    
    // Check if this is a domain error from backend
    if (message.startsWith('domain_error:')) {
      try {
        const domainErrorJson = message.substring('domain_error:'.length);
        const domainError = JSON.parse(domainErrorJson);
        
        const structuredError = new Error(`${context}: ${domainError.message}`);
        (structuredError as any).code = domainError.code;
        (structuredError as any).recoverable = domainError.recoverable;
        (structuredError as any).context = domainError.context;
        (structuredError as any).cause = domainError.cause;
        
        return structuredError;
      } catch (parseErr) {
        console.error('Failed to parse domain error:', parseErr);
      }
    }
    
    return new Error(`${context}: ${message}`);
  }
}