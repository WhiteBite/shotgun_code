// Infrastructure API module exports
export { WailsApiAdapter } from './WailsApiAdapter';
export { ToastNotificationAdapter, WailsDialogAdapter } from './adapters';

// Re-export the original API service for backward compatibility during migration
export { default as ApiService } from './api.service';