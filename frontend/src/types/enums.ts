export enum TreeMode {
  Navigation = 'navigation',
  Selection = 'selection',
}

export enum GitStatus {
  Clean = 'clean',
  Modified = 'modified',
  Untracked = 'untracked',
  // Staged = 'staged', // Not used yet
  Conflict = 'conflict',
}

export enum ContextOrigin {
  None = 'none',
  // Manual = 'manual', // Not used yet
  // Git = 'git', // Not used yet
  // AI = 'ai', // Not used yet
  // Dependency = 'dependency', // Not used yet
}