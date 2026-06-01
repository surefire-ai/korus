import { createBrowserRouter, Navigate } from "react-router-dom";
import { AppShell } from "@/components/layout/AppShell";
import { LoginPage } from "@/pages/LoginPage";
import { TenantListPage } from "@/pages/TenantListPage";
import { WorkspaceListPage } from "@/pages/WorkspaceListPage";
import { WorkspaceDetailPage } from "@/pages/WorkspaceDetailPage";
import { WorkspaceCreatePage } from "@/pages/WorkspaceCreatePage";
import { ProductAreaPage } from "@/pages/ProductAreaPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { AgentListPage } from "@/pages/AgentListPage";
import { AgentDetailPage } from "@/pages/AgentDetailPage";
import { AgentStudioPage } from "@/pages/AgentStudioPage";
import { AgentCreatePage } from "@/pages/AgentCreatePage";
import { EvaluationListPage } from "@/pages/EvaluationListPage";
import { EvaluationDetailPage } from "@/pages/EvaluationDetailPage";
import { ProviderListPage } from "@/pages/ProviderListPage";
import { ProviderDetailPage } from "@/pages/ProviderDetailPage";
import { ProviderCreatePage } from "@/pages/ProviderCreatePage";
import { RunListPage } from "@/pages/RunListPage";
import { RunDetailPage } from "@/pages/RunDetailPage";
import { AuthGuard } from "@/components/auth/AuthGuard";

function AuthenticatedLayout() {
  return (
    <AuthGuard>
      <AppShell />
    </AuthGuard>
  );
}

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/",
    element: <AuthenticatedLayout />,
    children: [
      { index: true, element: <Navigate to="/tenants" replace /> },
      { path: "tenants", element: <TenantListPage /> },
      { path: "tenants/:tenantId/workspaces", element: <WorkspaceListPage /> },
      { path: "tenants/:tenantId/workspaces/new", element: <WorkspaceCreatePage /> },
      { path: "tenants/:tenantId/workspaces/:workspaceId", element: <WorkspaceDetailPage /> },
      { path: "tenants/:tenantId/agents", element: <AgentListPage /> },
      { path: "tenants/:tenantId/agents/new", element: <AgentCreatePage /> },
      { path: "tenants/:tenantId/agents/:agentId", element: <AgentDetailPage /> },
      { path: "tenants/:tenantId/agents/:agentId/studio", element: <AgentStudioPage /> },
      { path: "tenants/:tenantId/evaluations", element: <EvaluationListPage /> },
      { path: "tenants/:tenantId/evaluations/:evaluationId", element: <EvaluationDetailPage /> },
      { path: "tenants/:tenantId/runs", element: <RunListPage /> },
      { path: "tenants/:tenantId/runs/:runId", element: <RunDetailPage /> },
      { path: "tenants/:tenantId/providers", element: <ProviderListPage /> },
      { path: "tenants/:tenantId/providers/new", element: <ProviderCreatePage /> },
      { path: "tenants/:tenantId/providers/:providerId", element: <ProviderDetailPage /> },
      { path: "tenants/:tenantId/settings", element: <ProductAreaPage area="settings" /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
