import { test, expect } from "@playwright/test";
import { cleanupAgent, seedAgent } from "./fixtures/api";

test.describe("Agent Studio", () => {
  test("opens from agent detail and builds a workflow preview", async ({ page }) => {
    await page.goto("/tenants/t_demo/agents/agent_ehs_react");

    await page.getByRole("button", { name: "Open Studio" }).click();
    await expect(page).toHaveURL(/\/tenants\/t_demo\/agents\/agent_ehs_react\/studio/);
    await expect(page.getByRole("heading", { name: "EHS ReAct Agent" })).toBeVisible();
    await expect(page.getByText(/Orchestration Studio · react · eino\/adk/)).toBeVisible();

    await page.getByRole("button", { name: /Workflow.*Deterministic DAG execution/ }).click();
    await expect(page.getByText("Drag components from the left panel")).toBeVisible();

    await page.getByRole("button", { name: /Start.*Workflow entry/ }).click();
    await expect(page.getByText(/1 Nodes · 0 Edges/)).toBeVisible();

    await page.getByRole("button", { name: /Model.*Call LLM/ }).click();
    await expect(page.getByText(/2 Nodes · 0 Edges/)).toBeVisible();

    await page.getByRole("button", { name: /End.*Workflow exit/ }).click();
    await expect(page.getByText(/3 Nodes · 0 Edges/)).toBeVisible();

    await page.getByRole("button", { name: "Preview" }).click();
    await expect(page.getByRole("heading", { name: "Graph Preview" })).toBeVisible();
    await expect(page.getByText("start", { exact: true })).toBeVisible();
    await expect(page.getByText("model", { exact: true })).toBeVisible();
    await expect(page.getByText("end", { exact: true })).toBeVisible();
  });

  test("persists workflow edits after saving", async ({ page }, testInfo) => {
    const agentId = `agent_studio_save_${testInfo.workerIndex}_${Date.now()}`;

    await seedAgent({
      id: agentId,
      tenantId: "t_enterprise",
      workspaceId: "ws_enterprise",
      slug: agentId.replaceAll("_", "-"),
      displayName: "Studio Save Agent",
      modelProvider: "openai",
      modelName: "gpt-4.1-mini",
    });

    try {
      await page.goto(`/tenants/t_enterprise/agents/${agentId}/studio`);
      await expect(page.getByRole("heading", { name: "Studio Save Agent" })).toBeVisible();

      await page.getByRole("button", { name: /Workflow.*Deterministic DAG execution/ }).click();
      await page.getByRole("button", { name: /Start.*Workflow entry/ }).click();
      await page.getByRole("button", { name: /End.*Workflow exit/ }).click();
      await expect(page.getByText(/2 Nodes · 0 Edges/)).toBeVisible();

      await page.getByRole("button", { name: "Save" }).click();
      await expect(page).toHaveURL(new RegExp(`/tenants/t_enterprise/agents/${agentId}$`));

      await page.getByRole("button", { name: "Open Studio" }).click();
      await expect(page).toHaveURL(new RegExp(`/tenants/t_enterprise/agents/${agentId}/studio$`));
      await expect(page.getByText(/Orchestration Studio · workflow · eino\/adk/)).toBeVisible();
      await expect(page.getByText(/2 Nodes · 0 Edges/)).toBeVisible();

      await page.getByRole("button", { name: "Preview" }).click();
      await expect(page.getByRole("heading", { name: "Graph Preview" })).toBeVisible();
      await expect(page.getByText("start", { exact: true })).toBeVisible();
      await expect(page.getByText("end", { exact: true })).toBeVisible();
    } finally {
      await cleanupAgent(agentId);
    }
  });
});
