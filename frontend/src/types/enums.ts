export enum GitStatus {
  Clean = 'clean',
  Modified = 'modified',
  Untracked = 'untracked',
  Conflict = 'conflict',
}

export enum ContextOrigin {
  None = 'none',
  Manual = 'manual',
  Git = 'git',
  AI = 'ai',
}