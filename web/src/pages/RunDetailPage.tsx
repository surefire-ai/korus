import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useRun, useDeleteRun } from "@/api/runs";
import { RunDetailCard } from "@/components/runs/RunDetailCard";
import { PageHeader } from "@/components/shared/PageHeader";
import { LoadingSkeleton } from "@/components/shared/LoadingSkeleton";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import { Button } from "@/components/shared/Button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Trash2 } from "lucide-react";

export function RunDetailPage() {
  const { t } = useTranslation();
  const { tenantId, runId } = useParams<{ tenantId: string; runId: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const deleteMutation = useDeleteRun();
  const { data: run, isLoading, isError, error, refetch } = useRun(runId);
  useDocumentTitle(run?.id);

  if (isLoading) return <LoadingSkeleton />;

  if (isError) {
    return (
      <ErrorAlert
        message={error instanceof Error ? error.message : t("run.loadError")}
        onRetry={() => refetch()}
      />
    );
  }

  if (!run) {
    return <ErrorAlert message={t("run.notFound")} />;
  }

  return (
    <div>
      <PageHeader
        title={run.id}
        subtitle={t("run.detailSubtitle")}
        actions={
          <Button variant="danger" onClick={() => setShowDeleteDialog(true)}>
            <Trash2 className="mr-2 h-4 w-4" />
            {t("common.delete")}
          </Button>
        }
      />
      <RunDetailCard run={run} />

      <ConfirmDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={() => {
          deleteMutation.mutate(runId!, {
            onSuccess: () => navigate(`/tenants/${tenantId}/runs`),
          });
        }}
        title={t("run.deleteTitle", "Delete Run")}
        message={t("run.deleteMessage", "Are you sure you want to delete this run? This action cannot be undone.")}
        confirmLabel={t("common.delete")}
        isDestructive
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
