import type { FileNode } from "@/types/dto";

function _updateParentSelection(
  nodesMap: Map<string, FileNode>,
  path: string | null,
) {
  if (path === null) return;
  const parent = nodesMap.get(path);
  if (!parent || !parent.children) return;

  const children = parent.children
    .map((c) => nodesMap.get(c.path)!)
    .filter(Boolean);
  if (children.length === 0) return;

  const selectedCount = children.filter((c) => c.selected === "on").length;
  const partialCount = children.filter((c) => c.selected === "partial").length;

  let newStatus: "on" | "off" | "partial" = "off";
  if (selectedCount === children.length) {
    newStatus = "on";
  } else if (selectedCount > 0 || partialCount > 0) {
    newStatus = "partial";
  }

  if (parent.selected !== newStatus) {
    parent.selected = newStatus;
    _updateParentSelection(nodesMap, parent.parentPath);
  }
}

export function toggleNodeSelection(
  nodesMap: Map<string, FileNode>,
  path: string,
) {
  const node = nodesMap.get(path);
  if (!node) return;

  const newSelection = node.selected === "on" ? "off" : "on";

  const childrenToUpdate = new Map<string, FileNode>();
  function gatherChildren(n: FileNode) {
    childrenToUpdate.set(n.path, n);
    if (n.children) {
      n.children.forEach((c) => {
        const childNode = nodesMap.get(c.path);
        if (childNode) gatherChildren(childNode);
      });
    }
  }
  gatherChildren(node);
  childrenToUpdate.forEach((child) => (child.selected = newSelection));
  _updateParentSelection(nodesMap, node.parentPath);
}
