// Infrastructure layer API adapter implementing ApiPort interface
import type { ApiPort } from '../../application/ports';

export class WailsApiAdapter implements ApiPort {
  async get<T>(endpoint: string, params?: Record<string, any>): Promise<T> {
    // Implementation using Wails runtime
    // This would call the appropriate Wails Go backend method
    const result = await window.go.main.App.ApiGet(endpoint, params || {});
    return result as T;
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    // Implementation using Wails runtime
    const result = await window.go.main.App.ApiPost(endpoint, data || {});
    return result as T;
  }

  async put<T>(endpoint: string, data?: any): Promise<T> {
    // Implementation using Wails runtime
    const result = await window.go.main.App.ApiPut(endpoint, data || {});
    return result as T;
  }

  async delete<T>(endpoint: string): Promise<T> {
    // Implementation using Wails runtime
    const result = await window.go.main.App.ApiDelete(endpoint);
    return result as T;
  }
}

// Type definitions for Wails runtime
declare global {
  interface Window {
    go: {
      main: {
        App: {
          ApiGet(endpoint: string, params: Record<string, any>): Promise<any>;
          ApiPost(endpoint: string, data: any): Promise<any>;
          ApiPut(endpoint: string, data: any): Promise<any>;
          ApiDelete(endpoint: string): Promise<any>;
        };
      };
    };
  }
}