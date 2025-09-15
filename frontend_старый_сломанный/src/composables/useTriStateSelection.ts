import type { FileNode } from "@/types/dto";
import { defaultSelectionManagementService, type SelectionState } from '@/domain/services/SelectionManagementService';

export type Sel = SelectionState;

/**
 * Composable for tri-state selection functionality
 * Delegates business logic to SelectionManagementService domain service
 * Following DDD principles by separating UI concerns from domain logic
 */
export function useTriStateSelection(
  nodesMap: Map<string, FileNode>,
  selectedPaths: Set<string>,
) {
  const selectionService = defaultSelectionManagementService;

  /**
   * Computes selection state - delegates to domain service
   */
  function computeSelection(node: FileNode): Sel {
    return selectionService.computeSelectionState(node, selectedPaths, nodesMap);
  }

  /**
   * Performs cascade selection - delegates to domain service
   */
  function toggleCascade(node: FileNode) {
    return selectionService.performCascadeSelection(node, selectedPaths, nodesMap);
  }

  /**
   * Counts files in subtree - delegates to domain service
   */
  function countFilesInSubtree(node: FileNode, currentDepth: number) {
    return selectionService.countFilesInSubtree(node, currentDepth, nodesMap);
  }

  /**
   * Validates selection operation - domain service method
   */
  function validateSelection(node: FileNode) {
    return selectionService.validateSelectionOperation(node, nodesMap);
  }

  /**
   * Gets selection statistics - domain service method
   */
  function getSelectionStats(node: FileNode) {
    return selectionService.getSelectionStats(node, selectedPaths, nodesMap);
  }

  /**
   * Optimizes selection set - domain service method
   */
  function optimizeSelection() {
    const optimized = selectionService.optimizeSelectionSet(selectedPaths, nodesMap);
    // Clear current selection and replace with optimized set
    selectedPaths.clear();
    optimized.forEach(path => selectedPaths.add(path));
    return optimized;
  }

  return { 
    computeSelection, 
    toggleCascade, 
    countFilesInSubtree,
    validateSelection,
    getSelectionStats,
    optimizeSelection
  };
}
