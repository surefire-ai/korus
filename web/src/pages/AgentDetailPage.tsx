import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAgent, useDeleteAgent } from "@/api/agents";
import { useRuns } from "@/api/runs";
import { useEvaluations } from "@/api/evaluations";
import { AgentDetailCard } from "@/components/agents/AgentDetailCard";
import { RevisionHistory } from "@/components/agents/RevisionHistory";
import { RunTable } from "@/components/runs/RunTable";
import { EvaluationTable } from "@/components/evaluations/EvaluationTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { LoadingSkeleton } from "@/components/shared/LoadingSkeleton";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import { Card } from "@/components/shared/Card";
import { EmptyState } from "@/components/shared/EmptyState";
import { Button } from "@/components/shared/Button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Settings, Trash2 } from "lucide-react";

const LINKED_LIMIT = 5;

export function AgentDetailPage() {
  const { t } = useTranslation();
  const { tenantId, agentId } = useParams<{ tenantId: string; agentId: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const deleteMutation = useDeleteAgent();
  const { data: agent, isLoading, isError, error, refetch } = useAgent(agentId);
  useDocumentTitle(agent?.displayName);

  const { data: runsData } = useRuns(1, LINKED_LIMIT, tenantId, undefined, agentId);
  const { data: evalsData } = useEvaluations(1, LINKED_LIMIT, tenantId, undefined, agentId);

  if (isLoading) return <LoadingSkeleton />;

  if (isError) {
    return (
      <ErrorAlert
        message={error instanceof Error ? error.message : t("agent.loadError")}
        onRetry={() => refetch()}
      />
    );
  }

  if (!agent) {
    return <ErrorAlert message={t("agent.notFound")} />;
  }

  return (
    <div>
      <PageHeader
        title={agent.displayName}
        subtitle={t("agent.detailSubtitle")}
        actions={
          <>
            <Button
              variant="secondary"
              onClick={() => navigate(`/tenants/${tenantId}/agents/${agentId}/studio`)}
            >
              <Settings className="mr-2 h-4 w-4" />
              {t("studio.openStudio")}
            </Button>
            <Button
              variant="danger"
              onClick={() => setShowDeleteDialog(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {t("common.delete")}
            </Button>
          </>
        }
      />

      <AgentDetailCard agent={agent} />

      {/* Revision History */}
      {agent.revisions && agent.revisions.length > 0 && (
        <RevisionHistory revisions={agent.revisions} />
      )}

      {/* Linked Runs */}
      <div className="mt-8">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-zinc-950">{t("agent.linkedRuns")}</h2>
            <p className="mt-1 text-sm text-zinc-500">{t("agent.linkedRunsSubtitle")}</p>
          </div>
          {runsData && runsData.runs.length > 0 && (
            <Link
              to={`/tenants/${tenantId}/runs`}
              className="text-sm font-medium text-teal-600 hover:text-teal-700"
            >
              {t("common.viewAll", "View all")}
            </Link>
          )}
        </div>
        {runsData && runsData.runs.length === 0 && (
          <Card className="p-6">
            <EmptyState title={t("run.emptyTitle")} description={t("run.emptyDescription")} />
          </Card>
        )}
        {runsData && runsData.runs.length > 0 && <RunTable runs={runsData.runs} />}
      </div>

      {/* Linked Evaluations */}
      <div className="mt-8">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-zinc-950">{t("agent.linkedEvaluations")}</h2>
            <p className="mt-1 text-sm text-zinc-500">{t("agent.linkedEvaluationsSubtitle")}</p>
          </div>
          {evalsData && evalsData.evaluations.length > 0 && (
            <Link
              to={`/tenants/${tenantId}/evaluations`}
              className="text-sm font-medium text-teal-600 hover:text-teal-700"
            >
              {t("common.viewAll", "View all")}
            </Link>
          )}
        </div>
        {evalsData && evalsData.evaluations.length === 0 && (
          <Card className="p-6">
            <EmptyState title={t("evaluation.emptyTitle")} description={t("evaluation.emptyDescription")} />
          </Card>
        )}
        {evalsData && evalsData.evaluations.length > 0 && <EvaluationTable evaluations={evalsData.evaluations} />}
      </div>

      <ConfirmDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={() => {
          deleteMutation.mutate(agentId!, {
            onSuccess: () => navigate(`/tenants/${tenantId}/agents`),
          });
        }}
        title={t("agent.deleteTitle", "Delete Agent")}
        message={t("agent.deleteMessage", "Are you sure you want to delete this agent? This action cannot be undone.")}
        confirmLabel={t("common.delete")}
        isDestructive
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
