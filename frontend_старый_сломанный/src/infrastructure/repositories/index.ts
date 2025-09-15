// Infrastructure API Adapters - Clean Architecture implementation
// These adapters implement the repository interfaces and handle communication with the backend

export { ProjectApiAdapter } from './ProjectApiAdapter';
export { ContextApiAdapter } from './ContextApiAdapter';
export { GitApiAdapter } from './GitApiAdapter';
export { AIApiAdapter } from './AIApiAdapter';
export { SettingsApiAdapter } from './SettingsApiAdapter';
export { 
  ReportsApiAdapter, 
  AutonomousApiAdapter, 
  ExportApiAdapter 
} from './AdditionalApiAdapters';