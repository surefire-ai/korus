import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useAgent, useUpdateAgent, usePublishAgent } from "@/api/agents";
import type {
  AgentSpecData,
  PatternConfig,
  ModelConfig,
  KnowledgeBinding,
  SkillBinding,
  SubAgentBinding,
  GraphConfig,
  WorkflowBindings,
  CompileResult,
} from "@/types/api";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/shared/Button";
import { LoadingSkeleton } from "@/components/shared/LoadingSkeleton";
import { ErrorAlert } from "@/components/shared/ErrorAlert";
import { PatternSelector } from "@/components/studio/PatternSelector";
import { PatternConfigForm } from "@/components/studio/PatternConfigForm";
import { ModelConfigForm } from "@/components/studio/ModelConfigForm";
import { BindingPanel, StringArrayBindingPanel } from "@/components/studio/BindingPanel";
import { GraphPreview } from "@/components/studio/GraphPreview";
import { WorkflowCanvas } from "@/components/studio/flow/WorkflowCanvas";
import { Input } from "@/components/shared/Input";
import { useToast, ToastContainer } from "@/components/shared/Toast";
import { CheckCircle2, XCircle, Save, Rocket } from "lucide-react";

type TabKey = "pattern" | "models" | "preview";

function defaultSpec(): AgentSpecData {
  return {
    pattern: { type: "react", modelRef: "" },
    models: {},
    toolRefs: [],
    knowledgeRefs: [],
    skillRefs: [],
    subAgentRefs: [],
    mcpRefs: [],
  };
}

export function AgentStudioPage() {
  const { t } = useTranslation();
  const { tenantId, agentId } = useParams<{ tenantId: string; agentId: string }>();
  const navigate = useNavigate();
  const { data: agent, isLoading, isError, error, refetch } = useAgent(agentId);
  const saveMutation = useUpdateAgent();
  const publishMutation = usePublishAgent();

  const [activeTab, setActiveTab] = useState<TabKey>("pattern");
  const [spec, setSpec] = useState<AgentSpecData>(defaultSpec());
  const [saveStatus, setSaveStatus] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [publishResult, setPublishResult] = useState<CompileResult | null>(null);
  const toast = useToast();

  // Ref for workflow validation function
  const workflowValidateRef = useRef<(() => string[]) | null>(null);

  // Derive available bindings from spec for workflow node selects
  const bindings: WorkflowBindings | undefined = useMemo(() => {
    if (spec.pattern?.type !== "workflow") return undefined;
    return {
      modelNames: Object.keys(spec.models ?? {}),
      toolNames: spec.toolRefs ?? [],
      knowledgeNames: (spec.knowledgeRefs ?? []).map((k) => k.name).filter(Boolean),
      agentNames: (spec.subAgentRefs ?? []).map((s) => s.name).filter(Boolean),
    };
  }, [spec.pattern?.type, spec.models, spec.toolRefs, spec.knowledgeRefs, spec.subAgentRefs]);

  useDocumentTitle(agent?.displayName ? `${agent.displayName} — Studio` : t("studio.title"));

  // Initialize spec from agent data
  useEffect(() => {
    if (agent?.spec) {
      setSpec({
        ...defaultSpec(),
        ...agent.spec,
        pattern: agent.spec.pattern ?? { type: agent.pattern ?? "react", modelRef: "" },
        models: agent.spec.models ?? {},
        toolRefs: agent.spec.toolRefs ?? [],
        knowledgeRefs: agent.spec.knowledgeRefs ?? [],
        skillRefs: agent.spec.skillRefs ?? [],
        subAgentRefs: agent.spec.subAgentRefs ?? [],
        mcpRefs: agent.spec.mcpRefs ?? [],
        graph: agent.spec.graph ?? { nodes: [], edges: [] },
      });
    } else if (agent) {
      setSpec({
        ...defaultSpec(),
        pattern: { type: agent.pattern ?? "react", modelRef: "" },
      });
    }
  }, [agent]);

  const handleSave = useCallback(() => {
    if (!agentId) return;

    // Run workflow validation if in workflow mode
    if (spec.pattern?.type === "workflow" && workflowValidateRef.current) {
      const errors = workflowValidateRef.current();
      if (errors.length > 0) {
        setSaveStatus("error");
        return; // Validation errors shown in canvas
      }
    }

    setSaveStatus("saving");
    const patchData = {
      pattern: spec.pattern?.type,
      spec,
    };
    saveMutation.mutate(
      { id: agentId, ...patchData },
      {
        onSuccess: () => {
          setSaveStatus("saved");
          toast.success(t("studio.saved"));
          setTimeout(() => {
            navigate(`/tenants/${tenantId}/agents/${agentId}`);
          }, 800);
        },
        onError: () => {
          setSaveStatus("error");
          toast.error(t("studio.saveError"));
        },
      }
    );
  }, [agentId, spec, saveMutation, tenantId, navigate]);

  const handleSaveAndPublish = useCallback(() => {
    if (!agentId) return;

    // Run workflow validation if in workflow mode
    if (spec.pattern?.type === "workflow" && workflowValidateRef.current) {
      const errors = workflowValidateRef.current();
      if (errors.length > 0) {
        setSaveStatus("error");
        return;
      }
    }

    setSaveStatus("saving");
    setPublishResult(null);
    const patchData = { pattern: spec.pattern?.type, spec };

    saveMutation.mutate(
      { id: agentId, ...patchData },
      {
        onSuccess: () => {
          publishMutation.mutate(agentId, {
            onSuccess: (result) => {
              setPublishResult(result);
              if (result.ok) {
                setSaveStatus("saved");
                toast.success(t("studio.published"));
              } else {
                setSaveStatus("error");
                toast.error(t("studio.publishError"));
              }
            },
            onError: () => {
              setSaveStatus("error");
              toast.error(t("studio.publishError"));
            },
          });
        },
        onError: () => {
          setSaveStatus("error");
          toast.error(t("studio.saveError"));
        },
      }
    );
  }, [agentId, spec, saveMutation, publishMutation, tenantId, toast, t]);

  const handleGraphChange = useCallback((graph: GraphConfig) => {
    setSpec((prev) => ({ ...prev, graph }));
  }, []);

  const handlePatternSelect = useCallback((pattern: string) => {
    setSpec((prev) => ({
      ...prev,
      pattern: { ...prev.pattern, type: pattern, modelRef: prev.pattern?.modelRef },
    }));
  }, []);

  const handlePatternConfigChange = useCallback((patternConfig: PatternConfig) => {
    setSpec((prev) => ({ ...prev, pattern: patternConfig }));
  }, []);

  const handleModelsChange = useCallback((models: Record<string, ModelConfig>) => {
    setSpec((prev) => ({ ...prev, models }));
  }, []);

  const handleToolRefsChange = useCallback((toolRefs: string[]) => {
    setSpec((prev) => ({ ...prev, toolRefs }));
  }, []);

  const handleMcpRefsChange = useCallback((mcpRefs: string[]) => {
    setSpec((prev) => ({ ...prev, mcpRefs }));
  }, []);

  const handleKnowledgeAdd = useCallback(() => {
    setSpec((prev) => ({ ...prev, knowledgeRefs: [...(prev.knowledgeRefs ?? []), { name: "", ref: "" }] }));
  }, []);

  const handleKnowledgeRemove = useCallback((index: number) => {
    setSpec((prev) => ({ ...prev, knowledgeRefs: (prev.knowledgeRefs ?? []).filter((_, i) => i !== index) }));
  }, []);

  const handleKnowledgeChange = useCallback((index: number, field: keyof KnowledgeBinding, value: string | number) => {
    setSpec((prev) => ({
      ...prev,
      knowledgeRefs: (prev.knowledgeRefs ?? []).map((k, i) => i === index ? { ...k, [field]: value } : k),
    }));
  }, []);

  const handleSkillAdd = useCallback(() => {
    setSpec((prev) => ({ ...prev, skillRefs: [...(prev.skillRefs ?? []), { name: "", ref: "" }] }));
  }, []);

  const handleSkillRemove = useCallback((index: number) => {
    setSpec((prev) => ({ ...prev, skillRefs: (prev.skillRefs ?? []).filter((_, i) => i !== index) }));
  }, []);

  const handleSkillChange = useCallback((index: number, field: keyof SkillBinding, value: string) => {
    setSpec((prev) => ({
      ...prev,
      skillRefs: (prev.skillRefs ?? []).map((s, i) => i === index ? { ...s, [field]: value } : s),
    }));
  }, []);

  const handleSubAgentAdd = useCallback(() => {
    setSpec((prev) => ({ ...prev, subAgentRefs: [...(prev.subAgentRefs ?? []), { name: "", ref: "" }] }));
  }, []);

  const handleSubAgentRemove = useCallback((index: number) => {
    setSpec((prev) => ({ ...prev, subAgentRefs: (prev.subAgentRefs ?? []).filter((_, i) => i !== index) }));
  }, []);

  const handleSubAgentChange = useCallback((index: number, field: keyof SubAgentBinding, value: string) => {
    setSpec((prev) => ({
      ...prev,
      subAgentRefs: (prev.subAgentRefs ?? []).map((s, i) => i === index ? { ...s, [field]: value } : s),
    }));
  }, []);

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

  const isWorkflow = spec.pattern?.type === "workflow";

  return (
    <div>
      <PageHeader
        title={agent.displayName}
        subtitle={`${t("studio.title")} · ${agent.pattern} · ${agent.runtimeEngine}/${agent.runnerClass}`}
      />

      {/* Toolbar */}
      <div className="mb-6 flex flex-col gap-4 rounded-xl border border-zinc-200/80 bg-white/80 px-5 py-3 shadow-sm backdrop-blur-sm sm:flex-row sm:items-center sm:justify-between">
        {/* Tabs */}
        <div className="flex gap-1 rounded-lg bg-zinc-100/80 p-1">
          {(["pattern", "models", "preview"] as TabKey[]).map((tab) => (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`rounded-md px-4 py-1.5 text-sm font-medium transition-all ${
                activeTab === tab
                  ? "bg-white text-zinc-900 shadow-sm"
                  : "text-zinc-500 hover:text-zinc-700"
              }`}
            >
              {tab === "pattern" ? (isWorkflow ? t("studio.tabs.workflow") : t("studio.tabs.pattern")) : t(`studio.tabs.${tab}`)}
            </button>
          ))}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {saveStatus === "saved" && !publishResult && (
            <span className="save-pulse inline-flex items-center gap-1 rounded-full bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-700">
              <CheckCircle2 className="h-3 w-3" />
              {t("studio.saved")}
            </span>
          )}
          {publishResult?.ok && (
            <span className="save-pulse inline-flex items-center gap-1 rounded-full bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-700">
              <CheckCircle2 className="h-3 w-3" />
              {t("studio.published")} {publishResult.revision && `· ${publishResult.revision.slice(0, 8)}`}
            </span>
          )}
          {publishResult && !publishResult.ok && publishResult.errors && (
            <span className="inline-flex items-center gap-1 rounded-full bg-rose-50 px-2.5 py-1 text-xs font-medium text-rose-700">
              <XCircle className="h-3 w-3" />
              {t("studio.compileError")} · {publishResult.errors.length}
            </span>
          )}
          <div className="h-5 w-px bg-zinc-200" />
          <Button variant="ghost" size="sm" onClick={() => navigate(`/tenants/${tenantId}/agents/${agentId}`)}>
            {t("studio.cancel")}
          </Button>
          <Button variant="secondary" size="sm" onClick={handleSave} disabled={saveStatus === "saving" || saveStatus === "saved"}>
            <Save className="mr-1.5 h-3.5 w-3.5" />
            {saveStatus === "saving" && t("studio.saving")}
            {saveStatus === "saved" && t("studio.saved")}
            {saveStatus === "error" && !publishResult && t("studio.saveError")}
            {saveStatus === "idle" && t("studio.save")}
          </Button>
          <Button size="sm" onClick={handleSaveAndPublish} disabled={saveStatus === "saving" || (saveStatus === "saved" && publishResult?.ok)}>
            <Rocket className="mr-1.5 h-3.5 w-3.5" />
            {saveStatus === "saving" && t("studio.publishing")}
            {saveStatus === "saved" && publishResult?.ok && t("studio.published")}
            {saveStatus === "error" && publishResult && !publishResult.ok && t("studio.publishError")}
            {saveStatus === "idle" && t("studio.saveAndPublish")}
          </Button>
        </div>
      </div>

      {/* Compile errors */}
      {publishResult && !publishResult.ok && publishResult.errors && publishResult.errors.length > 0 && (
        <div className="mb-6">
          <ErrorAlert message={`${t("studio.compileErrors")}: ${publishResult.errors.join("; ")}`} />
        </div>
      )}

      {/* Tab Content */}
      <div className="rounded-xl border border-zinc-200/80 bg-white p-6 shadow-sm">
        {activeTab === "pattern" && (
          <div key="pattern" className="tab-content-enter">
            <PatternSelector selected={spec.pattern?.type ?? "react"} onSelect={handlePatternSelect} />
            {isWorkflow ? (
              <WorkflowCanvas
                graph={spec.graph ?? { nodes: [], edges: [] }}
                onChange={handleGraphChange}
                onValidateRef={workflowValidateRef}
                bindings={bindings}
              />
            ) : (
              <PatternConfigForm
                pattern={spec.pattern?.type ?? "react"}
                config={spec.pattern ?? {}}
                onChange={handlePatternConfigChange}
              />
            )}
          </div>
        )}

        {activeTab === "models" && (
          <div key="models" className="tab-content-enter space-y-8">
            <ModelConfigForm models={spec.models ?? {}} onChange={handleModelsChange} />
            <hr className="border-zinc-200" />
            <StringArrayBindingPanel
              title={t("studio.bindings.tools")}
              description={t("studio.bindings.toolsDesc")}
              addLabel={t("studio.bindings.addTool")}
              items={spec.toolRefs ?? []}
              onChange={handleToolRefsChange}
              placeholder={t("studio.bindings.toolsPlaceholder")}
            />
            <hr className="border-zinc-200" />
            <BindingPanel
              title={t("studio.bindings.knowledge")}
              description={t("studio.bindings.knowledgeDesc")}
              addLabel={t("studio.bindings.addKnowledge")}
              items={spec.knowledgeRefs ?? []}
              onAdd={handleKnowledgeAdd}
              onRemove={handleKnowledgeRemove}
              renderItem={(item, index) => {
                const kb = item as KnowledgeBinding;
                return (
                  <div className="flex gap-2">
                    <Input
                      value={kb.name}
                      placeholder={t("studio.bindings.namePlaceholder")}
                      onChange={(e) => handleKnowledgeChange(index, "name", e.target.value)}
                    />
                    <Input
                      value={kb.ref}
                      placeholder={t("studio.bindings.refPlaceholder")}
                      onChange={(e) => handleKnowledgeChange(index, "ref", e.target.value)}
                    />
                  </div>
                );
              }}
            />
            <hr className="border-zinc-200" />
            <BindingPanel
              title={t("studio.bindings.skills")}
              description={t("studio.bindings.skillsDesc")}
              addLabel={t("studio.bindings.addSkill")}
              items={spec.skillRefs ?? []}
              onAdd={handleSkillAdd}
              onRemove={handleSkillRemove}
              renderItem={(item, index) => {
                const skill = item as SkillBinding;
                return (
                  <div className="flex gap-2">
                    <Input
                      value={skill.name}
                      placeholder={t("studio.bindings.namePlaceholder")}
                      onChange={(e) => handleSkillChange(index, "name", e.target.value)}
                    />
                    <Input
                      value={skill.ref}
                      placeholder={t("studio.bindings.refPlaceholder")}
                      onChange={(e) => handleSkillChange(index, "ref", e.target.value)}
                    />
                  </div>
                );
              }}
            />
            <hr className="border-zinc-200" />
            <BindingPanel
              title={t("studio.bindings.subAgents")}
              description={t("studio.bindings.subAgentsDesc")}
              addLabel={t("studio.bindings.addSubAgent")}
              items={spec.subAgentRefs ?? []}
              onAdd={handleSubAgentAdd}
              onRemove={handleSubAgentRemove}
              renderItem={(item, index) => {
                const sa = item as SubAgentBinding;
                return (
                  <div className="flex gap-2">
                    <Input
                      value={sa.name}
                      placeholder={t("studio.bindings.namePlaceholder")}
                      onChange={(e) => handleSubAgentChange(index, "name", e.target.value)}
                    />
                    <Input
                      value={sa.ref}
                      placeholder={t("studio.bindings.refPlaceholder")}
                      onChange={(e) => handleSubAgentChange(index, "ref", e.target.value)}
                    />
                  </div>
                );
              }}
            />
            <hr className="border-zinc-200" />
            <StringArrayBindingPanel
              title={t("studio.bindings.mcpServers")}
              description={t("studio.bindings.mcpServersDesc")}
              addLabel={t("studio.bindings.addMcpServer")}
              items={spec.mcpRefs ?? []}
              onChange={handleMcpRefsChange}
              placeholder={t("studio.bindings.mcpServersPlaceholder")}
            />
          </div>
        )}

        {activeTab === "preview" && <div key="preview" className="tab-content-enter"><GraphPreview spec={spec} agentId={agentId} /></div>}
      </div>

      <ToastContainer toasts={toast.toasts} onDismiss={toast.dismissToast} />
    </div>
  );
}
