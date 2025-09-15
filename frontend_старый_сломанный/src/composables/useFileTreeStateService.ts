import { FileTreeStateService } from '@/domain/services/FileTreeStateService'

export function useFileTreeStateService() {
  // Create service instance
  const fileTreeStateService = new FileTreeStateService()
  
  return {
    fileTreeStateService
  }
}