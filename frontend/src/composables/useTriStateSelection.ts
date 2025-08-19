import type { FileNode } from "@/types/dto";

export type Sel = "on" | "off" | "partial";

export function useTriStateSelection(
  nodesMap: Map<string, FileNode>,
  selectedPaths: Set<string>
) {
  function computeSelection(node: FileNode): Sel {
    if (!node.isDir) return selectedPaths.has(node.path) ? "on" : "off";
    const children = (node.children || [])
      .map(c => nodesMap.get(c.path))
      .filter(Boolean) as FileNode[];
    if (children.length === 0) return selectedPaths.has(node.path) ? "on" : "off";
    let hasOn = false, hasOff = false;
    for (const ch of children) {
      if (ch.isIgnored) continue;
      const st = computeSelection(ch);
      if (st === "partial") return "partial";
      if (st === "on") hasOn = true;
      if (st === "off") hasOff = true;
      if (hasOn && hasOff) return "partial";
    }
    if (hasOn && !hasOff) return "on";
    if (!hasOn && hasOff) return "off";
    return "off";
  }

  function toggleCascade(node: FileNode) {
    const select = !selectedPaths.has(node.path);
    const stack: FileNode[] = [node];
    while (stack.length) {
      const cur = stack.pop()!;
      if (!cur.isIgnored) {
        if (select) selectedPaths.add(cur.path);
        else selectedPaths.delete(cur.path);
        if (cur.isDir && cur.children) {
          for (const c of cur.children) {
            const n = nodesMap.get(c.path);
            if (n) stack.push(n);
          }
        }
      }
    }
  }

  return { computeSelection, toggleCascade };
}
