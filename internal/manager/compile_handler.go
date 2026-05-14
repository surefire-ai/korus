package manager

import (
	"encoding/json"
	"net/http"
	"strings"

	apiv1alpha1 "github.com/surefire-ai/korus/api/v1alpha1"
	"github.com/surefire-ai/korus/internal/compiler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CompileResult is the JSON response for the compile endpoint.
type CompileResult struct {
	OK       bool   `json:"ok"`
	Errors   []string `json:"errors,omitempty"`
	Revision string   `json:"revision,omitempty"`
	Artifact any      `json:"artifact,omitempty"`
}

func (s Server) handleCompileAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.Stores.Agents == nil {
		writeError(w, http.StatusServiceUnavailable, "agent store is not configured")
		return
	}

	agent, err := s.Stores.Agents.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	if agent.Spec == nil {
		writeJSON(w, http.StatusOK, CompileResult{OK: true})
		return
	}

	k8sAgent := agentRecordToK8s(*agent)
	refs := buildReferenceIndex(agent.Spec)

	result, compileErr := compiler.CompileAgent(k8sAgent, refs)
	if compileErr != nil {
		writeJSON(w, http.StatusOK, CompileResult{
			OK:     false,
			Errors: splitErrors(compileErr.Error()),
		})
		return
	}

	var artifact any
	if raw, err := json.Marshal(result.Artifact); err == nil {
		_ = json.Unmarshal(raw, &artifact)
	}

	writeJSON(w, http.StatusOK, CompileResult{
		OK:       true,
		Revision: result.Revision,
		Artifact: artifact,
	})
}

// agentRecordToK8s converts a manager AgentRecord to a K8s Agent CRD.
func agentRecordToK8s(record AgentRecord) apiv1alpha1.Agent {
	spec := agentSpecToK8s(record.Spec)
	return apiv1alpha1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      record.ID,
			Namespace: record.WorkspaceID,
		},
		Spec: spec,
	}
}

// agentSpecToK8s converts a manager AgentSpecData to a K8s AgentSpec.
func agentSpecToK8s(spec *AgentSpecData) apiv1alpha1.AgentSpec {
	if spec == nil {
		return apiv1alpha1.AgentSpec{}
	}

	k8s := apiv1alpha1.AgentSpec{
		Models:    convertModels(spec.Models),
		ToolRefs:  spec.ToolRefs,
		MCPRefs:   spec.MCPRefs,
		PolicyRef: spec.PolicyRef,
	}

	k8s.Runtime = apiv1alpha1.AgentRuntimeSpec{
		Engine:      spec.Runtime.Engine,
		RunnerClass: spec.Runtime.RunnerClass,
		Mode:        spec.Runtime.Mode,
		Entrypoint:  spec.Runtime.Entrypoint,
	}

	k8s.Identity = apiv1alpha1.AgentIdentitySpec{
		DisplayName: spec.Identity.DisplayName,
		Role:        spec.Identity.Role,
		Description: spec.Identity.Description,
	}

	if spec.Pattern != nil {
		k8s.Pattern = &apiv1alpha1.AgentPatternSpec{
			Type:             spec.Pattern.Type,
			Version:          spec.Pattern.Version,
			ModelRef:         spec.Pattern.ModelRef,
			ExecutorModelRef: spec.Pattern.ExecutorModelRef,
			ToolRefs:         spec.Pattern.ToolRefs,
			KnowledgeRefs:    spec.Pattern.KnowledgeRefs,
			MaxIterations:    spec.Pattern.MaxIterations,
			StopWhen:         spec.Pattern.StopWhen,
		}
		for _, route := range spec.Pattern.Routes {
			k8s.Pattern.Routes = append(k8s.Pattern.Routes, apiv1alpha1.PatternRoute{
				Label:    route.Label,
				AgentRef: route.AgentRef,
				ModelRef: route.ModelRef,
				Default:  route.Default,
			})
		}
	}

	k8s.PromptRefs = apiv1alpha1.AgentPromptRefs{
		System: spec.PromptRefs.System,
	}

	for _, kb := range spec.KnowledgeRefs {
		k8s.KnowledgeRefs = append(k8s.KnowledgeRefs, apiv1alpha1.KnowledgeBindingSpec{
			Name: kb.Name,
			Ref:  kb.Ref,
		})
	}

	for _, sk := range spec.SkillRefs {
		k8s.SkillRefs = append(k8s.SkillRefs, apiv1alpha1.SkillBindingSpec{
			Name: sk.Name,
			Ref:  sk.Ref,
		})
	}

	for _, sa := range spec.SubAgentRefs {
		k8s.SubAgentRefs = append(k8s.SubAgentRefs, apiv1alpha1.SubAgentBindingSpec{
			Name:      sa.Name,
			Ref:       sa.Ref,
			Namespace: sa.Namespace,
		})
	}

	k8s.Interfaces = apiv1alpha1.AgentInterfaceSpec{
		Input:  apiv1alpha1.SchemaEnvelope{Schema: apiv1alpha1.JSONSchema{}},
		Output: apiv1alpha1.SchemaEnvelope{Schema: apiv1alpha1.JSONSchema{}},
	}

	if spec.Graph != nil {
		for _, n := range spec.Graph.Nodes {
			k8s.Graph.Nodes = append(k8s.Graph.Nodes, apiv1alpha1.AgentGraphNode{
				Name:           n.Name,
				Kind:           n.Kind,
				ModelRef:       n.ModelRef,
				ToolRef:        n.ToolRef,
				KnowledgeRef:   n.KnowledgeRef,
				AgentRef:       n.AgentRef,
				Implementation: n.Implementation,
			})
		}
		for _, e := range spec.Graph.Edges {
			k8s.Graph.Edges = append(k8s.Graph.Edges, apiv1alpha1.AgentGraphEdge{
				From: e.From,
				To:   e.To,
				When: e.When,
			})
		}
	}

	return k8s
}

func convertModels(models map[string]ModelConfig) map[string]apiv1alpha1.ModelSpec {
	if models == nil {
		return nil
	}
	out := make(map[string]apiv1alpha1.ModelSpec, len(models))
	for name, m := range models {
		ms := apiv1alpha1.ModelSpec{
			Provider:       m.Provider,
			Model:          m.Model,
			BaseURL:        m.BaseURL,
			Temperature:    m.Temperature,
			MaxTokens:      m.MaxTokens,
			TimeoutSeconds: m.TimeoutSeconds,
		}
		if m.CredentialRef != nil {
			ms.CredentialRef = &apiv1alpha1.SecretKeyReference{
				Name: m.CredentialRef.Name,
				Key:  m.CredentialRef.Key,
			}
		}
		out[name] = ms
	}
	return out
}

// buildReferenceIndex constructs a minimal ReferenceIndex from the agent spec.
// It populates name sets so the compiler can resolve references that are
// declared in the spec itself. External resources (actual K8s CRDs) are not
// available in the manager, so only self-declared names are indexed.
func buildReferenceIndex(spec *AgentSpecData) compiler.ReferenceIndex {
	refs := compiler.ReferenceIndex{
		Prompts:         map[string]struct{}{},
		PromptTemplates: map[string]apiv1alpha1.PromptTemplateSpec{},
		KnowledgeBases:  map[string]struct{}{},
		KnowledgeSpecs:  map[string]apiv1alpha1.KnowledgeBaseSpec{},
		Tools:           map[string]struct{}{},
		ToolSpecs:       map[string]apiv1alpha1.ToolProviderSpec{},
		Skills:          map[string]struct{}{},
		SkillSpecs:      map[string]apiv1alpha1.SkillSpec{},
		SubAgents:       map[string]struct{}{},
		SubAgentRefs:    map[string][]apiv1alpha1.SubAgentBindingSpec{},
		MCPServers:      map[string]struct{}{},
		Policies:        map[string]struct{}{},
	}

	for _, name := range spec.ToolRefs {
		refs.Tools[name] = struct{}{}
	}

	for _, kb := range spec.KnowledgeRefs {
		if kb.Name != "" {
			refs.KnowledgeBases[kb.Name] = struct{}{}
		}
		if kb.Ref != "" {
			refs.KnowledgeBases[kb.Ref] = struct{}{}
		}
	}

	for _, sk := range spec.SkillRefs {
		if sk.Name != "" {
			refs.Skills[sk.Name] = struct{}{}
		}
		if sk.Ref != "" {
			refs.Skills[sk.Ref] = struct{}{}
		}
	}

	for _, sa := range spec.SubAgentRefs {
		if sa.Name != "" {
			refs.SubAgents[sa.Name] = struct{}{}
		}
		if sa.Ref != "" {
			refs.SubAgents[sa.Ref] = struct{}{}
		}
		refs.SubAgentRefs[sa.Ref] = nil
	}

	for _, name := range spec.MCPRefs {
		refs.MCPServers[name] = struct{}{}
	}

	if spec.PolicyRef != "" {
		refs.Policies[spec.PolicyRef] = struct{}{}
	}

	return refs
}

// splitErrors splits a compiler error string into individual error messages.
func splitErrors(errStr string) []string {
	parts := strings.Split(errStr, "; ")
	var errors []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			errors = append(errors, p)
		}
	}
	if len(errors) == 0 {
		errors = []string{errStr}
	}
	return errors
}
