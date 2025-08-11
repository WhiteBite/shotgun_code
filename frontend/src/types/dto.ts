import type { GitStatus, ContextOrigin } from './enums';

// DTO for data received from Go backend (raw file tree nodes)
export interface DomainFileNode {
  name: string;
  path: string;
  relPath: string;
  isDir: boolean;
  children?: DomainFileNode[] | null; // Children can be null or empty array from Go
  isGitignored: boolean;
  isCustomIgnored: boolean;
}

// Frontend-specific FileNode, extends DomainFileNode with UI state and computed properties
export interface FileNode {
  path: string;
  name: string;
  relPath: string;
  isDir: boolean;
  children?: { path: string }[]; // Children now reference by path

  depth: number;
  gitStatus: GitStatus;
  contextOrigin: ContextOrigin; // e.g., 'manual', 'git', 'ai', 'dependency'
  isBinary: boolean; // Is it a binary file?
  isIgnored: boolean; // Is it ignored by .gitignore or custom rules?

  // UI state
  expanded: boolean; // Is directory expanded?
  selected: 'on' | 'off' | 'partial'; // Selection state for context building
  parentPath: string | null; // Path to parent node
}

export interface FileStatus {
  path: string;
  status: string;
}

export interface Project {
  id: string;
  name: string;
  path: string;
  gitStatus: string; // Simplified for now, can be more detailed
}

export interface Task {
  id: string;
  name: string;
  date: string;
  contextSummary: string;
}

export interface CommitWithFiles {
  hash: string;
  subject: string;
  files: string[];
  isMerge: boolean;
}

export interface SettingsDTO {
  customIgnoreRules: string;
  customPromptRules: string;
  openAIAPIKey: string;
  geminiAPIKey: string;
  openRouterAPIKey: string;
  localAIAPIKey: string;
  localAIHost: string;
  localAIModelName: string;
  selectedProvider: string;
  selectedModels: Record<string, string>;
  availableModels: Record<string, string[]>;
  useGitignore: boolean;
  useCustomIgnore: boolean;
}

export interface Hunk {
  header: string;
  lines: string[];
}

export interface FileDiff {
  filePath: string;
  hunks: Hunk[];
  stats: { added: number; removed: number };
}

export interface LogEntry {
  message: string;
  type: 'info' | 'warn' | 'error' | 'success';
  timestamp: string;
}