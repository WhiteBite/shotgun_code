import { ContextMenuService } from '@/domain/services/ContextMenuService'

export function useContextMenuService() {
  // Create service instance
  const contextMenuService = new ContextMenuService()
  
  return {
    contextMenuService
  }
}