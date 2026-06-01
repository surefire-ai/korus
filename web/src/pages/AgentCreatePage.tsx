import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useWorkspaces } from "@/api/workspaces";
import { useCreateAgent } from "@/api/agents";
import { AgentCreateForm } from "@/components/agents/AgentCreateForm";
import type { WorkspaceOption } from "@/components/agents/AgentCreateForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import type { CreateAgentRequest } from "@/types/api";

export function AgentCreatePage() {
  const { t } = useTranslation();
  useDocumentTitle(t("agent.createTitle"));
  const { tenantId } = useParams<{ tenantId: string }>();
  const navigate = useNavigate();
  const createMutation = useCreateAgent();
  const { data: workspaceData } = useWorkspaces(1, 100, tenantId);
  const workspaceOptions: WorkspaceOption[] = (workspaceData?.workspaces ?? []).map((ws) => ({ value: ws.id, label: ws.displayName || ws.id }));

  const [values, setValues] = useState<CreateAgentRequest>({
    id: "",
    tenantId: tenantId ?? "",
    workspaceId: "",
    slug: "",
    displayName: "",
    status: "draft",
    runtimeEngine: "eino",
    runnerClass: "adk",
    pattern: "react",
  });

  const handleChange = (newValues: CreateAgentRequest) => {
    setValues((prev) => ({ ...prev, ...newValues, tenantId: tenantId ?? "" }));
  };

  const handleSubmit = () => {
    createMutation.mutate(values, {
      onSuccess: () => navigate(`/tenants/${tenantId}/agents`),
    });
  };

  return (
    <div>
      <PageHeader
        title={t("agent.createTitle")}
        subtitle={t("agent.createSubtitle")}
      />

      {createMutation.isError && (
        <div className="mb-4">
          <ErrorAlert
            message={
              createMutation.error instanceof Error
                ? createMutation.error.message
                : t("agent.createError")
            }
          />
        </div>
      )}

      <AgentCreateForm
        values={values}
        onChange={handleChange}
        onSubmit={handleSubmit}
        onCancel={() => navigate(`/tenants/${tenantId}/agents`)}
        isPending={createMutation.isPending}
        workspaceOptions={workspaceOptions}
      />
    </div>
  );
}
