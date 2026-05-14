import { test, expect } from "@playwright/test";
import { api, cleanupAgent, seedAgent } from "./fixtures/api";

test.describe("Workflow Validation", () => {
  test("valid workflow compiles and saves with persisted spec", async ({ page }, testInfo) => {
    const agentId = `agent_wf_valid_${testInfo.workerIndex}_${Date.now()}`;

    await seedAgent({
      id: agentId,
      tenantId: "t_enterprise",
      workspaceId: "ws_enterprise",
      slug: agentId.replaceAll("_", "-"),
      displayName: "Workflow Valid Agent",
      modelProvider: "openai",
      modelName: "gpt-4.1-mini",
    });

    // PATCH agent with a complete workflow spec (nodes + edges + model binding)
    await api.patch(`/agents/${agentId}`, {
      pattern: "workflow",
      spec: {
        pattern: { type: "workflow" },
        models: {
          planner: { provider: "openai", model: "gpt-4o" },
        },
        toolRefs: [],
        knowledgeRefs: [],
        skillRefs: [],
        subAgentRefs: [],
        mcpRefs: [],
        graph: {
          nodes: [
            { name: "start", kind: "start" },
            { name: "planner", kind: "model", modelRef: "planner" },
            { name: "end", kind: "end" },
          ],
          edges: [
            { from: "start", to: "planner" },
            { from: "planner", to: "end" },
          ],
        },
      },
    });

    try {
      // Open studio — spec should load from saved agent
      await page.goto(`/tenants/t_enterprise/agents/${agentId}/studio`);
      await expect(page.getByRole("heading", { name: "Workflow Valid Agent" })).toBeVisible();
      await expect(page.getByText(/Orchestration Studio · workflow/)).toBeVisible();

      // Verify the workflow graph loaded with 3 nodes
      await expect(page.getByText(/3 Nodes · 2 Edges/)).toBeVisible();

      // Switch to Preview tab and run compile
      await page.getByRole("button", { name: "Preview" }).click();
      await page.getByRole("button", { name: "Run Compile" }).click();
      await expect(page.getByText("Compilation successful")).toBeVisible({ timeout: 10_000 });

      // Save and verify redirect to detail page
      await page.getByRole("button", { name: "Save" }).click();
      await expect(page).toHaveURL(new RegExp(`/tenants/t_enterprise/agents/${agentId}$`));

      // Reopen studio and verify spec persisted
      await page.getByRole("button", { name: "Open Studio" }).click();
      await expect(page).toHaveURL(new RegExp(`/tenants/t_enterprise/agents/${agentId}/studio$`));
      await expect(page.getByText(/Orchestration Studio · workflow/)).toBeVisible();
      await expect(page.getByText(/3 Nodes · 2 Edges/)).toBeVisible();
    } finally {
      await cleanupAgent(agentId);
    }
  });

  test("model node without modelRef shows validation warning", async ({ page }, testInfo) => {
    const agentId = `agent_wf_invalid_${testInfo.workerIndex}_${Date.now()}`;

    await seedAgent({
      id: agentId,
      tenantId: "t_enterprise",
      workspaceId: "ws_enterprise",
      slug: agentId.replaceAll("_", "-"),
      displayName: "Workflow Invalid Agent",
      modelProvider: "openai",
      modelName: "gpt-4.1-mini",
    });

    try {
      await page.goto(`/tenants/t_enterprise/agents/${agentId}/studio`);
      await expect(page.getByRole("heading", { name: "Workflow Invalid Agent" })).toBeVisible();

      // Select workflow pattern
      await page.getByRole("button", { name: /Workflow.*Deterministic DAG execution/ }).click();

      // Add a Model node (without setting modelRef)
      await page.getByRole("button", { name: /Model.*Call LLM/ }).click();

      // Click the Model node on canvas to open config panel
      await page.locator('[data-id^="model_"]').first().click();
      await expect(page.getByText("Model Reference")).toBeVisible();

      // Verify the validation warning appears for missing modelRef
      await expect(page.getByText("Model reference is required")).toBeVisible();

      // Attempt to save — validation should block it (no edges, missing refs)
      await page.getByRole("button", { name: "Save" }).click();

      // Should NOT redirect (stays on studio page due to validation error)
      await expect(page).toHaveURL(new RegExp(`/tenants/t_enterprise/agents/${agentId}/studio$`));
    } finally {
      await cleanupAgent(agentId);
    }
  });

  test("compile endpoint returns errors for invalid workflow spec", async () => {
    const agentId = `agent_wf_compile_${Date.now()}`;

    await seedAgent({
      id: agentId,
      tenantId: "t_enterprise",
      workspaceId: "ws_enterprise",
      slug: agentId.replaceAll("_", "-"),
      displayName: "Workflow Compile Agent",
    });

    // PATCH the agent with a workflow spec containing a model node without modelRef
    await api.patch(`/agents/${agentId}`, {
      pattern: "workflow",
      spec: {
        pattern: { type: "workflow" },
        models: {},
        toolRefs: [],
        knowledgeRefs: [],
        skillRefs: [],
        subAgentRefs: [],
        mcpRefs: [],
        graph: {
          nodes: [
            { name: "start", kind: "start" },
            { name: "model1", kind: "model" },
            { name: "end", kind: "end" },
          ],
          edges: [
            { from: "start", to: "model1" },
            { from: "model1", to: "end" },
          ],
        },
      },
    });

    try {
      // Call compile endpoint directly
      const result = await api.post(`/agents/${agentId}/compile`, {});

      // Should return errors because model node has no modelRef
      expect(result.ok).toBe(false);
      expect(result.errors).toBeDefined();
      expect(result.errors.length).toBeGreaterThan(0);
    } finally {
      await cleanupAgent(agentId);
    }
  });
});
