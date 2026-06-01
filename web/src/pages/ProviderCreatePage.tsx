import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useWorkspaces } from "@/api/workspaces";
import { useCreateProvider } from "@/api/providers";
import { ProviderCreateForm } from "@/components/providers/ProviderCreateForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import type { CreateProviderRequest } from "@/types/api";

export function ProviderCreatePage() {
  const { t } = useTranslation();
  useDocumentTitle(t("provider.createTitle", "New Provider"));
  const { tenantId } = useParams<{ tenantId: string }>();
  const navigate = useNavigate();
  const createMutation = useCreateProvider();
  const { data: workspaceData } = useWorkspaces(1, 100, tenantId);
  const workspaceOptions = (workspaceData?.workspaces ?? []).map((ws) => ({
    value: ws.id,
    label: ws.displayName || ws.id,
  }));

  const [values, setValues] = useState<CreateProviderRequest>({
    id: "",
    tenantId: tenantId ?? "",
    provider: "openai",
    displayName: "",
    family: "openai-compatible",
    status: "active",
  });

  const handleChange = (newValues: CreateProviderRequest) => {
    setValues((prev) => ({ ...prev, ...newValues, tenantId: tenantId ?? "" }));
  };

  const handleSubmit = () => {
    createMutation.mutate(values, {
      onSuccess: () => navigate(`/tenants/${tenantId}/providers`),
    });
  };

  return (
    <div>
      <PageHeader
        title={t("provider.createTitle", "New Provider")}
        subtitle={t("provider.createSubtitle", "Configure a new model provider connection.")}
      />

      {createMutation.isError && (
        <div className="mb-4">
          <ErrorAlert
            message={
              createMutation.error instanceof Error
                ? createMutation.error.message
                : t("provider.createError", "Failed to create provider.")
            }
          />
        </div>
      )}

      <ProviderCreateForm
        values={values}
        onChange={handleChange}
        onSubmit={handleSubmit}
        onCancel={() => navigate(`/tenants/${tenantId}/providers`)}
        isPending={createMutation.isPending}
        workspaceOptions={workspaceOptions}
      />
    </div>
  );
}
