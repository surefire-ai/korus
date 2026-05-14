import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useProvider, useDeleteProvider } from "@/api/providers";
import { ProviderDetailCard } from "@/components/providers/ProviderDetailCard";
import { PageHeader } from "@/components/shared/PageHeader";
import { LoadingSkeleton } from "@/components/shared/LoadingSkeleton";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import { Button } from "@/components/shared/Button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Trash2 } from "lucide-react";

export function ProviderDetailPage() {
  const { t } = useTranslation();
  const { tenantId, providerId } = useParams<{ tenantId: string; providerId: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const deleteMutation = useDeleteProvider();
  const { data: provider, isLoading, isError, error, refetch } = useProvider(providerId);
  useDocumentTitle(provider?.displayName);

  if (isLoading) return <LoadingSkeleton />;

  if (isError) {
    return (
      <ErrorAlert
        message={error instanceof Error ? error.message : t("provider.loadError")}
        onRetry={() => refetch()}
      />
    );
  }

  if (!provider) {
    return <ErrorAlert message={t("provider.notFound")} />;
  }

  return (
    <div>
      <PageHeader
        title={provider.displayName}
        subtitle={t("provider.detailSubtitle")}
        actions={
          <Button variant="danger" onClick={() => setShowDeleteDialog(true)}>
            <Trash2 className="mr-2 h-4 w-4" />
            {t("common.delete")}
          </Button>
        }
      />
      <ProviderDetailCard provider={provider} />

      <ConfirmDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={() => {
          deleteMutation.mutate(providerId!, {
            onSuccess: () => navigate(`/tenants/${tenantId}/providers`),
          });
        }}
        title={t("provider.deleteTitle", "Delete Provider")}
        message={t("provider.deleteMessage", "Are you sure you want to delete this provider? This action cannot be undone.")}
        confirmLabel={t("common.delete")}
        isDestructive
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
