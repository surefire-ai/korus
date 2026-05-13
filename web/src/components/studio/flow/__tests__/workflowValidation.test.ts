import { describe, it, expect } from "vitest";
import type { Node, Edge } from "@xyflow/react";
import type { WorkflowNodeData } from "../WorkflowNode";
import { validateGraph } from "../workflowValidation";

function makeNode(id: string, data: Partial<WorkflowNodeData> & { label: string; kind: WorkflowNodeData["kind"] }): Node<WorkflowNodeData> {
  return { id, type: "workflow", position: { x: 0, y: 0 }, data: data as WorkflowNodeData };
}

function makeEdge(source: string, target: string): Edge {
  return { id: `${source}-${target}`, source, target };
}

describe("validateGraph", () => {
  it("returns no errors for a valid workflow", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("llm", { label: "Generate", kind: "model", modelRef: "planner" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "llm"), makeEdge("llm", "end")];
    expect(validateGraph(nodes, edges)).toEqual([]);
  });

  it("reports missing Start node", () => {
    const nodes = [makeNode("end", { label: "End", kind: "end" })];
    expect(validateGraph(nodes, [])).toContain("Workflow should have a Start node");
  });

  it("reports missing End node", () => {
    const nodes = [makeNode("start", { label: "Start", kind: "start" })];
    expect(validateGraph(nodes, [])).toContain("Workflow should have an End node");
  });

  it("reports duplicate node names", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("a", { label: "Step", kind: "model", modelRef: "planner" }),
      makeNode("b", { label: "Step", kind: "model", modelRef: "planner" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "a"), makeEdge("a", "end"), makeEdge("start", "b"), makeEdge("b", "end")];
    expect(validateGraph(nodes, edges)).toContain('Duplicate node name: "Step"');
  });

  it("reports edge referencing missing source", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("ghost", "end")];
    expect(validateGraph(nodes, edges)).toContain('Edge references missing source: "ghost"');
  });

  it("reports edge referencing missing target", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "ghost")];
    expect(validateGraph(nodes, edges)).toContain('Edge references missing target: "ghost"');
  });

  it("reports empty labels on non-terminal nodes", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("llm", { label: "", kind: "model", modelRef: "planner" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "llm"), makeEdge("llm", "end")];
    expect(validateGraph(nodes, edges)).toContain("1 node(s) have no name");
  });

  // ── Required field checks (compiler alignment) ──

  it("reports missing modelRef on model node", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("llm", { label: "Generate", kind: "model" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "llm"), makeEdge("llm", "end")];
    expect(validateGraph(nodes, edges)).toContain('Model node "Generate" requires a model reference');
  });

  it("reports missing toolRef on tool node", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("tool", { label: "Search", kind: "tool" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "tool"), makeEdge("tool", "end")];
    expect(validateGraph(nodes, edges)).toContain('Tool node "Search" requires a tool reference');
  });

  it("reports missing knowledgeRef on knowledge node", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("kb", { label: "Docs", kind: "knowledge" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "kb"), makeEdge("kb", "end")];
    expect(validateGraph(nodes, edges)).toContain('Knowledge node "Docs" requires a knowledge reference');
  });

  it("reports missing agentRef on agent node", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("agent", { label: "SubAgent", kind: "agent" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "agent"), makeEdge("agent", "end")];
    expect(validateGraph(nodes, edges)).toContain('Agent node "SubAgent" requires an agent reference');
  });

  it("accepts model node with modelRef set", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("llm", { label: "Generate", kind: "model", modelRef: "planner" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "llm"), makeEdge("llm", "end")];
    const errors = validateGraph(nodes, edges);
    expect(errors.filter((e) => e.includes("model reference"))).toEqual([]);
  });

  // ── Non-terminal node count ──

  it("reports when only Start and End exist with no processing nodes", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "end")];
    expect(validateGraph(nodes, edges)).toContain("Workflow needs at least one processing node between Start and End");
  });

  // ── Reachability ──

  it("reports orphan node not reachable from Start", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("a", { label: "Connected", kind: "model", modelRef: "planner" }),
      makeNode("b", { label: "Orphan", kind: "model", modelRef: "planner" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "a"), makeEdge("a", "end")];
    expect(validateGraph(nodes, edges)).toContain('Node "Orphan" is not reachable from Start');
  });

  it("does not report reachability errors when all nodes are connected", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("a", { label: "Step A", kind: "model", modelRef: "planner" }),
      makeNode("b", { label: "Step B", kind: "tool", toolRef: "search" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "a"), makeEdge("a", "b"), makeEdge("b", "end")];
    const errors = validateGraph(nodes, edges);
    expect(errors.filter((e) => e.includes("not reachable"))).toEqual([]);
  });

  it("handles multiple required field errors at once", () => {
    const nodes = [
      makeNode("start", { label: "Start", kind: "start" }),
      makeNode("m", { label: "Model", kind: "model" }),
      makeNode("t", { label: "Tool", kind: "tool" }),
      makeNode("end", { label: "End", kind: "end" }),
    ];
    const edges = [makeEdge("start", "m"), makeEdge("m", "t"), makeEdge("t", "end")];
    const errors = validateGraph(nodes, edges);
    expect(errors).toContain('Model node "Model" requires a model reference');
    expect(errors).toContain('Tool node "Tool" requires a tool reference');
  });
});
