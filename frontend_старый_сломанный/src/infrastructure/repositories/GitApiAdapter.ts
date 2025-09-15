import type { GitRepository } from '@/domain/repositories/GitRepository';
import type { FileStatus, CommitWithFiles } from '@/types/dto';
import { 
  IsGitAvailable,
  GetUncommittedFiles,
  GetRichCommitHistory,
  GetFileContentAtCommit,
  GetGitignoreContent,
  GetBranches,
  GetCurrentBranch
} from '../../../wailsjs/go/main/App';
import { defaultWailsApiAdapter, type WailsApiAdapter } from '../api/WailsApiAdapter';

/**
 * Git API Adapter - Infrastructure implementation of GitRepository
 * This handles git operations while conforming to Clean Architecture
 */
export class GitApiAdapter implements GitRepository {
  private readonly apiAdapter: WailsApiAdapter;

  constructor(apiAdapter: WailsApiAdapter = defaultWailsApiAdapter) {
    this.apiAdapter = apiAdapter;
  }
  async isGitAvailable(): Promise<boolean> {
    try {
      return await IsGitAvailable();
    } catch (error) {
      console.warn('Git availability check failed:', error);
      return false;
    }
  }

  async getUncommittedFiles(projectRoot: string): Promise<FileStatus[]> {
    try {
      return await GetUncommittedFiles(projectRoot);
    } catch (error) {
      throw this.handleError(error, 'Failed to get uncommitted files');
    }
  }

  async getRichCommitHistory(
    projectRoot: string,
    branchName: string,
    limit: number
  ): Promise<CommitWithFiles[]> {
    try {
      return await GetRichCommitHistory(projectRoot, branchName, limit);
    } catch (error) {
      throw this.handleError(error, 'Failed to get commit history');
    }
  }

  async getFileContentAtCommit(
    projectRoot: string,
    filePath: string,
    commitHash: string
  ): Promise<string> {
    try {
      return await GetFileContentAtCommit(projectRoot, filePath, commitHash);
    } catch (error) {
      throw this.handleError(error, 'Failed to get file content at commit');
    }
  }

  async getGitignoreContent(projectRoot: string): Promise<string> {
    try {
      return await GetGitignoreContent(projectRoot);
    } catch (error) {
      throw this.handleError(error, 'Failed to get gitignore content');
    }
  }

  async getCurrentBranch(projectRoot: string): Promise<string> {
    return this.apiAdapter.callApi<string, string>(GetCurrentBranch, projectRoot);
  }

  async getBranches(projectRoot: string): Promise<string[]> {
    return this.apiAdapter.callApi<string, string[]>(GetBranches, projectRoot);
  }

  async hasUncommittedChanges(projectRoot: string): Promise<boolean> {
    try {
      const uncommittedFiles = await this.getUncommittedFiles(projectRoot);
      return uncommittedFiles.length > 0;
    } catch (error) {
      // If we can't check git status, assume no changes
      console.warn('Failed to check uncommitted changes:', error);
      return false;
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