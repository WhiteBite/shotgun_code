import type { GitRepository } from '@/domain/repositories/GitRepository';
import type { GitCommit, GitBranch, GitStatus } from '@/types/api';

/**
 * Git Branch Management Use Cases
 */

export class GetGitBranchesUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string): Promise<GitBranch[]> {
    this.validateProjectPath(projectPath);
    
    try {
      return await this.gitRepo.getBranches(projectPath);
    } catch (error) {
      throw new Error(`Failed to get git branches: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}

export class SwitchGitBranchUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string, branchName: string): Promise<void> {
    this.validateInputs(projectPath, branchName);
    
    try {
      await this.gitRepo.switchBranch(projectPath, branchName);
    } catch (error) {
      throw new Error(`Failed to switch to branch '${branchName}': ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateInputs(projectPath: string, branchName: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    if (!branchName?.trim()) {
      throw new Error('Branch name is required');
    }
  }
}

export class CreateGitBranchUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string, branchName: string, fromBranch?: string): Promise<void> {
    this.validateInputs(projectPath, branchName);
    
    try {
      await this.gitRepo.createBranch(projectPath, branchName, fromBranch);
    } catch (error) {
      throw new Error(`Failed to create branch '${branchName}': ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateInputs(projectPath: string, branchName: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    if (!branchName?.trim()) {
      throw new Error('Branch name is required');
    }
    if (!/^[a-zA-Z0-9/_-]+$/.test(branchName)) {
      throw new Error('Branch name contains invalid characters');
    }
  }
}

/**
 * Git History and Commit Use Cases
 */

export class GetGitCommitsUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string, limit?: number, branch?: string): Promise<GitCommit[]> {
    this.validateProjectPath(projectPath);
    
    try {
      return await this.gitRepo.getCommits(projectPath, limit, branch);
    } catch (error) {
      throw new Error(`Failed to get git commits: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}

export class GetGitCommitDiffUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string, commitHash: string): Promise<string> {
    this.validateInputs(projectPath, commitHash);
    
    try {
      return await this.gitRepo.getCommitDiff(projectPath, commitHash);
    } catch (error) {
      throw new Error(`Failed to get commit diff: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateInputs(projectPath: string, commitHash: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
    if (!commitHash?.trim()) {
      throw new Error('Commit hash is required');
    }
    if (!/^[a-f0-9]{7,40}$/i.test(commitHash)) {
      throw new Error('Invalid commit hash format');
    }
  }
}

/**
 * Git Status and Changes Use Cases
 */

export class GetGitStatusUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string): Promise<GitStatus> {
    this.validateProjectPath(projectPath);
    
    try {
      return await this.gitRepo.getStatus(projectPath);
    } catch (error) {
      throw new Error(`Failed to get git status: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}

export class GetGitChangedFilesUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string, since?: string): Promise<string[]> {
    this.validateProjectPath(projectPath);
    
    try {
      return await this.gitRepo.getChangedFiles(projectPath, since);
    } catch (error) {
      throw new Error(`Failed to get changed files: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}

/**
 * Git Repository Validation Use Cases
 */

export class ValidateGitRepositoryUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string): Promise<{
    isValid: boolean;
    isGitRepo: boolean;
    hasRemote: boolean;
    currentBranch?: string;
    remoteUrl?: string;
    errors: string[];
    warnings: string[];
  }> {
    const result = {
      isValid: false,
      isGitRepo: false,
      hasRemote: false,
      currentBranch: undefined as string | undefined,
      remoteUrl: undefined as string | undefined,
      errors: [] as string[],
      warnings: [] as string[]
    };

    try {
      this.validateProjectPath(projectPath);
      
      // Check if it's a git repository
      result.isGitRepo = await this.gitRepo.isGitRepository(projectPath);
      
      if (!result.isGitRepo) {
        result.errors.push('Not a git repository');
        return result;
      }

      // Get current branch
      try {
        const branches = await this.gitRepo.getBranches(projectPath);
        const currentBranch = branches.find(b => b.isCurrent);
        result.currentBranch = currentBranch?.name;
      } catch (error) {
        result.warnings.push('Could not determine current branch');
      }

      // Check for remote
      try {
        result.hasRemote = await this.gitRepo.hasRemote(projectPath);
        if (result.hasRemote) {
          result.remoteUrl = await this.gitRepo.getRemoteUrl(projectPath);
        }
      } catch (error) {
        result.warnings.push('Could not check remote configuration');
      }

      result.isValid = result.isGitRepo && result.errors.length === 0;
      
    } catch (error) {
      result.errors.push(error instanceof Error ? error.message : 'Unknown validation error');
    }

    return result;
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}

/**
 * Composite Use Cases
 */

export class GetGitRepositoryInfoUseCase {
  constructor(private gitRepo: GitRepository) {}

  async execute(projectPath: string): Promise<{
    isGitRepo: boolean;
    currentBranch?: string;
    branches: GitBranch[];
    recentCommits: GitCommit[];
    status: GitStatus;
    hasUncommittedChanges: boolean;
    remoteInfo?: {
      hasRemote: boolean;
      url?: string;
    };
  }> {
    this.validateProjectPath(projectPath);

    const info = {
      isGitRepo: false,
      currentBranch: undefined as string | undefined,
      branches: [] as GitBranch[],
      recentCommits: [] as GitCommit[],
      status: {} as GitStatus,
      hasUncommittedChanges: false,
      remoteInfo: undefined as { hasRemote: boolean; url?: string } | undefined
    };

    try {
      // Check if git repository
      info.isGitRepo = await this.gitRepo.isGitRepository(projectPath);
      
      if (!info.isGitRepo) {
        return info;
      }

      // Get all git information in parallel for better performance
      const [branches, status, commits] = await Promise.allSettled([
        this.gitRepo.getBranches(projectPath),
        this.gitRepo.getStatus(projectPath),
        this.gitRepo.getCommits(projectPath, 10)
      ]);

      if (branches.status === 'fulfilled') {
        info.branches = branches.value;
        info.currentBranch = branches.value.find(b => b.isCurrent)?.name;
      }

      if (status.status === 'fulfilled') {
        info.status = status.value;
        info.hasUncommittedChanges = status.value.hasChanges || false;
      }

      if (commits.status === 'fulfilled') {
        info.recentCommits = commits.value;
      }

      // Get remote info
      try {
        const hasRemote = await this.gitRepo.hasRemote(projectPath);
        info.remoteInfo = { hasRemote };
        
        if (hasRemote) {
          const remoteUrl = await this.gitRepo.getRemoteUrl(projectPath);
          info.remoteInfo.url = remoteUrl;
        }
      } catch (error) {
        console.warn('Failed to get remote info:', error);
        info.remoteInfo = { hasRemote: false };
      }

    } catch (error) {
      console.error('Failed to get git repository info:', error);
      throw new Error(`Failed to get git repository info: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }

    return info;
  }

  private validateProjectPath(projectPath: string): void {
    if (!projectPath?.trim()) {
      throw new Error('Project path is required');
    }
  }
}