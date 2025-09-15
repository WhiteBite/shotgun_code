import type { 
  GenericReport, 
  AutonomousTaskRequest, 
  AutonomousTaskResponse, 
  AutonomousTaskStatus 
} from '@/types/dto';
import type { ExportSettings, ExportResult } from '@/types/api';

/**
 * Repository interface for report operations
 */
export interface ReportsRepository {
  /**
   * Get a specific report by ID
   * @param reportId Report identifier
   * @returns Report data
   */
  getReport(reportId: string): Promise<GenericReport>;

  /**
   * List all reports, optionally filtered by type
   * @param reportType Optional report type filter
   * @returns Array of reports
   */
  listReports(reportType?: string): Promise<GenericReport[]>;

  /**
   * Generate a new report
   * @param reportType Type of report to generate
   * @param parameters Report parameters
   * @returns Generated report
   */
  generateReport(reportType: string, parameters?: Record<string, unknown>): Promise<GenericReport>;

  /**
   * Delete a report
   * @param reportId Report identifier
   */
  deleteReport(reportId: string): Promise<void>;

  /**
   * Export report to file
   * @param reportId Report identifier
   * @param format Export format (pdf, json, csv, etc.)
   * @returns Export result with file path
   */
  exportReport(reportId: string, format: string): Promise<ExportResult>;
}

/**
 * Repository interface for autonomous task operations
 */
export interface AutonomousRepository {
  /**
   * Start an autonomous task
   * @param request Task request
   * @returns Task response with ID
   */
  startAutonomousTask(request: AutonomousTaskRequest): Promise<AutonomousTaskResponse>;

  /**
   * Cancel a running autonomous task
   * @param taskId Task identifier
   */
  cancelAutonomousTask(taskId: string): Promise<void>;

  /**
   * Get status of an autonomous task
   * @param taskId Task identifier
   * @returns Task status
   */
  getAutonomousTaskStatus(taskId: string): Promise<AutonomousTaskStatus>;

  /**
   * List all autonomous tasks for a project
   * @param projectPath Project path
   * @returns Array of task statuses
   */
  listAutonomousTasks(projectPath?: string): Promise<AutonomousTaskStatus[]>;

  /**
   * Get task execution logs
   * @param taskId Task identifier
   * @returns Task logs
   */
  getTaskLogs(taskId: string): Promise<string[]>;

  /**
   * Pause a running task
   * @param taskId Task identifier
   */
  pauseTask(taskId: string): Promise<void>;

  /**
   * Resume a paused task
   * @param taskId Task identifier
   */
  resumeTask(taskId: string): Promise<void>;
}

/**
 * Repository interface for export operations
 */
export interface ExportRepository {
  /**
   * Export context to file
   * @param settings Export settings
   * @returns Export result
   */
  exportContext(settings: ExportSettings): Promise<ExportResult>;

  /**
   * Clean up temporary files
   * @param filePath Temporary file path
   */
  cleanupTempFiles(filePath: string): Promise<void>;

  /**
   * Export project data
   * @param projectPath Project path
   * @param format Export format
   * @param options Export options
   * @returns Export result
   */
  exportProject(
    projectPath: string, 
    format: string, 
    options?: Record<string, unknown>
  ): Promise<ExportResult>;

  /**
   * Get export history
   * @param projectPath Optional project path filter
   * @returns Array of export records
   */
  getExportHistory(projectPath?: string): Promise<ExportResult[]>;
}