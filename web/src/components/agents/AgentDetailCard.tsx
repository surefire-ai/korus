import { useTranslation } from "react-i18next";
import type { Agent } from "@/types/api";
import { Card } from "@/components/shared/Card";
import { StatusBadge } from "@/components/shared/StatusBadge";

interface AgentDetailCardProps {
  agent: Agent;
}

export function AgentDetailCard({ agent }: AgentDetailCardProps) {
  const { t } = useTranslation();
  const nd = t("common.noData");

  return (
    <Card className="p-6">
      <dl className="grid gap-4 sm:grid-cols-2">
        {/* Identity */}
        <dt className="detail-section-label sm:col-span-2">{t("detailSection.identity")}</dt>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.id")}</dt>
          <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.id}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.slug")}</dt>
          <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.slug}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.displayName")}</dt>
          <dd className="mt-2 text-sm font-semibold text-zinc-950">{agent.displayName}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.status")}</dt>
          <dd className="mt-2">
            <StatusBadge status={agent.status} />
          </dd>
        </div>
        <div className="surface-muted rounded-lg p-4 sm:col-span-2">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.description")}</dt>
          <dd className="mt-2 text-sm leading-6 text-zinc-800">{agent.description ?? nd}</dd>
        </div>

        {/* Runtime */}
        <dt className="detail-section-label sm:col-span-2">{t("detailSection.runtime")}</dt>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.pattern")}</dt>
          <dd className="mt-2 text-sm font-mono text-zinc-950">{agent.pattern}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.runtimeEngine")}</dt>
          <dd className="mt-2 text-sm font-mono text-zinc-950">{agent.runtimeEngine}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.runnerClass")}</dt>
          <dd className="mt-2 text-sm font-mono text-zinc-950">{agent.runnerClass}</dd>
        </div>

        {/* Model */}
        <dt className="detail-section-label sm:col-span-2">{t("detailSection.model")}</dt>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.modelProvider")}</dt>
          <dd className="mt-2 text-sm font-semibold text-zinc-800">{agent.modelProvider ?? nd}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.modelName")}</dt>
          <dd className="mt-2 text-sm font-mono text-zinc-800">{agent.modelName ?? nd}</dd>
        </div>

        {/* References */}
        <dt className="detail-section-label sm:col-span-2">{t("detailSection.references")}</dt>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.tenantId")}</dt>
          <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.tenantId}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.workspaceId")}</dt>
          <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.workspaceId}</dd>
        </div>
        <div className="surface-muted rounded-lg p-4 sm:col-span-2">
          <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.latestRevision")}</dt>
          <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.latestRevision ?? nd}</dd>
        </div>

        {/* Compile Status */}
        {agent.compileStatus && (
          <>
            <dt className="detail-section-label sm:col-span-2">{t("detailSection.compile")}</dt>
            <div className="surface-muted rounded-lg p-4">
              <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.compileStatus")}</dt>
              <dd className="mt-2">
                <span
                  className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                    agent.compileStatus === "ok"
                      ? "bg-emerald-50 text-emerald-700"
                      : "bg-red-50 text-red-700"
                  }`}
                >
                  {agent.compileStatus}
                </span>
              </dd>
            </div>
            <div className="surface-muted rounded-lg p-4">
              <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.latestRevision")}</dt>
              <dd className="mt-2 break-all text-sm font-mono text-zinc-950">{agent.latestRevision ?? nd}</dd>
            </div>
            {agent.compileErrors && agent.compileErrors.length > 0 && (
              <div className="surface-muted rounded-lg p-4 sm:col-span-2">
                <dt className="text-xs font-semibold uppercase text-zinc-500">{t("agent.fields.compileErrors")}</dt>
                <dd className="mt-2">
                  <ul className="list-inside list-disc space-y-1 text-sm text-red-600">
                    {agent.compileErrors.map((err, i) => (
                      <li key={i}>{err}</li>
                    ))}
                  </ul>
                </dd>
              </div>
            )}
          </>
        )}
      </dl>
    </Card>
  );
}
