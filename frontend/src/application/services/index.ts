// Application services coordinate between use-cases and external dependencies
import type { ApiPort, NotificationPort, DialogPort } from '../ports';

export class ProjectApplicationService {
  constructor(
    private apiPort: ApiPort,
    private notificationPort: NotificationPort,
    private dialogPort: DialogPort
  ) {}

  async openProject(): Promise<boolean> {
    try {
      const result = await this.dialogPort.showOpenDialog({
        properties: ['openDirectory'],
        title: 'Select Project Directory'
      });

      if (result) {
        await this.apiPort.post('/api/project/open', { path: result });
        this.notificationPort.showSuccess('Project opened successfully');
        return true;
      }
      return false;
    } catch (error) {
      this.notificationPort.showError(`Failed to open project: ${error}`);
      return false;
    }
  }

  async loadProject(path: string): Promise<boolean> {
    try {
      await this.apiPort.post('/api/project/load', { path });
      return true;
    } catch (error) {
      this.notificationPort.showError(`Failed to load project: ${error}`);
      return false;
    }
  }
}

export class ContextApplicationService {
  constructor(
    private apiPort: ApiPort,
    private notificationPort: NotificationPort
  ) {}

  async buildContext(files: string[]): Promise<string | null> {
    try {
      const result = await this.apiPort.post<{ content: string }>('/api/context/build', { files });
      this.notificationPort.showSuccess('Context built successfully');
      return result.content;
    } catch (error) {
      this.notificationPort.showError(`Failed to build context: ${error}`);
      return null;
    }
  }

  async streamContext(files: string[]): Promise<ReadableStream | null> {
    try {
      // This would return a streaming response
      const result = await this.apiPort.post<ReadableStream>('/api/context/stream', { files });
      return result;
    } catch (error) {
      this.notificationPort.showError(`Failed to stream context: ${error}`);
      return null;
    }
  }
}

export class AIApplicationService {
  constructor(
    private apiPort: ApiPort,
    private notificationPort: NotificationPort
  ) {}

  async generateCode(prompt: string, context: string): Promise<string | null> {
    try {
      const result = await this.apiPort.post<{ code: string }>('/api/ai/generate', { 
        prompt, 
        context 
      });
      this.notificationPort.showSuccess('Code generated successfully');
      return result.code;
    } catch (error) {
      this.notificationPort.showError(`Failed to generate code: ${error}`);
      return null;
    }
  }

  async validateCode(code: string): Promise<boolean> {
    try {
      await this.apiPort.post('/api/ai/validate', { code });
      return true;
    } catch (error) {
      this.notificationPort.showError(`Code validation failed: ${error}`);
      return false;
    }
  }
}