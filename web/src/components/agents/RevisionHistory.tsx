import { useTranslation } from "react-i18next";
import type { RevisionEntry } from "@/types/api";
import { Card } from "@/components/shared/Card";

interface RevisionHistoryProps {
  revisions: RevisionEntry[];
}

export function RevisionHistory({ revisions }: RevisionHistoryProps) {
  const { t } = useTranslation();

  if (revisions.length === 0) return null;

  return (
    <Card className="mt-8 p-6">
      <h2 className="mb-4 text-lg font-semibold text-zinc-950">{t("agent.revisionHistory")}</h2>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-200 text-left">
              <th className="pb-2 text-xs font-semibold uppercase text-zinc-500">
                {t("agent.fields.revision")}
              </th>
              <th className="pb-2 text-xs font-semibold uppercase text-zinc-500">
                {t("agent.fields.createdAt")}
              </th>
              <th className="pb-2 text-xs font-semibold uppercase text-zinc-500">
                {t("agent.fields.status")}
              </th>
            </tr>
          </thead>
          <tbody>
            {revisions.map((entry, i) => (
              <tr key={i} className="border-b border-zinc-100 last:border-0">
                <td className="py-2 font-mono text-xs text-zinc-800">
                  {entry.revision ? entry.revision.slice(0, 12) + "..." : "—"}
                </td>
                <td className="py-2 text-zinc-600">{entry.createdAt}</td>
                <td className="py-2">
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                      entry.status === "ok"
                        ? "bg-emerald-50 text-emerald-700"
                        : "bg-red-50 text-red-700"
                    }`}
                  >
                    {entry.status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
