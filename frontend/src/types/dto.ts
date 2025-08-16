import type { GitStatus, ContextOrigin } from "./enums";

export interface DomainFileNode {
  name: string; path: string; relPath: string; isDir: boolean;
  children?: DomainFileNode[]; isGitignored: boolean; isCustomIgnored: boolean; size: number;
}
export interface FileNode {
  name: string; path: string; relPath: string; isDir: boolean;
  children?: { path: string }[]; depth: number; gitStatus: GitStatus;
  contextOrigin: ContextOrigin; isBinary: boolean; isIgnored: boolean;
  isGitignored: boolean; isCustomIgnored: boolean; parentPath: string | null;
  size: number;
}
export interface FileStatus { path: string; status: string; }
export interface CommitWithFiles {
  hash: string; subject: string; author: string; date: string; isMerge: boolean; files: string[];
}
export interface SettingsDTO {
  customIgnoreRules: string; customPromptRules: string; openAIAPIKey: string;
  geminiAPIKey: string; openRouterAPIKey: string; localAIAPIKey: string;
  localAIHost: string; localAIModelName: string; selectedProvider: string;
  selectedModels: Record<string, string>; availableModels: Record<string, string[]>;
  useGitignore: boolean; useCustomIgnore: boolean;
}
export interface Hunk { header: string; lines: string[]; }
export interface FileDiff { filePath: string; hunks: Hunk[]; stats: { added: number; removed: number }; }
export type ToastType = "info" | "success" | "error";
export type LogType = ToastType | "warn" | "debug";
export interface LogEntry { id: number; message: string; type: LogType; timestamp: string; }