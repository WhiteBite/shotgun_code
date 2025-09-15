import type { ReportsRepository } from '@/domain/repositories/ReportsRepository';
import type { GenericReport } from '@/types/dto';

/**
 * Report Management Use Cases
 */

export class GetReportUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(reportId: string): Promise<GenericReport> {
    this.validateReportId(reportId);
    
    try {
      const report = await this.reportsRepo.getReport(reportId);
      
      // Validate report structure
      this.validateReportStructure(report);
      
      return report;
    } catch (error) {
      throw new Error(`Failed to get report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateReportId(reportId: string): void {
    if (!reportId?.trim()) {
      throw new Error('Report ID is required');
    }
    
    if (!/^[a-zA-Z0-9_-]+$/.test(reportId)) {
      throw new Error('Invalid report ID format');
    }
  }

  private validateReportStructure(report: GenericReport): void {
    if (!report || typeof report !== 'object') {
      throw new Error('Invalid report structure');
    }
    
    if (!report.id) {
      throw new Error('Report missing ID');
    }
    
    if (!report.type) {
      throw new Error('Report missing type');
    }
    
    if (!report.createdAt) {
      console.warn('Report missing creation date');
    }
  }
}

export class ListReportsUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(
    filters?: {
      type?: string;
      createdAfter?: Date;
      createdBefore?: Date;
      status?: string;
      limit?: number;
      offset?: number;
    }
  ): Promise<{
    reports: GenericReport[];
    totalCount: number;
    hasMore: boolean;
  }> {
    this.validateListFilters(filters);
    
    try {
      const reports = await this.reportsRepo.listReports(filters?.type);
      
      // Apply additional filters
      let filteredReports = reports;
      
      if (filters?.createdAfter) {
        filteredReports = filteredReports.filter(report => 
          new Date(report.createdAt) >= filters.createdAfter!
        );
      }
      
      if (filters?.createdBefore) {
        filteredReports = filteredReports.filter(report => 
          new Date(report.createdAt) <= filters.createdBefore!
        );
      }
      
      if (filters?.status) {
        filteredReports = filteredReports.filter(report => 
          report.status === filters.status
        );
      }
      
      // Sort by creation date (newest first)
      filteredReports.sort((a, b) => 
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      );
      
      // Apply pagination
      const offset = filters?.offset || 0;
      const limit = filters?.limit || 50;
      const totalCount = filteredReports.length;
      const paginatedReports = filteredReports.slice(offset, offset + limit);
      const hasMore = offset + limit < totalCount;
      
      return {
        reports: paginatedReports,
        totalCount,
        hasMore
      };
    } catch (error) {
      throw new Error(`Failed to list reports: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateListFilters(filters?: any): void {
    if (!filters) return;
    
    if (filters.limit !== undefined) {
      if (typeof filters.limit !== 'number' || filters.limit < 1 || filters.limit > 1000) {
        throw new Error('Limit must be between 1 and 1000');
      }
    }
    
    if (filters.offset !== undefined) {
      if (typeof filters.offset !== 'number' || filters.offset < 0) {
        throw new Error('Offset must be a non-negative number');
      }
    }
    
    if (filters.createdAfter && !(filters.createdAfter instanceof Date)) {
      throw new Error('CreatedAfter must be a Date object');
    }
    
    if (filters.createdBefore && !(filters.createdBefore instanceof Date)) {
      throw new Error('CreatedBefore must be a Date object');
    }
  }
}

export class DeleteReportUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(reportId: string): Promise<void> {
    this.validateReportId(reportId);
    
    try {
      // Check if report exists before deletion
      await this.reportsRepo.getReport(reportId);
      
      // Delete the report
      await this.reportsRepo.deleteReport(reportId);
    } catch (error) {
      if (error instanceof Error && error.message.includes('not found')) {
        throw new Error(`Report not found: ${reportId}`);
      }
      throw new Error(`Failed to delete report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateReportId(reportId: string): void {
    if (!reportId?.trim()) {
      throw new Error('Report ID is required');
    }
  }
}

/**
 * Report Generation Use Cases
 */

export class GenerateProjectAnalysisReportUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(
    projectPath: string,
    options?: {
      includeFileTree?: boolean;
      includeMetrics?: boolean;
      includeGitHistory?: boolean;
      analyzeDependencies?: boolean;
      outputFormat?: 'json' | 'html' | 'markdown';
    }
  ): Promise<{
    reportId: string;
    report: GenericReport;
    generatedAt: Date;
  }> {
    this.validateProjectAnalysisInputs(projectPath, options);
    
    try {
      const reportData = await this.analyzeProject(projectPath, options);
      
      const report: GenericReport = {
        id: this.generateReportId('project-analysis'),
        type: 'project-analysis',
        title: `Project Analysis - ${this.extractProjectName(projectPath)}`,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        status: 'completed',
        data: reportData,
        metadata: {
          projectPath,
          options: options || {},
          generatedBy: 'system',
          version: '1.0.0'
        }
      };
      
      // Save the report
      await this.reportsRepo.createReport(report);
      
      return {
        reportId: report.id,
        report,
        generatedAt: new Date(report.createdAt)
      };
    } catch (error) {
      throw new Error(`Failed to generate project analysis report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateProjectAnalysisInputs(projectPath: string, options?: any): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    
    if (options?.outputFormat && !['json', 'html', 'markdown'].includes(options.outputFormat)) {
      throw new Error('Invalid output format. Must be json, html, or markdown');
    }
  }

  private async analyzeProject(projectPath: string, options?: any): Promise<any> {
    // This would typically call various analysis services
    // For now, return a basic structure
    const analysis = {
      summary: {
        projectPath,
        analyzedAt: new Date().toISOString(),
        analysisOptions: options || {}
      },
      metrics: {
        fileCount: 0,
        totalLines: 0,
        languages: [],
        dependencies: []
      },
      findings: [],
      recommendations: []
    };
    
    // Add more detailed analysis based on options
    if (options?.includeMetrics) {
      analysis.metrics.fileCount = await this.countFiles(projectPath);
    }
    
    if (options?.includeGitHistory) {
      analysis.metrics = {
        ...analysis.metrics,
        gitStats: await this.getGitStats(projectPath)
      };
    }
    
    return analysis;
  }

  private async countFiles(projectPath: string): Promise<number> {
    // Placeholder - would integrate with file system service
    return 100;
  }

  private async getGitStats(projectPath: string): Promise<any> {
    // Placeholder - would integrate with git service
    return {
      commitCount: 50,
      branches: 3,
      contributors: 2
    };
  }

  private extractProjectName(projectPath: string): string {
    return projectPath.split(/[/\\]/).pop() || 'Unknown Project';
  }

  private generateReportId(type: string): string {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substr(2, 9);
    return `${type}-${timestamp}-${random}`;
  }
}

export class GenerateCodeQualityReportUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(
    projectPath: string,
    filePaths?: string[],
    options?: {
      checkComplexity?: boolean;
      checkDuplication?: boolean;
      checkSecurity?: boolean;
      checkPerformance?: boolean;
      includeMetrics?: boolean;
    }
  ): Promise<{
    reportId: string;
    report: GenericReport;
    summary: {
      overallScore: number;
      issuesFound: number;
      critical: number;
      major: number;
      minor: number;
    };
  }> {
    this.validateCodeQualityInputs(projectPath, filePaths, options);
    
    try {
      const analysisResult = await this.analyzeCodeQuality(projectPath, filePaths, options);
      
      const report: GenericReport = {
        id: this.generateReportId('code-quality'),
        type: 'code-quality',
        title: `Code Quality Report - ${this.extractProjectName(projectPath)}`,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        status: 'completed',
        data: analysisResult,
        metadata: {
          projectPath,
          filePaths: filePaths || [],
          options: options || {},
          generatedBy: 'system',
          version: '1.0.0'
        }
      };
      
      await this.reportsRepo.createReport(report);
      
      return {
        reportId: report.id,
        report,
        summary: this.calculateQualitySummary(analysisResult)
      };
    } catch (error) {
      throw new Error(`Failed to generate code quality report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateCodeQualityInputs(projectPath: string, filePaths?: string[], options?: any): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    
    if (filePaths && filePaths.length > 1000) {
      throw new Error('Too many files to analyze (maximum 1000)');
    }
  }

  private async analyzeCodeQuality(projectPath: string, filePaths?: string[], options?: any): Promise<any> {
    // Placeholder for actual code quality analysis
    return {
      summary: {
        totalFiles: filePaths?.length || 0,
        analyzedFiles: filePaths?.length || 0,
        skippedFiles: 0
      },
      issues: [
        {
          type: 'complexity',
          severity: 'major',
          file: 'example.ts',
          line: 45,
          message: 'Function complexity too high',
          rule: 'max-complexity'
        }
      ],
      metrics: {
        complexity: {
          average: 3.2,
          highest: 12,
          files: []
        },
        duplication: {
          percentage: 5.2,
          blocks: []
        }
      }
    };
  }

  private calculateQualitySummary(analysisResult: any): any {
    const issues = analysisResult.issues || [];
    const critical = issues.filter((i: any) => i.severity === 'critical').length;
    const major = issues.filter((i: any) => i.severity === 'major').length;
    const minor = issues.filter((i: any) => i.severity === 'minor').length;
    
    // Simple scoring algorithm
    const totalIssues = critical + major + minor;
    const baseScore = 100;
    const deduction = (critical * 10) + (major * 5) + (minor * 1);
    const overallScore = Math.max(0, baseScore - deduction);
    
    return {
      overallScore,
      issuesFound: totalIssues,
      critical,
      major,
      minor
    };
  }

  private extractProjectName(projectPath: string): string {
    return projectPath.split(/[/\\]/).pop() || 'Unknown Project';
  }

  private generateReportId(type: string): string {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substr(2, 9);
    return `${type}-${timestamp}-${random}`;
  }
}

/**
 * Report Export Use Cases
 */

export class ExportReportUseCase {
  constructor(private reportsRepo: ReportsRepository) {}

  async execute(
    reportId: string,
    format: 'json' | 'csv' | 'html' | 'pdf' = 'json',
    options?: {
      includeRawData?: boolean;
      includeMetadata?: boolean;
      compress?: boolean;
    }
  ): Promise<{
    exportData: string | ArrayBuffer;
    filename: string;
    mimeType: string;
    size: number;
  }> {
    this.validateExportInputs(reportId, format);
    
    try {
      const report = await this.reportsRepo.getReport(reportId);
      const exportData = await this.formatReportForExport(report, format, options);
      
      const filename = this.generateExportFilename(report, format);
      const mimeType = this.getMimeType(format);
      const size = typeof exportData === 'string' ? 
        new Blob([exportData]).size : 
        exportData.byteLength;
      
      return {
        exportData,
        filename,
        mimeType,
        size
      };
    } catch (error) {
      throw new Error(`Failed to export report: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateExportInputs(reportId: string, format: string): void {
    if (!reportId?.trim()) {
      throw new Error('Report ID is required');
    }
    
    const validFormats = ['json', 'csv', 'html', 'pdf'];
    if (!validFormats.includes(format)) {
      throw new Error(`Invalid export format: ${format}`);
    }
  }

  private async formatReportForExport(report: GenericReport, format: string, options?: any): Promise<string> {
    switch (format) {
      case 'json':
        const exportObject = {
          report: options?.includeRawData === false ? 
            { ...report, data: undefined } : report,
          exportedAt: new Date().toISOString(),
          exportOptions: options || {}
        };
        return JSON.stringify(exportObject, null, 2);
        
      case 'csv':
        return this.convertToCSV(report);
        
      case 'html':
        return this.convertToHTML(report);
        
      case 'pdf':
        // This would typically use a PDF generation library
        return this.convertToHTML(report); // Fallback to HTML for now
        
      default:
        throw new Error(`Unsupported export format: ${format}`);
    }
  }

  private convertToCSV(report: GenericReport): string {
    // Basic CSV export - would be enhanced based on report structure
    const lines = [
      'Field,Value',
      `ID,${report.id}`,
      `Type,${report.type}`,
      `Title,${report.title}`,
      `Created,${report.createdAt}`,
      `Status,${report.status}`
    ];
    
    return lines.join('\n');
  }

  private convertToHTML(report: GenericReport): string {
    return `
<!DOCTYPE html>
<html>
<head>
    <title>${report.title}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #ccc; padding-bottom: 20px; }
        .content { margin-top: 20px; }
        .metadata { background: #f5f5f5; padding: 15px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>${report.title}</h1>
        <p>Report ID: ${report.id}</p>
        <p>Generated: ${new Date(report.createdAt).toLocaleString()}</p>
        <p>Status: ${report.status}</p>
    </div>
    
    <div class="content">
        <h2>Report Data</h2>
        <pre>${JSON.stringify(report.data, null, 2)}</pre>
    </div>
    
    <div class="metadata">
        <h3>Metadata</h3>
        <pre>${JSON.stringify(report.metadata, null, 2)}</pre>
    </div>
</body>
</html>`;
  }

  private generateExportFilename(report: GenericReport, format: string): string {
    const sanitizedTitle = report.title.replace(/[^a-zA-Z0-9-_]/g, '_');
    const timestamp = new Date().toISOString().split('T')[0];
    return `${sanitizedTitle}_${timestamp}.${format}`;
  }

  private getMimeType(format: string): string {
    const mimeTypes: Record<string, string> = {
      json: 'application/json',
      csv: 'text/csv',
      html: 'text/html',
      pdf: 'application/pdf'
    };
    
    return mimeTypes[format] || 'application/octet-stream';
  }
}