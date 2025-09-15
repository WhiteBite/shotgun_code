import { useUiStore } from "@/stores/ui.store";
import { useFileTreeStore } from "@/stores/file-tree.store";
import { useContextBuilderStore } from "@/stores/context-builder.store";
import { useGenerationStore } from "@/stores/generation.store";
import { useTreeStateStore } from "@/stores/tree-state.store";
import { useProjectStore } from "@/stores/project.store";
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useVisibleNodes } from "./useVisibleNodes";
import type { FileNode } from "@/types/dto";

import { 
  attachGlobalNavigationShortcuts, 
  detachGlobalNavigationShortcuts 
} from "./useGlobalNavigationShortcuts";
import { 
  attachWorkspaceModeShortcuts, 
  detachWorkspaceModeShortcuts 
} from "./useWorkspaceModeShortcuts";
import { 
  attachPanelManagementShortcuts, 
  detachPanelManagementShortcuts 
} from "./usePanelManagementShortcuts";
import { 
  attachContextOperationsShortcuts, 
  detachContextOperationsShortcuts 
} from "./useContextOperationsShortcuts";
import { 
  attachGenerationExecutionShortcuts, 
  detachGenerationExecutionShortcuts 
} from "./useGenerationExecutionShortcuts";
import { 
  attachFileNavigationShortcuts, 
  detachFileNavigationShortcuts 
} from "./useFileNavigationShortcuts";

let _attached = false;

export function attachShortcuts() {
  if (_attached) return;

  attachGlobalNavigationShortcuts();
  attachWorkspaceModeShortcuts();
  attachPanelManagementShortcuts();
  attachContextOperationsShortcuts();
  attachGenerationExecutionShortcuts();
  attachFileNavigationShortcuts();

  _attached = true;
}

export function detachShortcuts() {
  if (!_attached) return;

  detachGlobalNavigationShortcuts();
  detachWorkspaceModeShortcuts();
  detachPanelManagementShortcuts();
  detachContextOperationsShortcuts();
  detachGenerationExecutionShortcuts();
  detachFileNavigationShortcuts();

  _attached = false;
}
