import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { CreateAgentRequest } from "@/types/api";
import { Button } from "@/components/shared/Button";
import { Input } from "@/components/shared/Input";
import { Textarea } from "@/components/shared/Textarea";
import { Select } from "@/components/shared/Select";
import { Field } from "@/components/shared/Field";
import { Card } from "@/components/shared/Card";

export interface WorkspaceOption {
  value: string;
  label: string;
}

interface AgentCreateFormProps {
  values: CreateAgentRequest;
  onChange: (values: CreateAgentRequest) => void;
  onSubmit: () => void;
  onCancel: () => void;
  isPending: boolean;
  workspaceOptions: WorkspaceOption[];
}

interface FormErrors {
  id?: string;
  slug?: string;
  displayName?: string;
  workspaceId?: string;
}

const ID_RE = /^[a-z][a-z0-9_-]*$/;
const SLUG_RE = /^[a-z][a-z0-9-]*$/;

export function AgentCreateForm({
  values,
  onChange,
  onSubmit,
  onCancel,
  isPending,
  workspaceOptions,
}: AgentCreateFormProps) {
  const { t } = useTranslation();
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitted, setSubmitted] = useState(false);

  const statusOptions = [
    { value: "draft", label: t("status.draft") },
    { value: "active", label: t("status.active") },
  ];

  const patternOptions = [
    { value: "react", label: "ReAct" },
    { value: "single", label: "Single" },
    { value: "router", label: "Router" },
    { value: "graph", label: "Graph" },
  ];

  const runtimeEngineOptions = [
    { value: "eino", label: "Eino" },
  ];

  const runnerClassOptions = [
    { value: "adk", label: "ADK" },
  ];

  const set = (key: keyof CreateAgentRequest, value: string) => {
    onChange({ ...values, [key]: value });
    if (submitted) validateField(key, value);
  };

  const validateField = (key: string, value: string) => {
    const newErrors = { ...errors };
    if (key === "id") {
      if (!value.trim()) {
        newErrors.id = t("validation.required");
      } else if (!ID_RE.test(value)) {
        newErrors.id = t("validation.idFormat");
      } else {
        delete newErrors.id;
      }
    } else if (key === "slug") {
      if (!value.trim()) {
        newErrors.slug = t("validation.required");
      } else if (!SLUG_RE.test(value)) {
        newErrors.slug = t("validation.slugFormat");
      } else {
        delete newErrors.slug;
      }
    } else if (key === "displayName") {
      if (!value.trim()) {
        newErrors.displayName = t("validation.required");
      } else {
        delete newErrors.displayName;
      }
    } else if (key === "workspaceId") {
      if (!value.trim()) {
        newErrors.workspaceId = t("validation.required");
      } else {
        delete newErrors.workspaceId;
      }
    }
    setErrors(newErrors);
  };

  const validate = (): boolean => {
    const newErrors: FormErrors = {};
    if (!values.id.trim()) newErrors.id = t("validation.required");
    else if (!ID_RE.test(values.id)) newErrors.id = t("validation.idFormat");
    if (!values.slug.trim()) newErrors.slug = t("validation.required");
    else if (!SLUG_RE.test(values.slug)) newErrors.slug = t("validation.slugFormat");
    if (!values.displayName.trim()) newErrors.displayName = t("validation.required");
    if (!values.workspaceId.trim()) newErrors.workspaceId = t("validation.required");
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    if (validate()) onSubmit();
  };

  return (
    <Card className="p-6">
      <form onSubmit={handleSubmit} className="space-y-6" noValidate>
        <Field label={t("agent.fields.id")} htmlFor="id" required error={errors.id}>
          <Input
            id="id"
            placeholder="agent_my_agent"
            value={values.id}
            onChange={(e) => set("id", e.target.value)}
            hasError={!!errors.id}
          />
        </Field>

        <Field label={t("agent.fields.slug")} htmlFor="slug" required error={errors.slug}>
          <Input
            id="slug"
            placeholder="my-agent"
            value={values.slug}
            onChange={(e) => set("slug", e.target.value)}
            hasError={!!errors.slug}
          />
        </Field>

        <Field label={t("agent.fields.displayName")} htmlFor="displayName" required error={errors.displayName}>
          <Input
            id="displayName"
            placeholder="My Agent"
            value={values.displayName}
            onChange={(e) => set("displayName", e.target.value)}
            hasError={!!errors.displayName}
          />
        </Field>

        <Field label={t("agent.fields.workspaceId")} htmlFor="workspaceId" required error={errors.workspaceId}>
          <Select
            id="workspaceId"
            options={workspaceOptions}
            placeholder={workspaceOptions.length === 0 ? t("common.noOptions", "No workspaces available") : undefined}
            value={values.workspaceId}
            onChange={(e) => set("workspaceId", e.target.value)}
          />
        </Field>

        <Field label={t("agent.fields.description")} htmlFor="description">
          <Textarea
            id="description"
            rows={3}
            value={values.description ?? ""}
            onChange={(e) => set("description", e.target.value)}
          />
        </Field>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <Field label={t("agent.fields.status")} htmlFor="status">
            <Select
              id="status"
              options={statusOptions}
              value={values.status ?? "draft"}
              onChange={(e) => set("status", e.target.value)}
            />
          </Field>

          <Field label={t("agent.fields.pattern")} htmlFor="pattern">
            <Select
              id="pattern"
              options={patternOptions}
              value={values.pattern ?? "react"}
              onChange={(e) => set("pattern", e.target.value)}
            />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <Field label={t("agent.fields.runtimeEngine")} htmlFor="runtimeEngine">
            <Select
              id="runtimeEngine"
              options={runtimeEngineOptions}
              value={values.runtimeEngine ?? "eino"}
              onChange={(e) => set("runtimeEngine", e.target.value)}
            />
          </Field>

          <Field label={t("agent.fields.runnerClass")} htmlFor="runnerClass">
            <Select
              id="runnerClass"
              options={runnerClassOptions}
              value={values.runnerClass ?? "adk"}
              onChange={(e) => set("runnerClass", e.target.value)}
            />
          </Field>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onCancel}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? t("common.creating") : t("agent.createButton")}
          </Button>
        </div>
      </form>
    </Card>
  );
}
