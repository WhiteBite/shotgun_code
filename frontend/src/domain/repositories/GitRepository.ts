import type { FileStatus, CommitWithFiles } from '@/types/dto';

/**
 * Repository interface for Git operations
 */
export interface GitRepository {
  /**
   * Check if Git is available in the system
   * @returns True if Git is available
   */
  isGitAvailable(): Promise<boolean>;

  /**
   * Get uncommitted files in the project
   * @param projectRoot Project root path
   * @returns Array of file statuses
   */
  getUncommittedFiles(projectRoot: string): Promise<FileStatus[]>;

  /**
   * Get rich commit history with file changes
   * @param projectRoot Project root path
   * @param branchName Branch name
   * @param limit Number of commits to fetch
   * @returns Array of commits with files
   */
  getRichCommitHistory(
    projectRoot: string,
    branchName: string,
    limit: number
  ): Promise<CommitWithFiles[]>;

  /**
   * Get file content at a specific commit
   * @param projectRoot Project root path
   * @param filePath File path
   * @param commitHash Commit hash
   * @returns File content at commit
   */
  getFileContentAtCommit(
    projectRoot: string,
    filePath: string,
    commitHash: string
  ): Promise<string>;

  /**
   * Get gitignore content
   * @param projectRoot Project root path
   * @returns Gitignore content
   */
  getGitignoreContent(projectRoot: string): Promise<string>;

  /**
   * Get current branch name
   * @param projectRoot Project root path
   * @returns Current branch name
   */
  getCurrentBranch(projectRoot: string): Promise<string>;

  /**
   * Get list of all branches
   * @param projectRoot Project root path
   * @returns Array of branch names
   */
  getBranches(projectRoot: string): Promise<string[]>;

  /**
   * Check if repository has uncommitted changes
   * @param projectRoot Project root path
   * @returns True if there are uncommitted changes
   */
  hasUncommittedChanges(projectRoot: string): Promise<boolean>;
}