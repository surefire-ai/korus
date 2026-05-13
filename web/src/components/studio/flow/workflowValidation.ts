import type { Node, Edge } from "@xyflow/react";
import type { WorkflowNodeData } from "./WorkflowNode";

export function validateGraph(nodes: Node<WorkflowNodeData>[], edges: Edge[]): string[] {
  const errors: string[] = [];
  const names = nodes.map((n) => n.data.label).filter(Boolean);
  const nameSet = new Set<string>();

  // Duplicate names
  for (const name of names) {
    if (nameSet.has(name)) {
      errors.push(`Duplicate node name: "${name}"`);
    }
    nameSet.add(name);
  }

  // Start/End check
  const hasStart = nodes.some((n) => n.data.kind === "start");
  const hasEnd = nodes.some((n) => n.data.kind === "end");
  if (!hasStart) errors.push("Workflow should have a Start node");
  if (!hasEnd) errors.push("Workflow should have an End node");

  // Edge references
  const nodeIds = new Set(nodes.map((n) => n.id));
  for (const e of edges) {
    if (!nodeIds.has(e.source)) errors.push(`Edge references missing source: "${e.source}"`);
    if (!nodeIds.has(e.target)) errors.push(`Edge references missing target: "${e.target}"`);
  }

  // Empty labels
  const unlabeled = nodes.filter((n) => !n.data.label && n.data.kind !== "start" && n.data.kind !== "end");
  if (unlabeled.length > 0) {
    errors.push(`${unlabeled.length} node(s) have no name`);
  }

  // Required fields per kind (mirrors compiler validateWorkflowGraphStruct)
  const nonTerminal = nodes.filter((n) => n.data.kind !== "start" && n.data.kind !== "end");
  for (const node of nonTerminal) {
    const label = node.data.label || node.id;
    switch (node.data.kind) {
      case "model":
        if (!node.data.modelRef?.trim()) {
          errors.push(`Model node "${label}" requires a model reference`);
        }
        break;
      case "tool":
        if (!node.data.toolRef?.trim()) {
          errors.push(`Tool node "${label}" requires a tool reference`);
        }
        break;
      case "knowledge":
        if (!node.data.knowledgeRef?.trim()) {
          errors.push(`Knowledge node "${label}" requires a knowledge reference`);
        }
        break;
      case "agent":
        if (!node.data.agentRef?.trim()) {
          errors.push(`Agent node "${label}" requires an agent reference`);
        }
        break;
    }
  }

  // At least one non-terminal node between Start and End
  if (hasStart && hasEnd && nonTerminal.length === 0) {
    errors.push("Workflow needs at least one processing node between Start and End");
  }

  // Reachability from Start
  if (hasStart && nonTerminal.length > 0) {
    const startNode = nodes.find((n) => n.data.kind === "start");
    if (startNode) {
      const adjacency = new Map<string, string[]>();
      for (const e of edges) {
        if (!adjacency.has(e.source)) adjacency.set(e.source, []);
        adjacency.get(e.source)!.push(e.target);
      }
      const reachable = new Set<string>();
      const queue = [startNode.id];
      while (queue.length > 0) {
        const current = queue.shift()!;
        if (reachable.has(current)) continue;
        reachable.add(current);
        for (const neighbor of adjacency.get(current) ?? []) {
          queue.push(neighbor);
        }
      }
      for (const node of nonTerminal) {
        if (!reachable.has(node.id)) {
          errors.push(`Node "${node.data.label || node.id}" is not reachable from Start`);
        }
      }
    }
  }

  return errors;
}
