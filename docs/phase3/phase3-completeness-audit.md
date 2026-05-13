# Phase 3 Completeness Audit

This audit captures the current gap between the intended Phase 3 product
surface and the implementation in the repository. Phase 3 should not be treated
as complete until the build, evaluate, release, and debug loop is coherent from
the Web Console through the manager API, CRD sync, compiler, and worker runtime.

## Current Verdict

Status: in progress.

Implemented foundations:

- Manager-backed CRUD and sync paths exist for core resources.
- The Web Console has tenant/workspace navigation and list/detail pages for
  agents, evaluations, runs, and providers.
- Agent Studio can select six patterns and persist a workflow graph shape.
- The compiler normalizes Studio workflow node kinds and rejects several
  invalid workflow graph structures.

Not complete:

- The Studio does not provide a compiler-backed validation or preview loop.
- Workflow node references are raw text inputs rather than selections from
  declared models, tools, knowledge bindings, and SubAgents.
- Saving a Studio draft does not create a revision, show compile status, or
  expose compiler errors.
- Existing E2E coverage proves shallow persistence and navigation, not a valid
  compiled workflow or runnable agent.
- Evaluation, release readiness, revision history, provider capability
  comparison, and run debugging remain mostly detail views and static metadata.

## Workflow Studio Gaps

### 1. Validation is not aligned with the compiler

Frontend validation currently checks only basic shape: duplicate names,
Start/End presence, missing edge endpoints, and empty node names. The compiler
also requires:

- `model` nodes to resolve to backend `llm` nodes with `modelRef`.
- `tool` nodes to include `toolRef`.
- `knowledge` nodes to resolve to backend `retrieval` nodes with
  `knowledgeRef`.
- `agent` nodes to include `agentRef`.
- edges that normalize to `START` and `END`.
- all non-terminal nodes to be reachable from `START`.

Required work:

- Extract a shared Studio graph validation module.
- Match compiler errors closely enough that users see the same failure before
  save.
- Add tests for missing `modelRef`, missing `toolRef`, missing `knowledgeRef`,
  missing `agentRef`, no path from Start, no path to End, orphan nodes, duplicate
  names, and missing edge endpoints.

### 2. Node configuration is stringly typed

Workflow node config fields are free text. A usable enterprise Studio should
prefer known bindings:

- model nodes choose from `spec.models`.
- tool nodes choose from `spec.toolRefs`.
- knowledge nodes choose from `spec.knowledgeRefs[].name`.
- agent nodes choose from `spec.subAgentRefs[].name`.
- custom/function nodes choose from available built-in skill functions when
  known, or clearly mark unsupported custom implementations.

Required work:

- Pass current binding options into the workflow canvas.
- Replace raw text-only inputs with select/autocomplete controls plus manual
  escape hatches where needed.
- Show missing binding warnings inline, not only at save time.

### 3. Preview is illustrative, not authoritative

The current preview uses hard-coded diagrams for pattern presets and a local SVG
for workflow graphs. It does not show the runner artifact the compiler will
produce.

Required work:

- Add a manager endpoint or local adapter that returns a compiler-style preview
  for an agent spec without committing it as a published revision.
- Display normalized node kinds, resolved references, pattern expansion
  metadata, selected models/tools/knowledge, and validation errors.
- Keep raw compiled artifact data inspectable for debugging.

### 4. Save semantics are too thin

Saving updates the manager agent spec and navigates back to detail. It does not
communicate whether the synced CRD compiled successfully, whether the revision is
usable, or what changed.

Required work:

- Distinguish draft save, compile validation, publish, and release promotion.
- Show compile status and compiler errors on the Agent detail page.
- Introduce revision history before claiming Phase 3 release readiness.

### 5. E2E coverage is misleading

Existing Studio E2E tests can pass while saving an invalid workflow. They should
cover at least one valid graph that compiles through the backend path.

Required work:

- Add a valid workflow E2E: Start -> Model -> End with a declared model binding.
- Assert the PATCH payload contains the expected spec graph and pattern.
- Assert the manager reopens the saved spec correctly.
- Add a negative E2E for missing required node references.
- Prefer a local Kubernetes smoke path for CRD sync and compiler status before
  calling the Studio flow complete.

## Broader Phase 3 Gaps

### Agents

Missing:

- agent create/edit flow in the console.
- prompt and interface editing.
- revision history.
- publish and release status.
- compile status and compiler error display.

### Evaluations

Missing:

- dataset library UX.
- evaluator and threshold configuration.
- current versus baseline comparison.
- metric breakdown and regression history.
- release gate decision surface.

### Runs

Missing:

- structured output inspection.
- artifact and trace browsing beyond references.
- policy outcomes.
- runtime logs and failure reason workflow.
- navigation from a failed run back to the exact agent revision/spec.

### Providers

Missing:

- model catalog.
- capability matrix comparison.
- workspace-scoped provider binding editor.
- credential-reference management UX that avoids exposing secret values.
- cost, latency, and provider support metadata.

### Governance and Administration

Missing:

- approval queues.
- release gates and waivers.
- audit history.
- members, roles, quotas, runtime backend configuration, and global provider
  settings.

## Suggested Phase 3 Completion Milestones

1. Workflow Studio correctness:
   compiler-aligned validation, reference-aware node config, and valid workflow
   E2E coverage.
2. Agent draft and revision loop:
   save draft, compile preview, compile status, revision history, and publish
   action.
3. Evaluation decision loop:
   dataset selection, thresholds, baseline comparison, metric breakdown, and
   release gate status.
4. Run debugging loop:
   structured output, trace/artifact references, failure reasons, and links back
   to the exact agent revision.
5. Provider management loop:
   capability matrix, workspace bindings, credential references, and model
   selection.

Phase 3 can be called complete only after these flows are covered by focused
unit/integration tests plus at least one end-to-end build/evaluate/release smoke
path.
