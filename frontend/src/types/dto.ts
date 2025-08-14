import type { GitStatus, ContextOrigin } from "./enums";

export interface DomainFileNode {
  name: string;
  path: string;
  relPath: string;
  isDir: boolean;
  children?: DomainFileNode[] | null;
  isGitignored: boolean;
  isCustomIgnored: boolean;
}

export interface FileNode {
  path: string;
  name: string;
  relPath: string;
  isDir: boolean;
  children?: { path: string }[];
  depth: number;
  gitStatus: GitStatus;
  contextOrigin: ContextOrigin;
  isBinary: boolean;
  isIgnored: boolean;
  isGitignored: boolean;
  isCustomIgnored: boolean;
  // перешли на tree-state.store как единственный источник
  expanded?: boolean;
  selected?: "on" | "off" | "partial";
  parentPath: string | null;
}

export interface FileStatus {
  path: string;
  status: string;
}

export interface Project {
  id: string;
  name: string;
  path: string;
}

export interface CommitWithFiles {
  hash: string;
  subject: string;
  author: string;
  date: string;
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

export interface LogEntry {
  message: string;
  type: "info" | "warn" | "error" | "success";
  timestamp: string;
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
