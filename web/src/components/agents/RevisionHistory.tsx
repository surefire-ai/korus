import { useTranslation } from "react-i18next";
import type { RevisionEntry } from "@/types/api";
import { Card } from "@/components/shared/Card";
import { StatusBadge } from "@/components/shared/StatusBadge";

interface RevisionHistoryProps {
  revisions: RevisionEntry[];
}

export function RevisionHistory({ revisions }: RevisionHistoryProps) {
  const { t } = useTranslation();

  if (revisions.length === 0) return null;

  return (
    <Card className="mt-8 overflow-hidden">
      <div className="border-b border-zinc-200/60 px-6 py-4">
        <h2 className="text-lg font-semibold text-zinc-950">{t("agent.revisionHistory")}</h2>
      </div>
      <div className="overflow-x-auto">
        <table className="data-table min-w-full">
          <thead>
            <tr>
              <th scope="col">{t("agent.fields.revision")}</th>
              <th scope="col">{t("agent.fields.createdAt")}</th>
              <th scope="col">{t("agent.fields.status")}</th>
            </tr>
          </thead>
          <tbody>
            {revisions.map((entry, i) => (
              <tr key={i}>
                <td className="font-mono text-xs text-zinc-800">
                  {entry.revision ? entry.revision.slice(0, 12) + "..." : "—"}
                </td>
                <td className="text-zinc-600">{entry.createdAt}</td>
                <td>
                  <StatusBadge status={entry.status} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
