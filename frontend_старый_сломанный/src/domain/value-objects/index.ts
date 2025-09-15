/**
 * Domain Value Objects
 * Immutable objects that represent domain concepts without identity
 */

export class ProjectId {
  private readonly _value: string;

  constructor(value: string) {
    if (!value || value.trim().length === 0) {
      throw new Error('ProjectId cannot be empty');
    }
    this._value = value.trim();
  }

  get value(): string {
    return this._value;
  }

  equals(other: ProjectId): boolean {
    return this._value === other._value;
  }

  toString(): string {
    return this._value;
  }
}

export class ProjectPath {
  private readonly _value: string;

  constructor(value: string) {
    if (!value || value.trim().length === 0) {
      throw new Error('ProjectPath cannot be empty');
    }
    
    // Normalize path separators for Windows
    const normalized = value.replace(/\\/g, '/').trim();
    this._value = normalized;
  }

  get value(): string {
    return this._value;
  }

  get name(): string {
    return this._value.split('/').pop() || '';
  }

  equals(other: ProjectPath): boolean {
    return this._value === other._value;
  }

  toString(): string {
    return this._value;
  }
}

export class NodePath {
  private readonly _value: string;

  constructor(value: string) {
    if (!value || value.trim().length === 0) {
      throw new Error('NodePath cannot be empty');
    }
    
    // Normalize path separators
    const normalized = value.replace(/\\/g, '/').trim();
    this._value = normalized;
  }

  get value(): string {
    return this._value;
  }

  get fileName(): string {
    return this._value.split('/').pop() || '';
  }

  get directory(): string {
    const parts = this._value.split('/');
    return parts.slice(0, -1).join('/');
  }

  get extension(): string {
    const fileName = this.fileName;
    const lastDot = fileName.lastIndexOf('.');
    return lastDot === -1 ? '' : fileName.substring(lastDot + 1);
  }

  isChildOf(parent: NodePath): boolean {
    return this._value.startsWith(parent._value + '/');
  }

  equals(other: NodePath): boolean {
    return this._value === other._value;
  }

  toString(): string {
    return this._value;
  }
}

export class FileSize {
  private readonly _bytes: number;

  constructor(bytes: number) {
    if (bytes < 0) {
      throw new Error('FileSize cannot be negative');
    }
    this._bytes = bytes;
  }

  get bytes(): number {
    return this._bytes;
  }

  get humanReadable(): string {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = this._bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`;
  }

  equals(other: FileSize): boolean {
    return this._bytes === other._bytes;
  }

  isLargerThan(other: FileSize): boolean {
    return this._bytes > other._bytes;
  }
}

export class TokenCount {
  private readonly _count: number;

  constructor(count: number) {
    if (count < 0) {
      throw new Error('TokenCount cannot be negative');
    }
    this._count = Math.floor(count);
  }

  get count(): number {
    return this._count;
  }

  get humanReadable(): string {
    if (this._count < 1000) {
      return this._count.toString();
    }
    
    if (this._count < 1000000) {
      return `${(this._count / 1000).toFixed(1)}K`;
    }
    
    return `${(this._count / 1000000).toFixed(1)}M`;
  }

  add(other: TokenCount): TokenCount {
    return new TokenCount(this._count + other._count);
  }

  equals(other: TokenCount): boolean {
    return this._count === other._count;
  }
}

export enum NodeType {
  FILE = 'file',
  DIRECTORY = 'directory'
}

export enum SelectionState {
  UNSELECTED = 'unselected',
  SELECTED = 'selected',
  PARTIALLY_SELECTED = 'partially_selected'
}

export enum WorkspaceMode {
  FILE_EXPLORER = 'file_explorer',
  CONTEXT_BUILDER = 'context_builder',
  CODE_GENERATOR = 'code_generator',
  TASK_COMPOSER = 'task_composer'
}

export class ContextSummary {
  constructor(
    public readonly fileCount: number,
    public readonly estimatedTokens: TokenCount,
    public readonly totalSize: FileSize,
    public readonly selectedPaths: readonly string[]
  ) {}

  get isEmpty(): boolean {
    return this.fileCount === 0;
  }

  get isValid(): boolean {
    return this.fileCount > 0 && this.estimatedTokens.count > 0;
  }
}