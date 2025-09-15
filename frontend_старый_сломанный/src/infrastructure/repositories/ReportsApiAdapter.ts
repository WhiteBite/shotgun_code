import type { ReportsRepository } from '@/domain/repositories/ReportsRepository';
import type { GenericReport } from '@/types/dto';
import {
  ListReports,
  GetReport
} from '../../../wailsjs/go/main/App';

/**
 * Reports API Adapter - Infrastructure implementation of ReportsRepository
 * This handles report operations via the backend API
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

  async createReport(report: GenericReport): Promise<GenericReport> {
    try {
      // For now, we'll store it in memory or return as-is
      // In a real implementation, this would call a backend API
      console.log('Creating report:', report);
      return report;
    } catch (error) {
      throw this.handleError(error, 'Failed to create report');
    }
  }

  async updateReport(report: GenericReport): Promise<GenericReport> {
    try {
      // For now, we'll return as-is
      // In a real implementation, this would call a backend API
      console.log('Updating report:', report);
      return report;
    } catch (error) {
      throw this.handleError(error, 'Failed to update report');
    }
  }

  async deleteReport(reportId: string): Promise<void> {
    try {
      // For now, we'll just log
      // In a real implementation, this would call a backend API
      console.log('Deleting report:', reportId);
    } catch (error) {
      throw this.handleError(error, 'Failed to delete report');
    }
  }

  async generateReport(
    type: string,
    params: Record<string, any>
  ): Promise<GenericReport> {
    try {
      // For now, we'll generate a basic report structure
      // In a real implementation, this would call a backend API
      const report: GenericReport = {
        id: `${type}-${Date.now()}`,
        type,
        title: `Generated ${type} Report`,
        summary: `Report generated with parameters: ${JSON.stringify(params)}`,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        status: 'completed',
        data: params,
        metadata: {
          generator: 'ReportsApiAdapter',
          version: '1.0.0',
          generated: true
        }
      };

      console.log('Generated report:', report);
      return report;
    } catch (error) {
      throw this.handleError(error, 'Failed to generate report');
    }
  }

  async exportReport(
    reportId: string,
    format: 'json' | 'csv' | 'html' | 'pdf' = 'json'
  ): Promise<{
    data: string | ArrayBuffer;
    filename: string;
    mimeType: string;
  }> {
    try {
      const report = await this.getReport(reportId);
      
      let data: string;
      let mimeType: string;
      let extension: string;

      switch (format) {
        case 'json':
          data = JSON.stringify(report, null, 2);
          mimeType = 'application/json';
          extension = 'json';
          break;
        case 'csv':
          data = this.convertToCSV(report);
          mimeType = 'text/csv';
          extension = 'csv';
          break;
        case 'html':
          data = this.convertToHTML(report);
          mimeType = 'text/html';
          extension = 'html';
          break;
        case 'pdf':
          // For PDF, we'll generate HTML for now
          data = this.convertToHTML(report);
          mimeType = 'text/html';
          extension = 'html';
          break;
        default:
          throw new Error(`Unsupported export format: ${format}`);
      }

      const filename = `${report.title.replace(/[^a-zA-Z0-9]/g, '_')}_${new Date().toISOString().split('T')[0]}.${extension}`;

      return {
        data,
        filename,
        mimeType
      };
    } catch (error) {
      throw this.handleError(error, 'Failed to export report');
    }
  }

  async getReportsByType(type: string): Promise<GenericReport[]> {
    try {
      const allReports = await this.listReports();
      return allReports.filter(report => report.type === type);
    } catch (error) {
      throw this.handleError(error, 'Failed to get reports by type');
    }
  }

  async getRecentReports(limit: number = 10): Promise<GenericReport[]> {
    try {
      const allReports = await this.listReports();
      return allReports
        .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
        .slice(0, limit);
    } catch (error) {
      throw this.handleError(error, 'Failed to get recent reports');
    }
  }

  // Private helper methods
  private convertToCSV(report: GenericReport): string {
    const lines = [
      'Field,Value',
      `ID,${report.id}`,
      `Type,${report.type}`,
      `Title,${report.title}`,
      `Summary,${report.summary || ''}`,
      `Created,${report.createdAt}`,
      `Updated,${report.updatedAt}`,
      `Status,${report.status}`
    ];

    if (report.data && typeof report.data === 'object') {
      Object.entries(report.data).forEach(([key, value]) => {
        lines.push(`${key},${String(value).replace(/,/g, ';')}`);
      });
    }

    return lines.join('\n');
  }

  private convertToHTML(report: GenericReport): string {
    return `
<!DOCTYPE html>
<html>
<head>
    <title>${report.title}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { border-bottom: 2px solid #ccc; padding-bottom: 10px; margin-bottom: 20px; }
        .content { margin-bottom: 20px; }
        .data { background: #f5f5f5; padding: 10px; border-radius: 4px; }
        .metadata { font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>${report.title}</h1>
        <p><strong>Type:</strong> ${report.type}</p>
        <p><strong>ID:</strong> ${report.id}</p>
        <p><strong>Created:</strong> ${new Date(report.createdAt).toLocaleString()}</p>
        <p><strong>Status:</strong> ${report.status}</p>
    </div>
    
    ${report.summary ? `<div class="content"><h2>Summary</h2><p>${report.summary}</p></div>` : ''}
    
    ${report.data ? `
      <div class="content">
        <h2>Data</h2>
        <div class="data">
          <pre>${JSON.stringify(report.data, null, 2)}</pre>
        </div>
      </div>
    ` : ''}
    
    ${report.metadata ? `
      <div class="metadata">
        <h3>Metadata</h3>
        <pre>${JSON.stringify(report.metadata, null, 2)}</pre>
      </div>
    ` : ''}
</body>
</html>`;
  }

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