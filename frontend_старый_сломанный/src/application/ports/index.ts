// Application layer port interfaces for Clean Architecture
// These define contracts that infrastructure layer must implement

export interface ApiPort {
  get<T>(endpoint: string, params?: Record<string, unknown>): Promise<T>;
  post<T>(endpoint: string, data?: unknown): Promise<T>;
  put<T>(endpoint: string, data?: unknown): Promise<T>;
  delete<T>(endpoint: string): Promise<T>;
}

export interface FileSystemPort {
  readFile(path: string): Promise<string>;
  writeFile(path: string, content: string): Promise<void>;
  exists(path: string): Promise<boolean>;
  listDirectory(path: string): Promise<string[]>;
}

export interface StoragePort {
  getItem(key: string): Promise<string | null>;
  setItem(key: string, value: string): Promise<void>;
  removeItem(key: string): Promise<void>;
  clear(): Promise<void>;
}

export interface NotificationPort {
  showSuccess(message: string): void;
  showError(message: string): void;
  showWarning(message: string): void;
  showInfo(message: string): void;
}

export interface DialogPort {
  showOpenDialog(options: unknown): Promise<string | null>;
  showSaveDialog(options: unknown): Promise<string | null>;
  showMessageBox(options: unknown): Promise<unknown>;
}