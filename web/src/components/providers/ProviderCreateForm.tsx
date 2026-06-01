import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { CreateProviderRequest } from "@/types/api";
import { Button } from "@/components/shared/Button";
import { Input } from "@/components/shared/Input";
import { Select } from "@/components/shared/Select";
import { Field } from "@/components/shared/Field";
import { Card } from "@/components/shared/Card";

interface ProviderCreateFormProps {
  values: CreateProviderRequest;
  onChange: (values: CreateProviderRequest) => void;
  onSubmit: () => void;
  onCancel: () => void;
  isPending: boolean;
  workspaceOptions: { value: string; label: string }[];
}

interface FormErrors {
  id?: string;
  provider?: string;
  displayName?: string;
}

const ID_RE = /^[a-z][a-z0-9_-]*$/;

export function ProviderCreateForm({
  values,
  onChange,
  onSubmit,
  onCancel,
  isPending,
  workspaceOptions,
}: ProviderCreateFormProps) {
  const { t } = useTranslation();
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitted, setSubmitted] = useState(false);

  const providerOptions = [
    { value: "openai", label: "OpenAI" },
    { value: "qwen", label: "Qwen" },
    { value: "deepseek", label: "DeepSeek" },
    { value: "anthropic", label: "Anthropic" },
    { value: "custom", label: t("provider.custom", "Custom") },
  ];

  const familyOptions = [
    { value: "openai-compatible", label: "OpenAI Compatible" },
  ];

  const statusOptions = [
    { value: "active", label: t("status.active") },
    { value: "inactive", label: t("status.inactive") },
  ];

  const set = (key: keyof CreateProviderRequest, value: string | boolean | undefined) => {
    onChange({ ...values, [key]: value });
    if (submitted && typeof value === "string") validateField(key, value);
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
    } else if (key === "provider") {
      if (!value.trim()) {
        newErrors.provider = t("validation.required");
      } else {
        delete newErrors.provider;
      }
    } else if (key === "displayName") {
      if (!value.trim()) {
        newErrors.displayName = t("validation.required");
      } else {
        delete newErrors.displayName;
      }
    }
    setErrors(newErrors);
  };

  const validate = (): boolean => {
    const newErrors: FormErrors = {};
    if (!values.id.trim()) newErrors.id = t("validation.required");
    else if (!ID_RE.test(values.id)) newErrors.id = t("validation.idFormat");
    if (!values.provider.trim()) newErrors.provider = t("validation.required");
    if (!values.displayName.trim()) newErrors.displayName = t("validation.required");
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
        <Field label={t("provider.fields.id")} htmlFor="id" required error={errors.id}>
          <Input
            id="id"
            placeholder="provider_my_provider"
            value={values.id}
            onChange={(e) => set("id", e.target.value)}
            hasError={!!errors.id}
          />
        </Field>

        <Field label={t("provider.fields.displayName")} htmlFor="displayName" required error={errors.displayName}>
          <Input
            id="displayName"
            placeholder="My Provider"
            value={values.displayName}
            onChange={(e) => set("displayName", e.target.value)}
            hasError={!!errors.displayName}
          />
        </Field>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <Field label={t("provider.fields.provider")} htmlFor="provider" required error={errors.provider}>
            <Select
              id="provider"
              options={providerOptions}
              value={values.provider}
              onChange={(e) => set("provider", e.target.value)}
            />
          </Field>

          <Field label={t("provider.fields.family")} htmlFor="family">
            <Select
              id="family"
              options={familyOptions}
              value={values.family ?? "openai-compatible"}
              onChange={(e) => set("family", e.target.value)}
            />
          </Field>
        </div>

        <Field label={t("provider.fields.baseUrl")} htmlFor="baseUrl">
          <Input
            id="baseUrl"
            placeholder="https://api.openai.com/v1"
            value={values.baseUrl ?? ""}
            onChange={(e) => set("baseUrl", e.target.value)}
          />
        </Field>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <Field label={t("provider.fields.workspaceId")} htmlFor="workspaceId">
            <Select
              id="workspaceId"
              options={workspaceOptions}
              placeholder={workspaceOptions.length === 0 ? t("common.noOptions", "No workspaces") : undefined}
              value={values.workspaceId ?? ""}
              onChange={(e) => set("workspaceId", e.target.value || undefined)}
            />
          </Field>

          <Field label={t("provider.fields.status")} htmlFor="status">
            <Select
              id="status"
              options={statusOptions}
              value={values.status ?? "active"}
              onChange={(e) => set("status", e.target.value)}
            />
          </Field>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={onCancel}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? t("common.creating") : t("provider.createButton")}
          </Button>
        </div>
      </form>
    </Card>
  );
}
